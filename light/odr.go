// Copyright 2015 The go-ethereum Authors
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

// Package light implements on-demand retrieval capable state and chain objects
// for the Ethereum Light Client.
package light

import (
	"context"
	"errors"

	"github.com/Venachain/Venachain/common"
	"github.com/Venachain/Venachain/core"
	"github.com/Venachain/Venachain/core/rawdb"
	"github.com/Venachain/Venachain/core/state"
	"github.com/Venachain/Venachain/core/types"
	"github.com/Venachain/Venachain/crypto"
	"github.com/Venachain/Venachain/ethdb/dbhandle"
	"github.com/Venachain/Venachain/log"
	"github.com/Venachain/Venachain/trie"
)

// NoOdr is the default context passed to an ODR capable function when the ODR
// service is not required.
var NoOdr = context.Background()

// ErrNoPeers is returned if no peers capable of serving a queued request are available
var ErrNoPeers = errors.New("no suitable peers available")

// OdrBackend is an interface to a backend service that handles ODR retrievals type
type OdrBackend interface {
	Database() dbhandle.Database
	ChtIndexer() *core.ChainIndexer
	BloomTrieIndexer() *core.ChainIndexer
	BloomIndexer() *core.ChainIndexer
	Retrieve(ctx context.Context, req OdrRequest) error
	RetrieveTxStatus(ctx context.Context, req *TxStatusRequest) (err error)
	IndexerConfig() *IndexerConfig
}

// OdrRequest is an interface for retrieval requests
type OdrRequest interface {
	StoreResult(db dbhandle.Database)
}

// TrieID identifies a state or account storage trie
type TrieID struct {
	BlockHash, Root common.Hash
	BlockNumber     uint64
	AccKey          []byte
}

// StateTrieID returns a TrieID for a state trie belonging to a certain block
// header.
func StateTrieID(header *types.Header) *TrieID {
	return &TrieID{
		BlockHash:   header.Hash(),
		BlockNumber: header.Number.Uint64(),
		AccKey:      nil,
		Root:        header.Root,
	}
}

// StorageTrieID returns a TrieID for a contract storage trie at a given account
// of a given state trie. It also requires the root hash of the trie for
// checking Merkle proofs.
func StorageTrieID(state *TrieID, addrHash, root common.Hash) *TrieID {
	return &TrieID{
		BlockHash:   state.BlockHash,
		BlockNumber: state.BlockNumber,
		AccKey:      addrHash[:],
		Root:        root,
	}
}

// TrieRequest is the ODR request type for state/storage trie entries
type TrieRequest struct {
	OdrRequest
	Id            *TrieID
	Key           []byte
	Proof         *NodeSet
	StorageValues [][]byte
}

// StoreResult stores the retrieved data in local database
func (req *TrieRequest) StoreResult(db dbhandle.Database) {
	req.Proof.Store(db)
	for _, storageValue := range req.StorageValues {
		valueKey := crypto.Keccak256Hash(append([]byte(state.StoragePrefix), storageValue...))
		db.Put(secureKey(valueKey[:]), storageValue)
		log.Debug("store result", "valueKey", valueKey, "value", req.StorageValues)
	}
}

func secureKey(key []byte) []byte {
	return append(trie.SecureKeyPrefix, key...)
}

// CodeRequest is the ODR request type for retrieving contract code
type CodeRequest struct {
	OdrRequest
	Id   *TrieID // references storage trie of the account
	Hash common.Hash
	Data []byte
}

// StoreResult stores the retrieved data in local database
func (req *CodeRequest) StoreResult(db dbhandle.Database) {
	db.Put(req.Hash[:], req.Data)
}

// BlockRequest is the ODR request type for retrieving block bodies
type BlockRequest struct {
	OdrRequest
	Hash   common.Hash
	Number uint64
	Rlp    []byte
}

// StoreResult stores the retrieved data in local database
func (req *BlockRequest) StoreResult(db dbhandle.Database) {
	rawdb.WriteBodyRLP(db, req.Hash, req.Number, req.Rlp)
}

// ReceiptsRequest is the ODR request type for retrieving block bodies
type ReceiptsRequest struct {
	OdrRequest
	Hash     common.Hash
	Number   uint64
	Receipts types.Receipts
}

// StoreResult stores the retrieved data in local database
func (req *ReceiptsRequest) StoreResult(db dbhandle.Database) {
	rawdb.WriteReceipts(db, req.Hash, req.Number, req.Receipts)
}

// ChtRequest is the ODR request type for state/storage trie entries
type ChtRequest struct {
	OdrRequest
	Config           *IndexerConfig
	ChtNum, BlockNum uint64
	ChtRoot          common.Hash
	Header           *types.Header
	Proof            *NodeSet
}

// StoreResult stores the retrieved data in local database
func (req *ChtRequest) StoreResult(db dbhandle.Database) {
	hash, num := req.Header.Hash(), req.Header.Number.Uint64()

	rawdb.WriteHeader(db, req.Header)
	rawdb.WriteCanonicalHash(db, hash, num)
}

// BloomRequest is the ODR request type for retrieving bloom filters from a CHT structure
type BloomRequest struct {
	OdrRequest
	Config           *IndexerConfig
	BloomTrieNum     uint64
	BitIdx           uint
	SectionIndexList []uint64
	BloomTrieRoot    common.Hash
	BloomBits        [][]byte
	Proofs           *NodeSet
}

// StoreResult stores the retrieved data in local database
func (req *BloomRequest) StoreResult(db dbhandle.Database) {
	for i, sectionIdx := range req.SectionIndexList {
		sectionHead := rawdb.ReadCanonicalHash(db, (sectionIdx+1)*req.Config.BloomTrieSize-1)
		// if we don't have the canonical hash stored for this section head number, we'll still store it under
		// a key with a zero sectionHead. GetBloomBits will look there too if we still don't have the canonical
		// hash. In the unlikely case we've retrieved the section head hash since then, we'll just retrieve the
		// bit vector again from the network.
		rawdb.WriteBloomBits(db, req.BitIdx, sectionIdx, sectionHead, req.BloomBits[i])
	}
}

type TxStatus struct {
	Status core.TxStatus
	Lookup *rawdb.TxLookupEntry `rlp:"nil"`
	Error  string
}

type TxStatusRequest struct {
	Hashes []common.Hash
	Status []TxStatus
}

// StoreResult stores the retrieved data in local database
func (req *TxStatusRequest) StoreResult(db dbhandle.Database) {}
