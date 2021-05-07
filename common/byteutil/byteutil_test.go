package byteutil

import (
	"testing"
)

func TestConverter(t *testing.T) {
	var u64 uint64 = 1
	b := Uint64ToBytes(u64)
	u32 := BytesToUint32(b)

	if u64 != uint64(u32) {
		t.Fatal("failed")
	}
}
