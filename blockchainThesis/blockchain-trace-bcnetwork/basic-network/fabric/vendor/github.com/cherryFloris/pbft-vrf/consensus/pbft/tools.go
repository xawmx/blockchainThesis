package pbft

import (
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"crypto/x509"
	"encoding/binary"
	"encoding/pem"
	"fmt"
	pb "github.com/cherryFloris/pbft-vrf/protos"
	"github.com/google/keytransparency/core/crypto/vrf"
	"github.com/google/keytransparency/core/crypto/vrf/p256"
	"strconv"

	"github.com/cherryFloris/pbft-vrf/consensus/util/events"
	"github.com/golang/protobuf/proto"

	math "math/rand"
	"time"
)

var curve = elliptic.P256()

/*
***得到vrf随机数和证明
 */
func getVRF(m []byte) (number [32]byte, proof []byte, pubkey ecdsa.PublicKey) {
	//k, pk := p256.GenerateKey()
	k, err := ecdsa.GenerateKey(curve, rand.Reader)
	if err != nil {
		fmt.Println("error")
	}
	pubkey = k.PublicKey
	var sk vrf.PrivateKey = &p256.PrivateKey{PrivateKey: k} //sk是vrf.PrivteKey型
	number, proof = sk.Evaluate(m)
	return
}

/*
***验证vrf随机数
 */
func checkVRF(m []byte, number [32]byte, proof []byte, pubkey interface{}) bool {
	epk := pubkey.(*ecdsa.PublicKey) //ecdsa.PublicKey类型
	pk, _ := p256.NewVRFVerifier(epk)
	got, _ := pk.ProofToHash(m, proof)
	if got == number {
		return true
	}
	return false
}

/*
***公钥转为[]byte
 */
func PublicKeyToPEM(pk *ecdsa.PublicKey) ([]byte, error) {
	PubASN1, err := x509.MarshalPKIXPublicKey(pk)
	if err != nil {
		return nil, err
	}

	return pem.EncodeToMemory(
		&pem.Block{
			Type:  "ECDSA PUBLIC KEY",
			Bytes: PubASN1,
		},
	), nil
}

/*
***[]byte还原为公钥
 */
func PEMToPublicKey(bpk []byte) (interface{}, error) {
	if len(bpk) == 0 {
		return nil, fmt.Errorf("[% x]'s length is 0", bpk)
	}
	block, _ := pem.Decode(bpk)
	if block == nil {
		return nil, fmt.Errorf("Failed decoding [% x]", bpk)
	}
	key, err := x509.ParsePKIXPublicKey(block.Bytes)
	if err != nil {
		return nil, err
	}
	return key, err
}

func GetFirstNumber(str string) int {
	s := string(str[0])
	if s == "a" {
		s = "10"
	}
	if s == "b" {
		s = "11"
	}
	if s == "c" {
		s = "12"
	}
	if s == "d" {
		s = "13"
	}
	if s == "e" {
		s = "14"
	}
	if s == "f" {
		s = "15"
	}
	num2, err := strconv.Atoi(s)
	if err != nil {
		fmt.Println(err)
	}
	return num2
}

func (op *obcBatch) sendVRF() {
	logger.Infof("ifSendVrf:%v", op.pbft.ifSendVrf)

	if op.pbft.ifSendVrf == false {
		op.pbft.ifKnowResult = false
		op.pbft.okNum = 0

		op.pbft.ifSendVrf = true
		//产生vrf随机数
		m := make([]byte, 8)
		binary.BigEndian.PutUint64(m, op.pbft.id)
		number, proof, epk := getVRF(m)
		bpk, _ := PublicKeyToPEM(&epk)

		//封装vrf随机数和证明
		num2 := []byte{}
		for i := 0; i < len(number); i++ {
			num2 = append(num2, number[i])
		}
		//记录自己的vrf
		flag := 0
		for _, value := range op.pbft.vrf_peers {
			if value == op.pbft.id {
				flag = 1
				break
			}
		}
		if flag == 0 {
			//校验节点可信度
			s, ok := op.pbft.scores[op.pbft.id]
			var level int = -6
			if ok && s <= level {
				//节点不可信，将其vrf设为0
				logger.Warningf("replica %d is not trusted", op.pbft.id)
				op.pbft.vrf_peers[uint64(0)] = op.pbft.id
				op.pbft.vrfs = append(op.pbft.vrfs, uint64(0))
			} else {
				op.pbft.vrf_peers[binary.BigEndian.Uint64(num2)] = op.pbft.id
				op.pbft.vrfs = append(op.pbft.vrfs, binary.BigEndian.Uint64(num2))
			}
		}

		data := []byte{}
		data = append(data, m[7])
		data = append(data, byte(op.pbft.view))
		for i := 0; i < len(number); i++ {
			data = append(data, number[i])
		}
		data = append(data, byte(len(proof)))
		for i := 0; i < len(proof); i++ {
			data = append(data, proof[i])
		}
		for i := 0; i < len(bpk); i++ {
			data = append(data, bpk[i])
		}
		msg := &pb.Message{
			Type:    pb.Message_VRFPROVE,
			Payload: data,
		}

		//发送消息
		op.broadcaster.Broadcast(msg)
	}
}

