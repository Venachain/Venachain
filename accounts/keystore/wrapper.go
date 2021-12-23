package keystore

import "github.com/Venachain/Venachain/common"

func KeyFileName(keyAddr common.Address) string {
	return keyFileName(keyAddr)
}
