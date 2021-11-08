// Copyright 2016 The go-ethereum Authors
// This file is part of the go-ethereum library.
//
// The go-ethereum library is free software: you can redistribute it and/or modify
// it under the terms of the GNU Lesser General Public License as published by
// the Free Software Foundation, either version 3 of the License, or
// (at your option) any later version.
//
// The go-ethereum library is distributed in the hope that it will be useful,
// but WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE. See the
// GNU Lesser General Public License for more details.
//
// You should have received a copy of the GNU Lesser General Public License
// along with the go-ethereum library. If not, see <http://www.gnu.org/licenses/>.

package core

import (
	"testing"

	"github.com/PlatONEnetwork/PlatONE-Go/common"
	"github.com/PlatONEnetwork/PlatONE-Go/crypto"
)

func TestTxQueuedMap(t *testing.T) {
	queueMap, txHashList := makeTxQueuedMapWithTxs(100)

	if queueMap.Len() != 100 {
		t.Errorf("queueMap transaction mismatch: have %d, want %d", queueMap.Len(), 100)
	}

	queueMap.Remove(txHashList[3])
	if queueMap.data.Len() != 99 {
		t.Errorf("queueMap transaction mismatch: have %d, want %d", queueMap.Len(), 99)
	}

	_, count := queueMap.GetByCount(50)
	if count != 50 {
		t.Errorf("queueMap get transaction mismatch: got %d, want %d", queueMap.Len(), 50)
	}

	txs := queueMap.Get()
	if len(txs) != 99 {
		t.Errorf("queueMap get transaction mismatch: got %d, want %d", queueMap.Len(), 99)
	}
}

func makeTxQueuedMapWithTxs(txsCount int) (*txQueuedMap, []common.Hash) {
	queueMap := newTxQueuedMap()
	key, _ := crypto.GenerateKey()
	txsHash := make([]common.Hash, 0, txsCount)

	for nonce := 0; nonce < txsCount; nonce++ {
		tx := transaction(uint64(nonce), 0, key)
		queueMap.Put(tx.Hash(), tx)
		txsHash = append(txsHash, tx.Hash())
	}

	return queueMap, txsHash
}
