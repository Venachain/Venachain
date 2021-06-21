package crypto

import (
	"fmt"
	"math/big"
	"testing"
)

func TestVerify(t *testing.T) {
	witness1 := big.NewInt(7)
	witness2 := big.NewInt(-1)
	witness3 := big.NewInt(0)
	Bp()
	proof1, _ := Prove(witness1, bpparam)
	res1, _ := proof1.Verify(bpparam)
	proof2, _ := Prove(witness2, bpparam)
	res2, _ := proof2.Verify(bpparam)
	proof3, _ := Prove(witness3, bpparam)
	res3, _ := proof3.Verify(bpparam)
	witness4 := big.NewInt(4294967290)
	proof4, _ := Prove(witness4, bpparam)
	res4, _ := proof4.Verify(bpparam)
	fmt.Println(res1, res2, res3, res4)
}
