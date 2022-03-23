package vm

import (
	"github.com/Venachain/Venachain/common"
	"testing"
)

func TestPLPaillierAdd(t *testing.T) {
	address := []byte("0x80989fb9a8eb623dad541c1525828484e5fab75a")
	e := NewPLWrapper(nil)
	str := "3332efac4d4e0dae3f8f6aa2adac08a367fcbb445e23f895c4f1e851a29e449c3a9fecd2aef21d71a4d8fd176b45d283360a921b761370335895eee4de1fb863," +
			"22923c0e2999842a9802cdd28804bf6ead5ba99c3e17ffbda7ef9dfc5ad2e3074ff6ac88be64705f64e4799f2a925093f5b113096ffc76313b39a1662ee7eb2f, " +
			"2452d327ba5100ae18aceb72fa2e360d7a2127f63b7cb922d59d90a98918bc21ce707c5ab9db6b774c3942decc8542501910393d09e24d9a5ce7c9e08ab9368c"

	pub := "95812e8c3bafc54adae58e01e2777a3de717312e2efe8b764c3a684a46106db7"
	e.base.caller = common.BytesToAddress(address)

	res ,err := e.paillierAdd(str,pub)
	if res != "3481d75c52a08586ab68f4fa1ebd4189a6417e6796a8201ee58a2bc6565d11dcc526f319c34f7cf9c499b9c26cbc5e0bd4500a02174a91db34c8cff180f24417" || err != nil {
		t.Error("wrong res")
	}
}

func TestPaillierWeightAdd(t *testing.T) {
	address := []byte("0x80989fb9a8eb623dad541c1525828484e5fab75a")
	e := NewPLWrapper(nil)
	str := "3332efac4d4e0dae3f8f6aa2adac08a367fcbb445e23f895c4f1e851a29e449c3a9fecd2aef21d71a4d8fd176b45d283360a921b761370335895eee4de1fb863," +
		"22923c0e2999842a9802cdd28804bf6ead5ba99c3e17ffbda7ef9dfc5ad2e3074ff6ac88be64705f64e4799f2a925093f5b113096ffc76313b39a1662ee7eb2f, " +
		"2452d327ba5100ae18aceb72fa2e360d7a2127f63b7cb922d59d90a98918bc21ce707c5ab9db6b774c3942decc8542501910393d09e24d9a5ce7c9e08ab9368c"

	pub := "95812e8c3bafc54adae58e01e2777a3de717312e2efe8b764c3a684a46106db7"
	e.base.caller = common.BytesToAddress(address)

	res ,err := e.paillierWeightAdd(str,"3,5,2",pub)
	if res != "16f6c4a45ce250d511bae176f84172bfb4dbbec82f39e9d732ba6249e744eca99194ad1bd402c7f5d8c7308762819baf51943052237ea5931fea5010c7a5d386" || err != nil {
		t.Error("wrong res")
	}
}
func TestPaillierMul(t *testing.T) {
	address := []byte("0x80989fb9a8eb623dad541c1525828484e5fab75a")
	e := NewPLWrapper(nil)
	str := "3332efac4d4e0dae3f8f6aa2adac08a367fcbb445e23f895c4f1e851a29e449c3a9fecd2aef21d71a4d8fd176b45d283360a921b761370335895eee4de1fb863"
	pub := "95812e8c3bafc54adae58e01e2777a3de717312e2efe8b764c3a684a46106db7"
	e.base.caller = common.BytesToAddress(address)

	res ,err := e.paillierMul(str,"3",pub)
	if res != "29429a0b4935e75fe3a0e1ed8c71e39741a51651231f543143fc44810c033dcd4b3fad4c8d33ee640bf36d8945e1f10bf986e7ac40b14f7f58006e8cbe5f440" || err != nil {
		t.Error("wrong res")
	}
}
