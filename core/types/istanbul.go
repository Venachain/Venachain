// Copyright 2017 The go-ethereum Authors
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

package types

import (
	"errors"
	"io"

	"github.com/Venachain/Venachain/common"
	"github.com/Venachain/Venachain/rlp"
)

var (
	// IrisDigest represents a hash of "Istanbul practical byzantine fault tolerance"
	// to identify whether the block is from Istanbul consensus engine
	IrisDigest = common.HexToHash("0x63746963616c2062797a616e74696e65206661756c7420746f6c6572616e6365")

	IrisExtraVanity = 32 // Fixed number of extra-data bytes reserved for validator vanity
	IrisExtraSeal   = 65 // Fixed number of extra-data bytes reserved for validator seal

	// ErrInvalidIrisHeaderExtra is returned if the length of extra-data is less than 32 bytes
	ErrInvalidIrisHeaderExtra = errors.New("invalid iris header extra-data")
)

type IrisExtra struct {
	Validators    []common.Address
	Seal          []byte
	CommittedSeal [][]byte
}

// EncodeRLP serializes ist into the Ethereum RLP format.
func (ist *IrisExtra) EncodeRLP(w io.Writer) error {
	return rlp.Encode(w, []interface{}{
		ist.Validators,
		ist.Seal,
		ist.CommittedSeal,
	})
}

// DecodeRLP implements rlp.Decoder, and load the istanbul fields from a RLP stream.
func (ist *IrisExtra) DecodeRLP(s *rlp.Stream) error {
	var irisExtra struct {
		Validators    []common.Address
		Seal          []byte
		CommittedSeal [][]byte
	}
	if err := s.Decode(&irisExtra); err != nil {
		return err
	}
	ist.Validators, ist.Seal, ist.CommittedSeal = irisExtra.Validators, irisExtra.Seal, irisExtra.CommittedSeal
	return nil
}

// ExtractIrisExtra extracts all values of the IrisExtra from the header. It returns an
// error if the length of the given extra-data is less than 32 bytes or the extra-data can not
// be decoded.
func ExtractIrisExtra(h *Header) (*IrisExtra, error) {
	if len(h.Extra) < IrisExtraVanity {
		return nil, ErrInvalidIrisHeaderExtra
	}

	var irisExtra *IrisExtra
	err := rlp.DecodeBytes(h.Extra[IrisExtraVanity:], &irisExtra)
	if err != nil {
		return nil, err
	}
	return irisExtra, nil
}

// IrisFilteredHeader returns a filtered header which some information (like seal, committed seals)
// are clean to fulfill the Istanbul hash rules. It returns nil if the extra-data cannot be
// decoded/encoded by rlp.
func IrisFilteredHeader(h *Header, keepSeal bool) *Header {
	newHeader := CopyHeader(h)
	irisExtra, err := ExtractIrisExtra(newHeader)
	if err != nil {
		return nil
	}

	if !keepSeal {
		irisExtra.Seal = []byte{}
	}
	irisExtra.CommittedSeal = [][]byte{}

	payload, err := rlp.EncodeToBytes(&irisExtra)
	if err != nil {
		return nil
	}

	newHeader.Extra = append(newHeader.Extra[:IrisExtraVanity], payload...)

	return newHeader
}
