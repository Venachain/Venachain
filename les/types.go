package les

import (
	"github.com/Venachain/Venachain/light"
)

type ProofsResponseData struct {
	Nodes         light.NodeList
	StorageValues [][]byte
}
