/*
Copyright IBM Corp. 2016 All Rights Reserved.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

		 http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package pbft

import (
	"fmt"
	"time"

	"github.com/cherryFloris/pbft-vrf/consensus"
	"github.com/cherryFloris/pbft-vrf/consensus/util/events"
	pb "github.com/cherryFloris/pbft-vrf/protos"

	"github.com/golang/protobuf/proto"
	"github.com/golang/protobuf/ptypes/timestamp"
	"github.com/op/go-logging"
	"github.com/spf13/viper"
)

type obcBatch struct {
	obcGeneric
	externalEventReceiver
	pbft        *pbftCore
	broadcaster *broadcaster

	batchSize        int
	batchStore       []*Request
	batchTimer       events.Timer
	batchTimerActive bool
	batchTimeout     time.Duration

	manager events.Manager // TODO, remove eventually, the event manager

	incomingChan chan *batchMessage // Queues messages for processing by main thread
	idleChan     chan struct{}      // Idle channel, to be removed

	reqStore *requestStore // Holds the outstanding and pending requests

	deduplicator *deduplicator

	persistForward

	selectTimer       events.Timer
	selectTimerActive bool
	selectTimeout     time.Duration
}

type batchMessage struct {
	msg    *pb.Message
	sender *pb.PeerID
}

// Event types

// batchMessageEvent is sent when a consensus message is received that is then to be sent to pbft
type batchMessageEvent batchMessage

// batchTimerEvent is sent when the batch timer expires
type batchTimerEvent struct{}

type VRFEvent *VrfProve
type selectTimerEvent struct{}

func newObcBatch(id uint64, config *viper.Viper, stack consensus.Stack) *obcBatch {
	var err error

	op := &obcBatch{
		obcGeneric: obcGeneric{stack: stack},
	}

	op.persistForward.persistor = stack

	logger.Debugf("Replica %d obtaining startup information", id)

	op.manager = events.NewManagerImpl() // TODO, this is hacky, eventually rip it out
	op.manager.SetReceiver(op)
	etf := events.NewTimerFactoryImpl(op.manager)
	op.pbft = newPbftCore(id, config, op, etf)
	op.manager.Start()
	blockchainInfoBlob := stack.GetBlockchainInfoBlob()
	op.externalEventReceiver.manager = op.manager
	op.broadcaster = newBroadcaster(id, op.pbft.N, op.pbft.f, op.pbft.broadcastTimeout, stack)
	op.manager.Queue() <- workEvent(func() {
		op.pbft.stateTransfer(&stateUpdateTarget{
			checkpointMessage: checkpointMessage{
				seqNo: op.pbft.lastExec,
				id:    blockchainInfoBlob,
			},
		})
	})

	op.batchSize = config.GetInt("general.batchsize")
	op.batchStore = nil
	op.batchTimeout, err = time.ParseDuration(config.GetString("general.timeout.batch"))
	op.batchTimeout = op.batchTimeout / 2
	if err != nil {
		panic(fmt.Errorf("Cannot parse batch timeout: %s", err))
	}
	logger.Infof("PBFT Batch size = %d", op.batchSize)
	logger.Infof("PBFT Batch timeout = %v", op.batchTimeout)

	if op.batchTimeout >= op.pbft.requestTimeout {
		op.pbft.requestTimeout = 3 * op.batchTimeout / 2
		logger.Warningf("Configured request timeout must be greater than batch timeout, setting to %v", op.pbft.requestTimeout)
	}

	if op.pbft.requestTimeout >= op.pbft.nullRequestTimeout && op.pbft.nullRequestTimeout != 0 {
		op.pbft.nullRequestTimeout = 3 * op.pbft.requestTimeout / 2
		logger.Warningf("Configured null request timeout must be greater than request timeout, setting to %v", op.pbft.nullRequestTimeout)
	}

	op.incomingChan = make(chan *batchMessage)

	op.batchTimer = etf.CreateTimer()

	op.reqStore = newRequestStore()

	op.deduplicator = newDeduplicator()

	op.idleChan = make(chan struct{})
	close(op.idleChan) // TODO remove eventually

	op.selectTimer = etf.CreateTimer()
	op.selectTimeout, err = time.ParseDuration(config.GetString("general.timeout.select"))
	op.selectTimeout = op.selectTimeout / 2
	op.stopSelectTimer()

	return op
}

// Close tells us to release resources we are holding
func (op *obcBatch) Close() {
	op.batchTimer.Halt()
	op.pbft.close()
}

func (op *obcBatch) submitToLeader(req *Request) events.Event {
	// Broadcast the request to the network, in case we're in the wrong view

	op.broadcastMsg(&BatchMessage{Payload: &BatchMessage_Request{Request: req}})
	op.logAddTxFromRequest(req)
	op.reqStore.storeOutstanding(req)
	op.pbft.clientRequests = append(op.pbft.clientRequests, req)
	//logger.Infof("clientReq's length:%d", len(op.pbft.clientRequests))
	//op.startTimerIfOutstandingRequests()
	/*if op.pbft.primary(op.pbft.view) == op.pbft.id && op.pbft.activeView {
		return op.leaderProcReq(req)
	}*/
	return nil
}