func (instance *pbftCore) sendVRF2() {
	logger.Infof("ifSendVrf2:%v", instance.ifSendVrf)

	if instance.ifSendVrf == false {
		instance.ifKnowResult = false
		instance.okNum = 0

		instance.ifSendVrf = true
		//产生vrf随机数
		m := make([]byte, 8)
		binary.BigEndian.PutUint64(m, instance.id)
		number, proof, epk := getVRF(m)
		bpk, _ := PublicKeyToPEM(&epk)

		//封装vrf随机数和证明
		num2 := []byte{}
		for i := 0; i < len(number); i++ {
			num2 = append(num2, number[i])
		}
		//记录自己的vrf
		flag := 0
		for _, value := range instance.vrf_peers {
			if value == instance.id {
				flag = 1
				break
			}
		}
		if flag == 0 {
			//校验节点可信度
			s, ok := instance.scores[instance.id]
			var level int = -6
			if ok && s <= level {
				//节点不可信，将其vrf设为0
				logger.Warningf("replica %d is not trusted", instance.id)
				instance.vrf_peers[uint64(0)] = instance.id
				instance.vrfs = append(instance.vrfs, uint64(0))
			} else {
				instance.vrf_peers[binary.BigEndian.Uint64(num2)] = instance.id
				instance.vrfs = append(instance.vrfs, binary.BigEndian.Uint64(num2))
			}
		}

		data := []byte{}
		data = append(data, m[7])
		data = append(data, byte(instance.view))
		for i := 0; i < len(number); i++ {
			data = append(data, number[i])
		}
		data = append(data, byte(len(proof)))
		for i := 0; i < len(proof); i++ {
			data = append(data, proof[i])
		}
		for i := 0; i < len(bpk); i++ {
			data = append(data, bpk[i])
		}
		msg := &pb.Message{
			Type:    pb.Message_VRFPROVE,
			Payload: data,
		}

		//发送消息
		instance.consumer.broadcastToAll(msg)
	}
}

func (op *obcBatch) selectLeaderAndDelegate(choice int) {
	//如果selectTimer开启，则发送vrf。否则进行选主
	if op.selectTimerActive == true {
		op.sendVRF()
	} else {
		op.selectLeader(choice)
	}
}

/*
***验证VRF随机数和证明
 */
func (op *obcBatch) verifyVRF(event events.Event) events.Event {
	//还原数据
	data := event.([]byte)
	//logger.Infof("data:%v", data)
	pid := make([]byte, 8)
	binary.BigEndian.PutUint64(pid, uint64(data[0]))

	//v := uint64(data[1])
	//logger.Infof("received view:%v, wanted view:%v", v, op.pbft.view)
	//if v == op.pbft.view{
	number := [32]byte{}
	for i := 0; i < 32; i++ {
		number[i] = data[i+2]
	}
	plength := int(data[34])
	proof := []byte(data[35 : 35+plength])
	bpk := []byte(data[35+plength:])
	pk, _ := PEMToPublicKey(bpk)

	//校验节点可信度
	s, ok := op.pbft.scores[uint64(data[0])]
	var level int = -6
	if ok && s <= level {
		//节点不可信，将其vrf设为0
		logger.Warningf("replica %v is not trusted", uint64(data[0]))
		op.pbft.vrf_peers[uint64(0)] = uint64(data[0])
		op.pbft.vrfs = append(op.pbft.vrfs, uint64(0))
	} else {
		//验证
		//i1 := fmt.Sprintf("%x", number)//随机数number以string形式存储
		//num2 := GetFirstNumber(i1)
		/*if num2 >= 5{
			return nil
		}*/
		r := checkVRF(pid, number, proof, pk)
		if r == false {
			logger.Errorf("peer %v is cheating", pid)
		} else {
			flag := 0
			num2 := []byte{}
			for i := 0; i < len(number); i++ {
				num2 = append(num2, number[i])
			}
			//只接收每个节点第一次发送的vrf
			for _, value := range op.pbft.vrf_peers {
				if value == uint64(data[0]) {
					flag = 1
					break
				}
			}
			if flag == 0 {
				op.pbft.vrf_peers[binary.BigEndian.Uint64(num2)] = uint64(data[0])
				op.pbft.vrfs = append(op.pbft.vrfs, binary.BigEndian.Uint64(num2))
			}

		}
	}
	logger.Infof("receive vrf from replica %v,len of vrfs: %v", uint64(data[0]), len(op.pbft.vrfs))
	if op.pbft.ifSendVrf == false { //自己还没有发送vrf

		op.sendVRF()
	}

	//收到所有节点的vrf random之后，就开始选主
	if len(op.pbft.vrfs) >= op.pbft.N {
		op.selectLeader(1)
	}
	//}

	return nil
}

