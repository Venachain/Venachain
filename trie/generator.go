package trie

import (
	"bytes"
	"hash"

	"github.com/Venachain/Venachain/common"
	"github.com/Venachain/Venachain/crypto/sha3"
	"github.com/Venachain/Venachain/rlp"
)

type Generator interface {
	AddItem(index int, value []byte)

	Hash() common.Hash
}

var (
	keyMap = make(map[int][]byte, 40960)
)

type HashValue struct {
	hash hash.Hash
}

func NewHashValue() *HashValue {
	return &HashValue{
		hash: sha3.NewKeccak256(),
	}
}

func (h HashValue) AddItem(index int, value []byte) {
	h.hash.Write(value)
}

func (h HashValue) Hash() common.Hash {
	var res common.Hash
	h.hash.Sum(res[:0])
	return res
}

type HashTrie struct {
	data   *Trie
	keyBuf *bytes.Buffer
}

func NewHashTrie() *HashTrie {
	return &HashTrie{
		data:   new(Trie),
		keyBuf: new(bytes.Buffer),
	}
}

func (h *HashTrie) AddItem(index int, value []byte) {
	indexByte, ok := keyMap[index]
	if !ok {
		h.keyBuf.Reset()
		rlp.Encode(h.keyBuf, uint(index))
		tmp := h.keyBuf.Bytes()
		indexByte = make([]byte, len(tmp))
		copy(indexByte, tmp)
		keyMap[index] = indexByte
	}
	h.data.Update(indexByte, value)
}

func (h *HashTrie) Hash() common.Hash {
	return h.data.Hash()
}
