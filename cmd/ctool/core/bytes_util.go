package core

import (
	"bytes"
	"encoding/binary"
	"errors"
	"math"
	"math/big"
	"strconv"

	"github.com/Venachain/Venachain/common"
	math2 "github.com/Venachain/Venachain/common/math"
)

func BytesCombine(pBytes ...[]byte) []byte {
	return bytes.Join(pBytes, []byte(""))
}

func Int32ToBytes(n int32) []byte {
	tmp := int32(n)
	bytesBuffer := bytes.NewBuffer([]byte{})
	binary.Write(bytesBuffer, binary.BigEndian, tmp)
	return bytesBuffer.Bytes()
}

func BytesToInt32(b []byte) int32 {
	bytesBuffer := bytes.NewBuffer(b)
	var tmp int32
	binary.Read(bytesBuffer, binary.BigEndian, &tmp)
	return int32(tmp)
}

func Int64ToBytes(n int64) []byte {
	tmp := int64(n)
	bytesBuffer := bytes.NewBuffer([]byte{})
	binary.Write(bytesBuffer, binary.BigEndian, tmp)
	return bytesBuffer.Bytes()
}

func Uint64ToBytes(n uint64) []byte {
	tmp := uint64(n)
	bytesBuffer := bytes.NewBuffer([]byte{})
	binary.Write(bytesBuffer, binary.BigEndian, tmp)
	return bytesBuffer.Bytes()
}

func BytesToInt64(b []byte) int64 {
	bytesBuffer := bytes.NewBuffer(b)
	var tmp int64
	binary.Read(bytesBuffer, binary.BigEndian, &tmp)
	return int64(tmp)
}

func BytesToUint64(b []byte) uint64 {
	return binary.BigEndian.Uint64(b)
}

func Float32ToBytes(float float32) []byte {
	bits := math.Float32bits(float)
	bytes := make([]byte, 4)
	binary.BigEndian.PutUint32(bytes, bits)
	return bytes
}

func BytesToFloat32(bytes []byte) float32 {
	bits := binary.BigEndian.Uint32(bytes)
	return math.Float32frombits(bits)
}

func Float64ToBytes(float float64) []byte {
	bits := math.Float64bits(float)
	bytes := make([]byte, 8)
	binary.BigEndian.PutUint64(bytes, bits)
	return bytes
}

func BytesToFloat64(bytes []byte) float64 {
	bits := binary.BigEndian.Uint64(bytes)
	return math.Float64frombits(bits)
}

// bytes is made of 2 uint64 in big endian
func BytesToFloat128(bytes []byte) *big.Float {
	if len(bytes) < 16 {
		return &big.Float{}
	}
	if len(bytes) > 16 {
		bytes = bytes[:16]
	}

	low := binary.BigEndian.Uint64(bytes[8:])
	high := binary.BigEndian.Uint64(bytes[:8])

	F, _ := math2.NewFromBits(high, low).Big()

	return F
}

func BoolToBytes(b bool) []byte {
	buf := bytes.NewBuffer([]byte{})
	binary.Write(buf, binary.BigEndian, b)
	return buf.Bytes()
}

func BytesConverter(source []byte, t string) interface{} {
	switch t {
	case "int32":
		return common.CallResAsInt32(source)
	case "int64":
		return common.CallResAsInt64(source)
	case "int128":
		return common.CallResAsInt128(source)
	case "float32":
		return common.CallResAsFloat32(source)
	case "float64":
		return common.CallResAsFloat64(source)
	case "float128":
		return common.CallResAsFloat128(source)
	case "string", "int128_s", "uint128_s", "int256_s", "uint256_s":
		if len(source) < 64 {
			return string(source[:])
		} else {
			return string(source[64:])
		}
	case "uint128":
		return common.CallResAsUint128(source)
	case "uint64":
		return common.CallResAsUint64(source)
	case "uint32":
		return common.CallResAsUint32(source)
	default:
		return source
	}
}

func StringConverter(source string, t string) ([]byte, error) {
	switch t {
	case "int32", "uint32", "uint", "int":
		dest, err := strconv.Atoi(source)
		return Int32ToBytes(int32(dest)), err
	case "int64", "uint64":
		dest, err := strconv.ParseInt(source, 10, 64)
		return Int64ToBytes(dest), err
	case "int128", "uint128":
		I, success := new(big.Int).SetString(source, 10)
		if !success {
			return []byte(source), errors.New("parse string to int error")
		}
		if (t == "uint128" && I.Sign() < 0) || (t == "int128" && I.BitLen() > 127) {
			return []byte(source), errors.New("parse string to int error")
		}
		b, success := common.BigToByte128(I)
		if !success {
			return []byte(source), errors.New("parse string to int error")
		}
		return b, nil
	case "float32":
		dest, err := strconv.ParseFloat(source, 32)
		return Float32ToBytes(float32(dest)), err
	case "float64":
		dest, err := strconv.ParseFloat(source, 64)
		return Float64ToBytes(dest), err
	case "float128":
		F, _, err := big.ParseFloat(source, 10, math2.F128Precision, big.ToNearestEven)
		if err != nil {
			return []byte{}, err
		}
		F128, _ := math2.NewFromBig(F)
		return append(Uint64ToBytes(F128.High()), Uint64ToBytes(F128.Low())...), nil
	case "bool":
		if "true" == source || "false" == source {
			return BoolToBytes("true" == source), nil
		} else {
			return []byte{}, errors.New("invalid boolean param")
		}
	case "int128_s":
		I, success := new(big.Int).SetString(source, 10)
		if !success || I.BitLen() > 127 {
			return []byte(source), errors.New("not a valid number")
		}
		return []byte(source), nil
	case "uint128_s":
		I, success := new(big.Int).SetString(source, 10)
		if !success || I.BitLen() > 128 || I.Sign() < 0 {
			return []byte(source), errors.New("not a valid number")
		}
		return []byte(source), nil
	case "int256_s":
		I, success := new(big.Int).SetString(source, 10)
		if !success || I.BitLen() > 255 {
			return []byte(source), errors.New("not a valid number")
		}
		return []byte(source), nil
	case "uint256_s":
		I, success := new(big.Int).SetString(source, 10)
		if !success || I.BitLen() > 256 || I.Sign() < 0 {
			return []byte(source), errors.New("not a valid number")
		}
		return []byte(source), nil
	default:
		return []byte(source), nil
	}
}