/*
***从计票节点中选择主节点
 */
func (op *obcBatch) selectLeader(choice int) {
	//将vrf降序排列
	for i := 0; i < len(op.pbft.vrfs)-1; i++ {
		for j := i + 1; j < len(op.pbft.vrfs); j++ {
			if op.pbft.vrfs[i] < op.pbft.vrfs[j] {
				op.pbft.vrfs[i], op.pbft.vrfs[j] = op.pbft.vrfs[j], op.pbft.vrfs[i]
			}
		}
	}

	//for key, value := range(op.pbft.vrf_peers){
	//logger.Infof("peer:%v, vrf:%v", value, key)
	//}

	//vrf最大的是主节点
	op.pbft.leader = op.pbft.vrf_peers[op.pbft.vrfs[0]] + uint64(1)
	//找f+1个clerk
	for i := 1; i < op.pbft.f+2; i++ {
		op.pbft.clerk[op.pbft.vrf_peers[op.pbft.vrfs[i]]] = op.pbft.vrfs[i]
	}

	op.pbft.activeView = true
	logger.Infof("leader:%v", op.pbft.leader-uint64(1))
	for c, _ := range op.pbft.clerk {
		logger.Infof("clerk:%v", c)
	}

	//主节点知道自己是主节点后，要向其他节点说一下，且开启倒计时
	if op.pbft.primary(op.pbft.view) == op.pbft.id {
		data := []byte{}
		data = append(data, byte(op.pbft.view))
		msg := &pb.Message{
			Type:    pb.Message_LEADER,
			Payload: data,
		}
		op.broadcaster.Broadcast(msg)

		//开启发送交易列表倒计时
		op.startBatchTimer()
	}
}

/*
***计票节点计票功能
 */
func (instance *pbftCore) recvPrepareClerk(prep *Prepare) error {
	if instance.activeView == false {
		return nil
	}
	logger.Infof("ifSenfFeedback:%v", instance.ifSendFeedback)
	if instance.ifSendFeedback == true {
		return nil
	}

	//如果是计票员
	//for c,_ := range(instance.clerk){
	//if instance.id == c{

	if instance.primary(prep.View) == prep.ReplicaId {
		logger.Infof("Replica %d received prepare from primary, ignoring", instance.id)
		return nil
	}

	/*if !instance.inWV(prep.View, prep.SequenceNumber) {
		if prep.SequenceNumber != instance.h && !instance.skipInProgress {
			logger.Warningf("Replica %d ignoring prepare for view=%d/seqNo=%d: not in-wv, in view %d, low water mark %d", instance.id, prep.View, prep.SequenceNumber, instance.view, instance.h)
		} else {
			// This is perfectly normal
			logger.Warningf("Replica %d ignoring prepare for view=%d/seqNo=%d: not in-wv, in view %d, low water mark %d", instance.id, prep.View, prep.SequenceNumber, instance.view, instance.h)
		}
		return nil
	}*/

	if instance.view != prep.View { //上个视图中的prepare消息
		logger.Infof("prepare from last view,ignore.")
		return nil
	}

	logger.Infof("Clerk %d received prepare message from replica %d,have %v,want %v", instance.id, prep.ReplicaId, len(instance.clerkPrepares)+1, instance.f*2)

	instance.pdigest = prep.BatchDigest
	instance.pview = prep.View
	instance.pseq = prep.SequenceNumber
	//instance.prePrepare = prep.PrePrepare

	//主节点给不同的节点发送不同的交易列表
	if instance.pdigest != prep.BatchDigest {
		logger.Errorf("instance.pdigest:%s, prep.BatchDigest:%s", instance.pdigest, prep.BatchDigest)
		data := []byte{}

		ocMsg := &pb.Message{
			Type:    pb.Message_LEADERCHEAT,
			Payload: data,
		}
		instance.innerBroadcastFromClerk(ocMsg)
		instance.activeView = false
		return nil
	}

	instance.clerkPrepares = append(instance.clerkPrepares, prep.ReplicaId)
	instance.persistPSet()

	//if instance.prepared(prep.BatchDigest, prep.View, prep.SequenceNumber){
	//收到足够多的prepare消息，返回投票结果
	if len(instance.clerkPrepares) >= instance.f*2 {
		instance.ifSendFeedback = true
		//instance.innerBroadcast(&Message{Payload: &Message_PrePrepare{PrePrepare: cert.prePrepare}}, 10)
		//time.Sleep(instance.waitTimeout)

		feedback := &Feedback{
			//PrePrepare:prep.PrePrepare,
			Id:             instance.id,
			View:           instance.view,
			SequenceNumber: instance.pseq,
			BatchDigest:    instance.pdigest,
			Replicas:       instance.clerkPrepares,
		}

		msg := &Message{Payload: &Message_Feedback{Feedback: feedback}}
		instance.innerBroadcast(msg, 2)

		//instance.prepareFeedback(cert)

		//计票员自己肯定知道结果
		instance.dealFeedback(feedback)
		//instance.FeedbackNum++
		//if instance.FeedbackNum >= len(instance.clerk){
		//if instance.FeedbackNum >= 1{

		//}

	}

	//}
	//}

	return nil
}