func (op *obcBatch) broadcastMsg(msg *BatchMessage) {
	//logger.Infof("broadcastMsg")
	msgPayload, _ := proto.Marshal(msg)
	ocMsg := &pb.Message{
		Type:    pb.Message_CONSENSUS,
		Payload: msgPayload,
	}
	op.broadcaster.Broadcast(ocMsg)
}

// send a message to a specific replica
func (op *obcBatch) unicastMsg(msg *BatchMessage, receiverID uint64) {
	msgPayload, _ := proto.Marshal(msg)
	ocMsg := &pb.Message{
		Type:    pb.Message_CONSENSUS,
		Payload: msgPayload,
	}
	op.broadcaster.Unicast(ocMsg, receiverID)
}

// =============================================================================
// innerStack interface (functions called by pbft-core)
// =============================================================================

// multicast a message to all replicas
func (op *obcBatch) broadcast(msgPayload []byte) {
	op.broadcaster.Broadcast(op.wrapMessage(msgPayload))
}

// multicast a message to clerk
func (op *obcBatch) broadcastToClerk(msgPayload []byte) {
	for id, _ := range op.pbft.clerk {
		if id == op.pbft.id { //自己不需要用通道给自己发
			continue
		}
		op.broadcaster.send(op.wrapMessage(msgPayload), &id)
	}
	//op.broadcaster.Broadcast(op.wrapMessage(msgPayload))
}

// multicast a message to clerk
func (op *obcBatch) broadcastToAll(msg *pb.Message) {
	op.broadcaster.Broadcast(msg)
}

// send a message to a specific replica
func (op *obcBatch) unicastToOne(msg *pb.Message, receiverID uint64) {
	op.broadcaster.Unicast(msg, receiverID)
}

// send a message to a specific replica
func (op *obcBatch) unicast(msgPayload []byte, receiverID uint64) (err error) {
	return op.broadcaster.Unicast(op.wrapMessage(msgPayload), receiverID)
}

func (op *obcBatch) sign(msg []byte) ([]byte, error) {
	return op.stack.Sign(msg)
}

// verify message signature
func (op *obcBatch) verify(senderID uint64, signature []byte, message []byte) error {
	senderHandle, err := getValidatorHandle(senderID)
	if err != nil {
		return err
	}
	return op.stack.Verify(senderHandle, signature, message)
}

// execute an opaque request which corresponds to an OBC Transaction
func (op *obcBatch) execute(seqNo uint64, reqBatch *RequestBatch) {
	var txs []*pb.Transaction
	for _, req := range reqBatch.GetBatch() {
		tx := &pb.Transaction{}
		if err := proto.Unmarshal(req.Payload, tx); err != nil {
			logger.Warningf("Batch replica %d could not unmarshal transaction %s", op.pbft.id, err)
			continue
		}
		//logger.Infof("Batch replica %d executing request with transaction %s from outstandingReqs, seqNo=%d", op.pbft.id, tx.Txid, seqNo)
		if outstanding, pending := op.reqStore.remove(req); !outstanding || !pending {
			//logger.Infof("Batch replica %d missing transaction %s outstanding=%v, pending=%v", op.pbft.id, tx.Txid, outstanding, pending)
		}
		txs = append(txs, tx)
		op.deduplicator.Execute(req)
	}
	meta, _ := proto.Marshal(&Metadata{seqNo})
	logger.Infof("Batch replica %d received exec for seqNo %d containing %d transactions", op.pbft.id, seqNo, len(txs))
	op.stack.Execute(meta, txs) // This executes in the background, we will receive an executedEvent once it completes
}

