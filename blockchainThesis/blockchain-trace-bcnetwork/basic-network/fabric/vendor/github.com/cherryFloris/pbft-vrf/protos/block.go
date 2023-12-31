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

package protos

import (
	"fmt"

	"github.com/cherryFloris/pbft-vrf/core/util"
	"github.com/golang/protobuf/proto"
)

// NewBlock creates a new block with the specified proposer ID, list of,
// transactions, and hash of the state calculated by calling State.GetHash()
// after running all transactions in the block and updating the state.
//
// TODO Remove proposerID parameter. This should be fetched from the config.
// TODO Remove the stateHash parameter. The transactions in this block should
// be run when blockchain.AddBlock() is called. This function will then update
// the stateHash in this block.
//
// func NewBlock(proposerID string, transactions []transaction.Transaction, stateHash []byte) *Block {
// 	block := new(Block)
// 	block.ProposerID = proposerID
// 	block.transactions = transactions
// 	block.stateHash = stateHash
// 	return block
// }

// Bytes returns this block as an array of bytes.
func (block *Block) Bytes() ([]byte, error) {
	data, err := proto.Marshal(block)
	if err != nil {
		logger.Errorf("Error marshalling block: %s", err)
		return nil, fmt.Errorf("Could not marshal block: %s", err)
	}
	return data, nil
}

// NewBlock creates a new Block given the input parameters.
func NewBlock(transactions []*Transaction, metadata []byte) *Block {
	block := new(Block)
	block.Transactions = transactions
	block.ConsensusMetadata = metadata
	return block
}

// GetHash returns the hash of this block.
func (block *Block) GetHash() ([]byte, error) {

	// copy the block and remove the non-hash data
	blockBytes, err := block.Bytes()
	if err != nil {
		return nil, fmt.Errorf("Could not calculate hash of block: %s", err)
	}
	blockCopy, err := UnmarshallBlock(blockBytes)
	if err != nil {
		return nil, fmt.Errorf("Could not calculate hash of block: %s", err)
	}
	blockCopy.NonHashData = nil

	// Hash the block
	data, err := proto.Marshal(blockCopy)
	if err != nil {
		return nil, fmt.Errorf("Could not calculate hash of block: %s", err)
	}
	hash := util.ComputeCryptoHash(data)
	return hash, nil
}

// GetStateHash returns the stateHash stored in this block. The stateHash
// is the value returned by state.GetHash() after running all transactions in
// the block.
func (block *Block) GetStateHash() []byte {
	return block.StateHash
}

// SetPreviousBlockHash sets the hash of the previous block. This will be
// called by blockchain.AddBlock when then the block is added.
func (block *Block) SetPreviousBlockHash(previousBlockHash []byte) {
	block.PreviousBlockHash = previousBlockHash
}

// UnmarshallBlock converts a byte array generated by Bytes() back to a block.
func UnmarshallBlock(blockBytes []byte) (*Block, error) {
	block := &Block{}
	err := proto.Unmarshal(blockBytes, block)
	if err != nil {
		logger.Errorf("Error unmarshalling block: %s", err)
		return nil, fmt.Errorf("Could not unmarshal block: %s", err)
	}
	return block, nil
}