/*
***把cert.prepare封装
 */
func (instance *pbftCore) getPayload(cert *msgCert) []byte {
	data := []byte{}
	flag := false

	for _, pre := range cert.prepare {
		if flag == false {
			data = append(data, byte(pre.View))
			flag = true
		}

		//s := pre.String()
		b1 := make([]byte, 8)
		binary.BigEndian.PutUint64(b1, pre.View)
		for i := 0; i < len(b1); i++ {
			data = append(data, b1[i])
		}
		binary.BigEndian.PutUint64(b1, pre.SequenceNumber)
		for i := 0; i < len(b1); i++ {
			data = append(data, b1[i])
		}
		binary.BigEndian.PutUint64(b1, pre.ReplicaId)
		for i := 0; i < len(b1); i++ {
			data = append(data, b1[i])
		}
		b2 := []byte(pre.BatchDigest)
		data = append(data, byte(len(b2)))
		for i := 0; i < len(b2); i++ {
			data = append(data, b2[i])
		}
	}
	return data
}

/*
***计票节点返回计票结果
***把cert.prepare封装到消息中传给其他节点
 */
func (instance *pbftCore) prepareFeedback(cert *msgCert) {

	logger.Infof("Clerk %d send prepare-feedback", instance.id)
	instance.ifSendFeedback = true

	//把cert.prepare转为[]byte（clerk.id, pre.view, pre.seq, pre.id, len(digest), pre.digest）
	data := instance.getPayload(cert)

	//发送消息
	ocMsg := &pb.Message{
		Type:    pb.Message_FEEDBACK,
		Payload: data,
	}
	instance.innerBroadcastFromClerk(ocMsg)
}

/*
***节点接收到投票结果后的处理
 */