// =============================================================================
// functions specific to batch mode
// =============================================================================

//func (op *obcBatch) leaderProcReq(req *Request) events.Event {
func (op *obcBatch) leaderProcReq(choice int) events.Event {
	if op.reqStore.outstandingRequests.Len()-op.pbft.notConsensused >= op.batchSize { //如果交易够多了，就将固定数量的交易都打包
		for i := op.pbft.notConsensused; i < op.batchSize+op.pbft.notConsensused; i++ {
			op.batchStore = append(op.batchStore, op.pbft.clientRequests[i])
			op.reqStore.storePending(op.pbft.clientRequests[i])
		}
		//op.pbft.nextNotConsensused = op.pbft.notConsensused + op.batchSize
	} else { //如果时间到了，还没有足够多的交易，就将剩下的交易都打包
		for i := op.pbft.notConsensused; i < len(op.pbft.clientRequests); i++ {
			op.batchStore = append(op.batchStore, op.pbft.clientRequests[i])
			op.reqStore.storePending(op.pbft.clientRequests[i])
		}
		//op.pbft.nextNotConsensused = len(op.pbft.clientRequests)
	}
	return op.sendBatch(choice)
}

func (op *obcBatch) sendBatch(choice int) events.Event {
	op.stopBatchTimer()
	if len(op.batchStore) == 0 {
		logger.Error("Told to send an empty batch store for ordering, ignoring")
		return nil
	}
	b := op.batchStore
	/*if choice == 2{
		b = op.nbatchStore
	}*/
	//reqBatch := &RequestBatch{Batch: op.batchStore}
	reqBatch := &RequestBatch{Batch: b}
	op.batchStore = nil
	logger.Infof("leader %d Creating batch with %d requests", op.pbft.id, len(reqBatch.Batch))
	return reqBatch
}

func (op *obcBatch) txToReq(tx []byte) *Request {
	now := time.Now()
	req := &Request{
		Timestamp: &timestamp.Timestamp{
			Seconds: now.Unix(),
			Nanos:   int32(now.UnixNano() % 1000000000),
		},
		Payload:   tx,
		ReplicaId: op.pbft.id,
	}
	// XXX sign req
	return req
}

