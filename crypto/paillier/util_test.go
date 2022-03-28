package paillier

import (
	"testing"
)

func TestPaillierWeightAdd(t *testing.T) {
	str := []string{
		"3332efac4d4e0dae3f8f6aa2adac08a367fcbb445e23f895c4f1e851a29e449c3a9fecd2aef21d71a4d8fd176b45d283360a921b761370335895eee4de1fb863",
		"22923c0e2999842a9802cdd28804bf6ead5ba99c3e17ffbda7ef9dfc5ad2e3074ff6ac88be64705f64e4799f2a925093f5b113096ffc76313b39a1662ee7eb2f",
		"2452d327ba5100ae18aceb72fa2e360d7a2127f63b7cb922d59d90a98918bc21ce707c5ab9db6b774c3942decc8542501910393d09e24d9a5ce7c9e08ab9368c",
	}
	arr := []uint{3, 5, 2}
	result, err := PaillierWeightAdd(str, arr, "95812e8c3bafc54adae58e01e2777a3de717312e2efe8b764c3a684a46106db7")
	if err != nil {
		t.Errorf("error : %s\n", err.Error())
		return
	}
	if result == "" {
		t.Error("result is null!")
		return
	}
	t.Logf("WeightAdd Sum Cipher = %s\n", result)
}
func TestPaillierAdd(t *testing.T) {
	str := []string{
		"3332efac4d4e0dae3f8f6aa2adac08a367fcbb445e23f895c4f1e851a29e449c3a9fecd2aef21d71a4d8fd176b45d283360a921b761370335895eee4de1fb863",
		"22923c0e2999842a9802cdd28804bf6ead5ba99c3e17ffbda7ef9dfc5ad2e3074ff6ac88be64705f64e4799f2a925093f5b113096ffc76313b39a1662ee7eb2f",
		"2452d327ba5100ae18aceb72fa2e360d7a2127f63b7cb922d59d90a98918bc21ce707c5ab9db6b774c3942decc8542501910393d09e24d9a5ce7c9e08ab9368c",
	}

	result, err := PaillierAdd(str, "95812e8c3bafc54adae58e01e2777a3de717312e2efe8b764c3a684a46106db7")
	if err != nil {
		t.Errorf("error : %s\n", err.Error())
		return
	}
	if result == "" {
		t.Error("result is null!")
		return
	}
	t.Logf("Add Sum Cipher = %s\n", result)
}

func TestPaillierMulPro(t *testing.T) {
	str := "3332efac4d4e0dae3f8f6aa2adac08a367fcbb445e23f895c4f1e851a29e449c3a9fecd2aef21d71a4d8fd176b45d283360a921b761370335895eee4de1fb863"
	arr := uint(3)
	result, err := PaillierMul(str, arr, "95812e8c3bafc54adae58e01e2777a3de717312e2efe8b764c3a684a46106db7")
	if err != nil {
		t.Errorf("error : %s\n", err.Error())
		return
	}
	if result == "" {
		t.Error("result is null!")
		return
	}
	t.Logf("MulPro Sum Cipher = %s\n", result)
}