func (instance *pbftCore) dealFeedback(feedback *Feedback) error {
	logger.Infof("ifKnowResult:%v", instance.ifKnowResult)

	if instance.ifKnowResult == true {
		return nil
	}

	_, ok := instance.clerk[feedback.Id]
	if !ok {
		logger.Warningf("receive feedback not from clerk,ignore.")
		//	return nil
	}
	if feedback.View != instance.view {
		logger.Warningf("receive feedback from last view,ignore.")
		return nil
	}

	logger.Infof("replica %d receive prepare-feedback, have %v, want %v", instance.id, instance.FeedbackNum+1, len(instance.clerk))

	instance.stopClerkFeedbackTimer()
	instance.FeedbackNum++
	if instance.FeedbackNum >= len(instance.clerk) {
		//if instance.FeedbackNum >= 1{

		instance.ifKnowResult = true
		logger.Infof("replica %d receive enough prepare-feedback", instance.id)

		cert := instance.getCert(feedback.View, feedback.SequenceNumber)

		for _, rid := range feedback.GetReplicas() {
			prep := &Prepare{
				View:           feedback.View,
				SequenceNumber: feedback.SequenceNumber,
				BatchDigest:    feedback.BatchDigest,
				ReplicaId:      rid,
				//PrePrepare:     pre,
			}
			cert.prepare = append(cert.prepare, prep)
			commit := &Commit{
				View:           feedback.View,
				SequenceNumber: feedback.SequenceNumber,
				BatchDigest:    feedback.BatchDigest,
				ReplicaId:      rid,
			}
			cert.commit = append(cert.commit, commit)
		}
		//加上主节点的commit
		commit2 := &Commit{
			View:           feedback.View,
			SequenceNumber: feedback.SequenceNumber,
			BatchDigest:    feedback.BatchDigest,
			ReplicaId:      instance.leader - uint64(1),
		}
		cert.commit = append(cert.commit, commit2)
		if cert.digest == "" {
			cert.digest = commit2.BatchDigest
		}

		if cert.prePrepare != nil {
			instance.commitBatch(cert.prePrepare.View, cert.prePrepare.SequenceNumber)
			return nil
		}

	} else {
		instance.startClerkFeedbackTimer()
	}
	return nil
}

func (instance *pbftCore) commitBatch(view uint64, seq uint64) error {
	if instance.view != view {
		logger.Infof("have committed,ignore.")
		return nil
	}
	cert := instance.getCert(view, seq)

	//为重新选主做准备
	instance.ifSendVrf = false
	instance.leader = uint64(0)
	for index, _ := range instance.clerk {
		delete(instance.clerk, index)
	}
	instance.view += 1

	instance.seqNo += 1

	instance.FeedbackNum = 0
	instance.notConsensused += len(cert.prePrepare.RequestBatch.Batch)
	instance.vrfs = []uint64{}
	for index, _ := range instance.vrf_peers {
		delete(instance.vrf_peers, index)
	}
	instance.leaderCheatingNum = 0
	instance.pdigest = ""
	instance.pview = 0
	instance.pseq = 0
	instance.prePrepare = nil
	instance.ifSendFeedback = false
	instance.prepares = []*Prepare3{}
	instance.clerkPrepares = []uint64{}
	instance.ifKnowResult = false

	//告诉其他节点自己已经共识成功
	instance.okNum++
	data1 := []byte{}
	data1 = append(data1, byte(instance.view))
	ocMsg := &pb.Message{
		Type:    pb.Message_COK,
		Payload: data1,
	}
	instance.consumer.unicastToOne(ocMsg, uint64(0))

	if instance.id == uint64(0) {
		logger.Warningf("next consensused req:%v, amount req:%v", instance.notConsensused, len(instance.clientRequests))
		if instance.notConsensused < len(instance.clientRequests) {
			instance.startWaitTimer()
		}
	}

	//上链
	logger.Warningf("ready to commit %d reqs this time.", len(cert.prePrepare.RequestBatch.Batch))
	instance.stopTimer()
	instance.lastNewViewTimeout = instance.newViewTimeout
	delete(instance.outstandingReqBatches, string(cert.prePrepare.BatchDigest))
	//instance.executeOutstanding2(cert.prePrepare.View, cert.prePrepare.SequenceNumber)
	//instance.executeOutstanding()

	//instance.waitedCerts = append(instance.waitedCerts)
	//logger.Warningf("waitedCerts:%v, waited:%v", len(instance.waitedCerts), instance.waited)
	//logger.Warningf("ifExec:%v", instance.ifExec)
	if instance.currentExec == nil && instance.waited == len(instance.waitedCerts) { //当前没有再提交的区块
		logger.Warningf("nil")
		instance.executeOutstanding2(cert.prePrepare.View, cert.prePrepare.SequenceNumber)
	} else {
		logger.Warningf("not nil")
		instance.waitedCerts = append(instance.waitedCerts, cert)
	}

	if cert.prePrepare.SequenceNumber == instance.viewChangeSeqNo {
		logger.Infof("Replica %d cycling view for seqNo=%d", instance.id, cert.prePrepare.SequenceNumber)
		instance.sendViewChange()
	}
	return nil

}

func (instance *pbftCore) sendPrepare2() error {
	instance.ifAsking = true
	prep := &Prepare2{
		View:           instance.pview,
		SequenceNumber: instance.pseq,
		BatchDigest:    instance.pdigest,
		ReplicaId:      instance.id,
	}
	msg := &Message{Payload: &Message_Prepare2{Prepare2: prep}}
	msgRaw, err := proto.Marshal(msg)
	if err != nil {
		logger.Errorf("errors when sending ask.%v", err)
		return nil
	}
	instance.consumer.broadcast(msgRaw)
	return nil
	//return instance.innerBroadcast(&Message{Payload: &Message_Prepare2{Prepare2: prep}}, 2)
}