func (op *obcBatch) processMessage(ocMsg *pb.Message, senderHandle *pb.PeerID) events.Event {
	if ocMsg.Type == pb.Message_CHAIN_TRANSACTION {
		//logger.Infof("chain-transaction")
		req := op.txToReq(ocMsg.Payload)
		return op.submitToLeader(req)
	}

	if ocMsg.Type == pb.Message_VRFPROVE {
		r := op.verifyVRF(ocMsg.Payload)
		return r
		//if op.pbft.ifSendVrf == false{//自己还没有发送vrf
		//op.startSelectTimer()
		//op.selectLeaderAndDelegate(1)
		//	op.sendVRF()
		//}
		//return op.verifyVRF(ocMsg.Payload)
	}

	if ocMsg.Type == pb.Message_COK {
		return nil
		pview := uint64(ocMsg.Payload[0])
		if pview >= op.pbft.view {
			op.pbft.okNum++
			logger.Infof("receive ok,have %v, want %v", op.pbft.okNum, op.pbft.N)

			if op.pbft.okNum == op.pbft.N { //所有节点达成共识
				if op.pbft.notConsensused < len(op.pbft.clientRequests) {
					op.pbft.okNum = 0
					logger.Infof("have reqs not-consensused, selecting new leader and clerks.")
					op.sendVRF()
				}
			}
		} else {
			logger.Infof("ok from last view, ignore.")
		}

		return nil
	}

	if ocMsg.Type == pb.Message_LEADER { //知道"主节点知道了自己是主节点"
		vv := int(ocMsg.Payload[0])
		if uint64(vv) == op.pbft.view {
			logger.Infof("now leader know he is leader.")
			op.pbft.softStartTimer(op.pbft.requestTimeout, fmt.Sprintf("waiting for pre-prepare message"))
		}

		return nil
	}

	if ocMsg.Type == pb.Message_LEADERCHEAT { //主节点发送不同的交易列表
		op.pbft.leaderCheatingNum++
		if op.pbft.leaderCheatingNum == len(op.pbft.clerk) { //所有计票节点都认为主节点恶意行为
			//扣分
			_, ok := op.pbft.scores[op.pbft.primary(op.pbft.view)]
			if !ok { //op.pbft.scores[]还未初始化该项
				op.pbft.scores[op.pbft.primary(op.pbft.view)] = 0
			}
			var a int = 2
			op.pbft.scores[op.pbft.primary(op.pbft.view)] -= a

			//重新选主
			op.pbft.ifSendVrf = false
			op.pbft.leader = uint64(0)
			for index, _ := range op.pbft.clerk {
				delete(op.pbft.clerk, index)
			}
			op.pbft.view += 1
			op.pbft.FeedbackNum = 0
			op.pbft.vrfs = []uint64{}
			for index, _ := range op.pbft.vrf_peers {
				delete(op.pbft.vrf_peers, index)
			}
			op.pbft.leaderCheatingNum = 0
			op.pbft.pdigest = ""
			op.pbft.pview = 0
			op.pbft.pseq = 0
			op.pbft.prePrepare = nil
			op.pbft.ifAsking = false
			op.pbft.prepares = []*Prepare3{}
			op.pbft.ifSendFeedback = false
			op.pbft.clerkPrepares = []uint64{}
			op.pbft.ifKnowResult = false

			if op.pbft.id == uint64(0) {
				//op.startSelectTimer()
				logger.Infof("leader cheat event, selecting new leader and clerks.")
				//op.selectLeaderAndDelegate(2)
				op.sendVRF()
			}
		}
		return nil
	}

	if ocMsg.Type != pb.Message_CONSENSUS {
		logger.Errorf("Unexpected message type: %s", ocMsg.Type)
		return nil
	}

	batchMsg := &BatchMessage{}
	err := proto.Unmarshal(ocMsg.Payload, batchMsg)
	if err != nil {
		logger.Errorf("Error unmarshaling message: %s", err)
		return nil
	}

	if req := batchMsg.GetRequest(); req != nil {
		if !op.deduplicator.IsNew(req) {
			logger.Warningf("Replica %d ignoring request as it is too old", op.pbft.id)
			return nil
		}

		op.logAddTxFromRequest(req)
		op.reqStore.storeOutstanding(req)
		op.pbft.clientRequests = append(op.pbft.clientRequests, req)

		/*if (op.pbft.primary(op.pbft.view) == op.pbft.id) && op.pbft.activeView {
			return op.leaderProcReq(req)
		}*/
		//op.startTimerIfOutstandingRequests()
		return nil
	} else if pbftMsg := batchMsg.GetPbftMessage(); pbftMsg != nil {
		senderID, err := getValidatorID(senderHandle) // who sent this?
		if err != nil {
			panic("Cannot map sender's PeerID to a valid replica ID")
		}
		msg := &Message{}
		err = proto.Unmarshal(pbftMsg, msg)
		if err != nil {
			logger.Errorf("Error unpacking payload from message: %s", err)
			return nil
		}
		return pbftMessageEvent{
			msg:    msg,
			sender: senderID,
		}
	}

	logger.Errorf("Unknown request: %+v", batchMsg)

	return nil
}

func (op *obcBatch) logAddTxFromRequest(req *Request) {
	if logger.IsEnabledFor(logging.DEBUG) {
		// This is potentially a very large expensive debug statement, guard
		tx := &pb.Transaction{}
		err := proto.Unmarshal(req.Payload, tx)
		if err != nil {
			logger.Errorf("Replica %d was sent a transaction which did not unmarshal: %s", op.pbft.id, err)
		} else {
			logger.Debugf("Replica %d adding request from %d with transaction %s into outstandingReqs", op.pbft.id, req.ReplicaId, tx.Txid)
		}
	}
}

func (op *obcBatch) resubmitOutstandingReqs() events.Event {
	//op.startTimerIfOutstandingRequests()

	// If we are the primary, and know of outstanding requests, submit them for inclusion in the next batch until
	// we run out of requests, or a new batch message is triggered (this path will re-enter after execution)
	// Do not enter while an execution is in progress to prevent duplicating a request
	/*if op.pbft.primary(op.pbft.view) == op.pbft.id && op.pbft.activeView && op.pbft.currentExec == nil {
		needed := op.batchSize - len(op.batchStore)

		for op.reqStore.hasNonPending() {
			outstanding := op.reqStore.getNextNonPending(needed)

			// If we have enough outstanding requests, this will trigger a batch
			for _, nreq := range outstanding {
				if msg := op.leaderProcReq(nreq); msg != nil {
					op.manager.Inject(msg)
				}
			}
		}
	}*/
	return nil
}

