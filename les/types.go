package les

import (
	"github.com/PlatONEnetwork/PlatONE-Go/light"
)

type ProofsResponseData struct {
	Nodes         light.NodeList
	StorageValues [][]byte
}
