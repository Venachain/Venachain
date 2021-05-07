package core

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"math/big"
	"time"

	"github.com/PlatONEnetwork/PlatONE-Go/common"
	"github.com/PlatONEnetwork/PlatONE-Go/common/syscontracts"
	"github.com/PlatONEnetwork/PlatONE-Go/core/rawdb"
	"github.com/PlatONEnetwork/PlatONE-Go/core/types"
	"github.com/PlatONEnetwork/PlatONE-Go/core/vm"
	"github.com/PlatONEnetwork/PlatONE-Go/life/utils"
	"github.com/PlatONEnetwork/PlatONE-Go/log"
	"github.com/PlatONEnetwork/PlatONE-Go/rlp"
)

// Body is a simple (mutable, non-safe) data container for storing and moving
// a block's data contents (transactions) together.
type OldBody struct {
	Transactions []*OldTransaction
}

type OldTransaction struct {
	data OldTxdata
}

type OldTxdata struct {
	AccountNonce uint64          `json:"nonce"    gencodec:"required"`
	Price        *big.Int        `json:"gasPrice" gencodec:"required"`
	GasLimit     uint64          `json:"gas"      gencodec:"required"`
	Recipient    *common.Address `json:"to"       rlp:"nil"` // nil means contract creation
	Amount       *big.Int        `json:"value"    gencodec:"required"`
	Payload      []byte          `json:"input"    gencodec:"required"`

	TxType uint64 `json:"txType" gencodec:"required"`

	// Signature values
	V *big.Int `json:"v" gencodec:"required"`
	R *big.Int `json:"r" gencodec:"required"`
	S *big.Int `json:"s" gencodec:"required"`

	// This is only used when marshaling to JSON.
	Hash *common.Hash `json:"hash" rlp:"-"`
}

// DecodeRLP implements rlp.Decoder
func (tx *OldTransaction) DecodeRLP(s *rlp.Stream) error {
	err := s.Decode(&tx.data)
	return err
}

// Export writes the active chain to the given writer.
func (bc *BlockChain) Export(w io.Writer, version string) error {
	return bc.ExportN(w, version, uint64(0), bc.CurrentBlock().NumberU64())
}

// ExportN writes a subset of the active chain to the given writer.
func (bc *BlockChain) ExportN(w io.Writer, version string, first uint64, last uint64) error {
	bc.mu.RLock()
	defer bc.mu.RUnlock()

	if first > last {
		return fmt.Errorf("export failed: first (%d) is greater than last (%d)", first, last)
	}
	log.Info("Exporting batch of blocks", "count", last-first+1)

	// export pivot
	rlp.Encode(w, last)

	//export possible old system contracts
	var addrSuperAdmin string
	m := make(map[common.Address]string)
	for k, v := range vm.CnsSysContractsMap {
		if version < "1.0.0" {
			input := common.GenCallData("getContractAddress", []interface{}{k, "latest"})
			btsRes, err := bc.RunInterpreterDirectly(common.Address{}, syscontracts.CnsManagementAddress, input)
			if err != nil {
				continue
			}
			strRes := common.CallResAsString(btsRes)
			if len(strRes) == 0 || common.IsHexZeroAddress(strRes) {
				continue
			}
			taddr := common.HexToAddress(strRes)
			m[taddr] = k

			if k == "__sys_RoleManager" {
				input = common.GenCallData("getAccountsByRole", []interface{}{"chainCreator"})
				btsRes, err = bc.RunInterpreterDirectly(common.Address{}, taddr, input)
				if err != nil {
					continue
				}
				strRes = common.CallResAsString(btsRes)

				type tmpResInfo struct {
					Name    string `json:"name"`
					Address string `json:"address"`
				}
				type tmpResType struct {
					Code int          `json:"code"`
					Msg  string       `json:"msg"`
					Data []tmpResInfo `json:"data"`
				}
				var tmp tmpResType
				if err := json.Unmarshal(utils.String2bytes(strRes), &tmp); err != nil || tmp.Code != 0 || len(tmp.Data) < 1 {
					continue
				}
				addrSuperAdmin = tmp.Data[0].Address
			}
		} else {
			m[v] = k
		}
	}
	b, err := json.Marshal(m)
	if err != nil {
		return err
	}
	rlp.Encode(w, b)
	rlp.Encode(w, addrSuperAdmin)

	start, reported := time.Now(), time.Now()
	for nr := first; nr <= last; nr++ {
		hash := rawdb.ReadCanonicalHash(bc.db, nr)
		if hash == (common.Hash{}) {
			return fmt.Errorf("export failed on #%d: not found", nr)
		}
		block := bc.GetBlockMaybeOld(hash, nr, version)
		if block == nil {
			return fmt.Errorf("export failed on #%d: not found", nr)
		}
		if err := block.EncodeRLP(w); err != nil {
			return err
		}
		if time.Since(reported) >= statsReportLimit {
			log.Info("Exporting blocks", "exported", block.NumberU64()-first, "elapsed", common.PrettyDuration(time.Since(start)))
			reported = time.Now()
		}
	}
	return nil
}