// allow the primary to send a batch when the timer expires
func (op *obcBatch) ProcessEvent(event events.Event) events.Event {
	logger.Debugf("Replica %d batch main thread looping", op.pbft.id)
	switch et := event.(type) {
	case batchMessageEvent:
		if et.msg.Type == pb.Message_CHAIN_TRANSACTION {
			if len(op.pbft.vrf_peers) == 0 { //还没有选主成功
				if op.pbft.ifKnowResult == false {
					logger.Infof("batch message event, selecting new leader and clerks.")
					op.sendVRF()
				}
			}
		}
		ocMsg := et
		return op.processMessage(ocMsg.msg, ocMsg.sender)
	case executedEvent:
		op.stack.Commit(nil, et.tag.([]byte))
	case committedEvent:
		logger.Debugf("Replica %d received committedEvent", op.pbft.id)
		return execDoneEvent{}
	case execDoneEvent:

		op.pbft.ifKnowResult = false
		//op.pbft.okNum++

		res := op.pbft.ProcessEvent(event)
		if res != nil {
			op.pbft.ProcessEvent(res)
		}

		logger.Warningf("next consensused req:%v, amount req:%v", op.pbft.notConsensused, len(op.pbft.clientRequests))
		logger.Warningf("waitedCerts:%v, waited:%v,", len(op.pbft.waitedCerts), op.pbft.waited)

		if op.pbft.okNum == op.pbft.N { //所有节点达成共识
			if op.pbft.notConsensused < len(op.pbft.clientRequests) {
				op.pbft.okNum = 0
				logger.Infof("have reqs not-consensused, selecting new leader and clerks.")
				op.sendVRF()
			}
		}

		//如果有还没有提交的区块
		if op.pbft.waited != len(op.pbft.waitedCerts) {
			cert := op.pbft.waitedCerts[op.pbft.waited]
			op.pbft.waited++
			op.pbft.executeOutstanding2(cert.prePrepare.View, cert.prePrepare.SequenceNumber)
		}

		return op.resubmitOutstandingReqs()
	case batchTimerEvent:
		logger.Infof("Replica %d batch timer expired", op.pbft.id)
		reqBatch := op.leaderProcReq(1)
		op.ProcessEvent(reqBatch)
		//if op.pbft.activeView && (len(op.batchStore) > 0) {
		//return op.sendBatch(1)
		//}
	case *Commit:
		// TODO, this is extremely hacky, but should go away when batch and core are merged
		res := op.pbft.ProcessEvent(event)
		op.startTimerIfOutstandingRequests()
		return res
	case viewChangedEvent:
		op.batchStore = nil
		// Outstanding reqs doesn't make sense for batch, as all the requests in a batch may be processed
		// in a different batch, but PBFT core can't see through the opaque structure to see this
		// so, on view change, clear it out
		op.pbft.outstandingReqBatches = make(map[string]*RequestBatch)

		logger.Debugf("Replica %d batch thread recognizing new view", op.pbft.id)
		if op.batchTimerActive {
			op.stopBatchTimer()
		}

		if op.pbft.skipInProgress {
			// If we're the new primary, but we're in state transfer, we can't trust ourself not to duplicate things
			op.reqStore.outstandingRequests.empty()
		}

		op.reqStore.pendingRequests.empty()
		for i := op.pbft.h + 1; i <= op.pbft.h+op.pbft.L; i++ {
			if i <= op.pbft.lastExec {
				continue
			}

			cert, ok := op.pbft.certStore[msgID{v: op.pbft.view, n: i}]
			if !ok || cert.prePrepare == nil {
				continue
			}

			if cert.prePrepare.BatchDigest == "" {
				// a null request
				continue
			}

			if cert.prePrepare.RequestBatch == nil {
				logger.Warningf("Replica %d found a non-null prePrepare with no request batch, ignoring")
				continue
			}

			op.reqStore.storePendings(cert.prePrepare.RequestBatch.GetBatch())
		}

		return op.resubmitOutstandingReqs()
	case stateUpdatedEvent:
		// When the state is updated, clear any outstanding requests, they may have been processed while we were gone
		op.reqStore = newRequestStore()
		return op.pbft.ProcessEvent(event)
	case VRFEvent:
		return op.verifyVRF(et)
	case selectTimerEvent:
		op.sendVRF()
	case viewChangeQuorumEvent:
		logger.Infof("Replica %d received view change quorum, processing new view", op.pbft.id)
		op.pbft.stopTimer()
		op.pbft.activeView = false

		//扣分
		_, ok := op.pbft.scores[op.pbft.primary(op.pbft.view)]
		if !ok { //op.pbft.scores[]还未初始化该项
			op.pbft.scores[op.pbft.primary(op.pbft.view)] = 0
		}
		var a int = 2
		op.pbft.scores[op.pbft.primary(op.pbft.view)] -= a

		//重新选主
		op.pbft.ifSendVrf = false
		op.pbft.leader = uint64(0)
		for index, _ := range op.pbft.clerk {
			delete(op.pbft.clerk, index)
		}
		op.pbft.view += 1
		op.pbft.FeedbackNum = 0
		op.pbft.vrfs = []uint64{}
		for index, _ := range op.pbft.vrf_peers {
			delete(op.pbft.vrf_peers, index)
		}
		op.pbft.leaderCheatingNum = 0
		op.pbft.pdigest = ""
		op.pbft.pview = 0
		op.pbft.pseq = 0
		op.pbft.prePrepare = nil
		op.pbft.ifAsking = false
		op.pbft.prepares = []*Prepare3{}
		op.pbft.ifSendFeedback = false
		op.pbft.clerkPrepares = []uint64{}
		op.pbft.ifKnowResult = false

		if op.pbft.id == uint64(0) { //让节点0发起选主
			//op.startSelectTimer()
			logger.Infof("viewchange quorum event, selecting new leader and clerks.")
			op.startSelectTimer()
			//op.sendVRF()
			//op.selectLeaderAndDelegate(2)
		}
	default:
		return op.pbft.ProcessEvent(event)
	}

	return nil
}