func (instance *pbftCore) innerBroadcastFromClerk(msg *pb.Message) error {
	doByzantine := false
	/*if instance.byzantine {
		rand1 := math.New(math.NewSource(time.Now().UnixNano()))
		doIt := rand1.Intn(3) // go byzantine about 1/3 of the time
		if doIt == 1 {
			doByzantine = true
		}
	}*/
	if instance.byzantine {
		if instance.id != uint64(0) {
			if instance.id == uint64(instance.f) {
				doByzantine = true
			}
		}
	}

	// testing byzantine fault.
	if doByzantine {

		rand2 := math.New(math.NewSource(time.Now().UnixNano()))
		ignoreidx := rand2.Intn(instance.N)
		for i := 0; i < instance.N; i++ {
			if i != ignoreidx && uint64(i) != instance.id { //Pick a random replica and do not send message
				instance.consumer.unicastToOne(msg, uint64(i))
			} else {
				logger.Warningf("PBFT byzantine: not broadcasting to replica %v", i)
				instance.consumer.unicastToOne(msg, uint64(i))
			}
		}
	} else {
		instance.consumer.broadcastToAll(msg)
	}
	return nil
}

func (instance *pbftCore) recvPrepare2(prep *Prepare2) error {
	logger.Infof("Replica %d received ask from replica %d for view=%d/seqNo=%d",
		instance.id, prep.ReplicaId, prep.View, prep.SequenceNumber)
	pid := prep.ReplicaId
	cert := instance.getCert(prep.View, prep.SequenceNumber)
	pre := cert.prePrepare
	if pre != nil {
		logger.Infof("answer yes.")

		prep := &Prepare3{
			View:           pre.View,
			SequenceNumber: pre.SequenceNumber,
			BatchDigest:    pre.BatchDigest,
			ReplicaId:      instance.id,
		}
		msg := &Message{Payload: &Message_Prepare3{Prepare3: prep}}
		msgRaw, err := proto.Marshal(msg)
		if err != nil {
			logger.Errorf("errors when sending answer,%v.", err)
		} else {
			instance.consumer.unicast(msgRaw, pid)
		}
	}
	return nil

}

//询问者收到回复
func (instance *pbftCore) recvPrepare3(prep *Prepare3) error {

	if instance.ifKnowResult == true {
		logger.Infof("already know the result")
		return nil
	}
	logger.Infof("Replica %d received answer from replica %d for view=%d/seqNo=%d, have %d, want %d",
		instance.id, prep.ReplicaId, prep.View, prep.SequenceNumber, len(instance.prepares)+1, instance.f*2-1)
	//if instance.ifAsking == true{
	instance.prepares = append(instance.prepares, prep)
	if len(instance.prepares) >= instance.f*2-1 { //收到足够多的prepare2消息

		cert := instance.getCert(prep.View, prep.SequenceNumber)
		for _, prevPrep := range instance.prepares { //补充prepare
			pre := &Prepare{
				View:           prevPrep.View,
				SequenceNumber: prevPrep.SequenceNumber,
				BatchDigest:    prevPrep.BatchDigest,
				ReplicaId:      instance.id,
				//PrePrepare:     instance.prePrepare,
			}

			cert.prepare = append(cert.prepare, pre)
		}

		//补充cert.commit
		for _, pre2 := range cert.prepare {
			commit := &Commit{
				View:           pre2.View,
				SequenceNumber: pre2.SequenceNumber,
				BatchDigest:    pre2.BatchDigest,
				ReplicaId:      pre2.ReplicaId,
			}
			cert.commit = append(cert.commit, commit)
		}

		//加入主节点的commit
		commit := &Commit{
			View:           prep.View,
			SequenceNumber: prep.SequenceNumber,
			BatchDigest:    prep.BatchDigest,
			ReplicaId:      instance.leader - uint64(1),
		}
		cert.commit = append(cert.commit, commit)
		if cert.digest == "" {
			cert.digest = commit.BatchDigest
		}

		instance.ifKnowResult = true

		if cert.prePrepare != nil {
			instance.commitBatch(cert.prePrepare.View, cert.prePrepare.SequenceNumber)
		}

	}
	//}

	return nil
}