// GetBlockByNumber retrieves a block from the database by number, caching it
// (associated with its hash) if found.
func (bc *BlockChain) GetBlockMaybeOld(hash common.Hash, number uint64, version string) *types.Block {
	if version < "1.0.0" {
		header := rawdb.ReadHeader(bc.db, hash, number)
		if header == nil {
			return nil
		}

		data := rawdb.ReadBodyRLP(bc.db, hash, number)
		if len(data) == 0 {
			return nil
		}
		body := new(OldBody)
		if err := rlp.Decode(bytes.NewReader(data), body); err != nil {
			log.Error("Invalid block body RLP", "hash", hash, "err", err)
			return nil
		}
		if body == nil {
			return nil
		}

		txs := make([]*types.Transaction, 0)
		for _, tx := range body.Transactions {
			from, err := bc.getSenderFromOldTx(tx)
			if err != nil {
				log.Error("Invalid transaction Sender", "block", number, "hash", hash, "err", err)
			}
			payload := append(append(types.OldTxPrefix, from.Bytes()...), tx.data.Payload...)
			if tx.data.Recipient == nil {
				txs = append(txs, types.NewContractCreation(tx.data.AccountNonce, tx.data.Amount, tx.data.GasLimit, tx.data.Price, payload))
			} else {
				txs = append(txs, types.NewTransaction(tx.data.AccountNonce, *(tx.data.Recipient), tx.data.Amount, tx.data.GasLimit, tx.data.Price, payload))
			}
		}
		block := types.NewBlockWithHeader(header).WithBody(txs)
		return block
	}

	return bc.GetBlock(hash, number)
}

func (bc *BlockChain) getSenderFromOldTx(tx *OldTransaction) (common.Address, error) {
	isProtected := true
	V := tx.data.V
	if V.BitLen() <= 8 {
		v := V.Uint64()
		isProtected = v != 27 && v != 28
	}

	var sighash common.Hash
	if !isProtected {
		sighash = common.RlpHash([]interface{}{
			tx.data.AccountNonce,
			tx.data.Price,
			tx.data.GasLimit,
			tx.data.Recipient,
			tx.data.Amount,
			tx.data.Payload,
			tx.data.TxType,
		})
	} else {
		V = new(big.Int).Sub(V, new(big.Int).Mul(bc.chainConfig.ChainID, big.NewInt(2)))
		V.Sub(V, big.NewInt(8))

		sighash = common.RlpHash([]interface{}{
			tx.data.AccountNonce,
			tx.data.Price,
			tx.data.GasLimit,
			tx.data.Recipient,
			tx.data.Amount,
			tx.data.Payload,
			tx.data.TxType,
			bc.chainConfig.ChainID, uint(0), uint(0),
		})
	}
	return types.RecoverPlain(sighash, tx.data.R, tx.data.S, V, true)
}