func (op *obcBatch) startBatchTimer() {
	op.batchTimer.Reset(op.batchTimeout, batchTimerEvent{})
	logger.Debugf("Replica %d started the batch timer", op.pbft.id)
	op.batchTimerActive = true
}

func (op *obcBatch) stopBatchTimer() {
	op.batchTimer.Stop()
	logger.Debugf("Replica %d stopped the batch timer", op.pbft.id)
	op.batchTimerActive = false
}

func (op *obcBatch) startSelectTimer() {
	op.selectTimer.Reset(op.selectTimeout, selectTimerEvent{})
	logger.Debugf("Replica %d started the select timer", op.pbft.id)
	op.selectTimerActive = true
}

func (op *obcBatch) stopSelectTimer() {
	op.selectTimer.Stop()
	logger.Debugf("Replica %d stopped the select timer", op.pbft.id)
	op.selectTimerActive = false
}

// Wraps a payload into a batch message, packs it and wraps it into
// a Fabric message. Called by broadcast before transmission.
func (op *obcBatch) wrapMessage(msgPayload []byte) *pb.Message {
	batchMsg := &BatchMessage{Payload: &BatchMessage_PbftMessage{PbftMessage: msgPayload}}
	packedBatchMsg, _ := proto.Marshal(batchMsg)
	ocMsg := &pb.Message{
		Type:    pb.Message_CONSENSUS,
		Payload: packedBatchMsg,
	}
	return ocMsg
}

// Retrieve the idle channel, only used for testing
func (op *obcBatch) idleChannel() <-chan struct{} {
	return op.idleChan
}

// TODO, temporary
func (op *obcBatch) getManager() events.Manager {
	return op.manager
}

func (op *obcBatch) startTimerIfOutstandingRequests() {
	if op.pbft.skipInProgress || op.pbft.currentExec != nil || !op.pbft.activeView {
		// Do not start view change timer if some background event is in progress
		logger.Debugf("Replica %d not starting timer because skip in progress or current exec or in view change", op.pbft.id)
		return
	}

	if !op.reqStore.hasNonPending() {
		// Only start a timer if we are aware of outstanding requests
		logger.Debugf("Replica %d not starting timer because all outstanding requests are pending", op.pbft.id)
		return
	}
	op.pbft.softStartTimer(op.pbft.requestTimeout, "Batch outstanding requests")
}
