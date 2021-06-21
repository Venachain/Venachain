package core

import (
	"bytes"
	"compress/gzip"
	"fmt"
	"io"
	"strings"
)

func bindata_read(data []byte, name string) ([]byte, error) {
	gz, err := gzip.NewReader(bytes.NewBuffer(data))
	if err != nil {
		return nil, fmt.Errorf("Read %q: %v", name, err)
	}

	var buf bytes.Buffer
	_, err = io.Copy(&buf, gz)
	gz.Close()

	if err != nil {
		return nil, fmt.Errorf("Read %q: %v", name, err)
	}

	return buf.Bytes(), nil
}

var _privacy_contract_ptoken_sol_ptoken_abi = []byte("\x1f\x8b\x08\x00\x00\x00\x00\x00\x00\xff\xec\x56\x4d\x6b\xdc\x30\x10\xfd\x2f\x73\xd6\x69\x4b\x7b\xf0\x2d\xb4\xd0\x43\x5b\x28\xb4\x3d\x2d\x26\x8c\xed\xf1\xae\x58\xed\x48\x48\xa3\x6e\x4c\xd8\xff\x5e\xbc\xd8\xf1\x47\xbc\x4e\x69\xe3\xba\x84\xdc\x0c\x33\x63\x3d\xbd\xf7\xf4\xa4\xed\x3d\xe4\x96\x83\x20\x0b\x24\xe2\x23\x29\xd0\xec\xa2\x04\x48\xb6\xa9\x02\xc6\x23\x41\x02\x62\x05\xcd\xb7\xe8\x9c\xa9\x40\x81\x8d\xd2\x74\xdc\xb7\x0d\xa0\x40\x2a\x57\x7f\x45\xcd\xb2\x79\xfb\x0e\xce\xa9\x02\x87\x15\x66\x86\x20\x29\xd1\x04\x52\x10\x04\x85\xbe\x44\xc1\x4c\x1b\x2d\x15\x24\xf0\x53\xd3\xa9\x9b\x2d\x23\xe7\xa2\x2d\xc3\x59\xf5\x61\x35\xd3\x0f\xb8\xea\xda\xd1\x59\x26\x1e\x82\xb8\xeb\xfe\x94\x55\x42\xe1\xcd\xe6\xf2\xa3\xa6\x5a\x4d\x54\xbb\x1d\xf6\xaa\x12\x9d\xa1\xfe\x64\xf6\xdd\x23\x87\x92\xfc\xc4\x2e\xd5\x12\x60\xe2\x0c\x18\xe7\xad\x2d\x47\xd3\xfd\xd9\x93\x96\x7d\xe1\xf1\x34\xd0\xe9\xb7\xb4\x60\xcb\x6d\xd3\x53\x8a\x5c\x35\x0a\x39\x9b\xef\x3f\x13\xef\x64\xff\x6a\x94\xf1\x2e\x9b\xae\x32\x72\xb1\x8e\x3a\x06\x83\x7c\x34\x36\x43\xf3\xc3\x15\x28\xb4\x82\x44\x23\x6c\x0f\x8b\x6a\x2e\xe8\x6e\x96\xb6\x1d\xc9\xd7\x98\x19\x9d\x7f\xa2\x71\x08\x3d\xbf\xca\x07\x7a\xac\xf3\x82\x24\x2c\x6f\xd3\x6d\xda\x9f\xbd\x1c\xd3\x59\xb6\x83\x3e\x46\x83\x42\x37\x79\x6e\x63\x8d\x6b\x69\xc6\xb1\x5b\x68\x80\x7b\x93\xd6\xd0\x5f\x4e\x46\x5c\xb9\x44\x5a\xda\x67\x45\xf1\xb4\xd3\x41\x2e\xf1\xb2\x42\x7a\xb8\xc3\xcd\xb1\x96\xe8\x85\x06\xfb\xfb\xe9\x13\xf3\xfc\x0b\x7d\x98\x72\xc7\x3f\x4b\x80\xff\xeb\xb5\x22\xdd\x85\xb9\x82\xa3\xc5\x1e\x88\x9f\xb2\x33\x16\x85\xa7\x10\xfe\xde\xce\x8f\x2f\xbd\xdb\xe1\x73\xe9\x7a\x2c\xdc\xb6\x48\xff\x08\xd3\x14\x55\x17\x76\x7c\xcc\xc5\x7a\x38\xa7\xbf\x02\x00\x00\xff\xff\xb8\xaf\xe4\xa1\x06\x0c\x00\x00")

func privacy_contract_ptoken_sol_ptoken_abi() ([]byte, error) {
	return bindata_read(
		_privacy_contract_ptoken_sol_ptoken_abi,
		"privacy_contract/PToken_sol_PToken.abi",
	)
}

// Asset loads and returns the asset for the given name.
// It returns an error if the asset could not be found or
// could not be loaded.
func Asset(name string) ([]byte, error) {
	cannonicalName := strings.Replace(name, "\\", "/", -1)
	if f, ok := _bindata[cannonicalName]; ok {
		return f()
	}
	return nil, fmt.Errorf("Asset %s not found", name)
}

// AssetNames returns the names of the assets.
func AssetNames() []string {
	names := make([]string, 0, len(_bindata))
	for name := range _bindata {
		names = append(names, name)
	}
	return names
}

// _bindata is a table, holding each asset generator, mapped to its name.
var _bindata = map[string]func() ([]byte, error){
	"privacy_contract/PToken_sol_PToken.abi": privacy_contract_ptoken_sol_ptoken_abi,
}
// AssetDir returns the file names below a certain
// directory embedded in the file by go-bindata.
// For example if you run go-bindata on data/... and data contains the
// following hierarchy:
//     data/
//       foo.txt
//       img/
//         a.png
//         b.png
// then AssetDir("data") would return []string{"foo.txt", "img"}
// AssetDir("data/img") would return []string{"a.png", "b.png"}
// AssetDir("foo.txt") and AssetDir("notexist") would return an error
// AssetDir("") will return []string{"data"}.
func AssetDir(name string) ([]string, error) {
	node := _bintree
	if len(name) != 0 {
		cannonicalName := strings.Replace(name, "\\", "/", -1)
		pathList := strings.Split(cannonicalName, "/")
		for _, p := range pathList {
			node = node.Children[p]
			if node == nil {
				return nil, fmt.Errorf("Asset %s not found", name)
			}
		}
	}
	if node.Func != nil {
		return nil, fmt.Errorf("Asset %s not found", name)
	}
	rv := make([]string, 0, len(node.Children))
	for name := range node.Children {
		rv = append(rv, name)
	}
	return rv, nil
}

type _bintree_t struct {
	Func func() ([]byte, error)
	Children map[string]*_bintree_t
}
var _bintree = &_bintree_t{nil, map[string]*_bintree_t{
	"privacy_contract": &_bintree_t{nil, map[string]*_bintree_t{
		"PToken_sol_PToken.abi": &_bintree_t{privacy_contract_ptoken_sol_ptoken_abi, map[string]*_bintree_t{
		}},
	}},
}}
