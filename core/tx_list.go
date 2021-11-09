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
	"container/list"
	"sync"

	"github.com/PlatONEnetwork/PlatONE-Go/common"
	"github.com/PlatONEnetwork/PlatONE-Go/core/types"
)

// txQueueMap is a txHash -> transaction hash map
type txQueuedMap struct {
	mu    *sync.RWMutex
	items map[common.Hash]*list.Element // Hash map storing the transaction data
	data  *list.List                    // a transaction queue
}

func newTxQueuedMap() *txQueuedMap {
	return &txQueuedMap{
		mu:    &sync.RWMutex{},
		items: make(map[common.Hash]*list.Element),
		data:  list.New(),
	}
}

func (m *txQueuedMap) Get() types.Transactions {
	m.mu.RLock()
	defer m.mu.RUnlock()

	txs := make(types.Transactions, 0, m.data.Len())
	for e := m.data.Front(); e != nil; e = e.Next() {
		if tx, ok := e.Value.(*types.Transaction); ok {
			txs = append(txs, tx)
		}
	}
	if txs.Len() == 0 {
		return nil
	}
	return txs
}

func (m *txQueuedMap) GetByCount(max int) (types.Transactions, int) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	count := 0
	if max > m.data.Len() {
		max = m.data.Len()
	}
	txs := make(types.Transactions, 0, max)
	for e := m.data.Front(); e != nil && count < max; e = e.Next() {
		if tx, ok := e.Value.(*types.Transaction); ok {
			txs = append(txs, tx)
			count++
		}
	}
	if txs.Len() == 0 {
		return nil, count
	}

	return txs, count
}

func (m *txQueuedMap) Len() int {
	m.mu.RLock()
	defer m.mu.RUnlock()

	return m.data.Len()
}

func (m *txQueuedMap) Put(h common.Hash, tx *types.Transaction) {
	m.mu.Lock()
	defer m.mu.Unlock()

	_, ok := m.items[h]
	if ok {
		return
	}

	e := m.data.PushBack(tx)
	m.items[h] = e
}

func (m *txQueuedMap) Remove(h common.Hash) {
	m.mu.Lock()
	defer m.mu.Unlock()

	if e, ok := m.items[h]; ok {
		delete(m.items, h)
		m.data.Remove(e)
	}
}

func (m *txQueuedMap) RemoveTxs(txs types.Transactions) {
	m.mu.Lock()
	defer m.mu.Unlock()

	for _, tx := range txs {
		hash := tx.Hash()
		if e, ok := m.items[hash]; ok {
			delete(m.items, hash)
			m.data.Remove(e)
		}
	}
}
