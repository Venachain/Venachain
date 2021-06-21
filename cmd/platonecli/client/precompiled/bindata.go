// Code generated for package precompile by go-bindata DO NOT EDIT. (@generated)
// sources:
// ../../release/linux/conf/contracts/cnsInitRegEvent.json
// ../../release/linux/conf/contracts/cnsInvokeEvent.json
// ../../release/linux/conf/contracts/cnsManager.cpp.abi.json
// ../../release/linux/conf/contracts/contractdata.cpp.abi.json
// ../../release/linux/conf/contracts/fireWall.abi.json
// ../../release/linux/conf/contracts/groupManager.cpp.abi.json
// ../../release/linux/conf/contracts/nodeManager.cpp.abi.json
// ../../release/linux/conf/contracts/paramManager.cpp.abi.json
// ../../release/linux/conf/contracts/permissionDeniedEvent.json
// ../../release/linux/conf/contracts/userManager.cpp.abi.json
package precompile

import (
	"bytes"
	"compress/gzip"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"
	"time"
)

func bindataRead(data []byte, name string) ([]byte, error) {
	gz, err := gzip.NewReader(bytes.NewBuffer(data))
	if err != nil {
		return nil, fmt.Errorf("Read %q: %v", name, err)
	}

	var buf bytes.Buffer
	_, err = io.Copy(&buf, gz)
	clErr := gz.Close()

	if err != nil {
		return nil, fmt.Errorf("Read %q: %v", name, err)
	}
	if clErr != nil {
		return nil, err
	}

	return buf.Bytes(), nil
}

type asset struct {
	bytes []byte
	info  os.FileInfo
}

type bindataFileInfo struct {
	name    string
	size    int64
	mode    os.FileMode
	modTime time.Time
}

// Name return file name
func (fi bindataFileInfo) Name() string {
	return fi.name
}

// Size return file size
func (fi bindataFileInfo) Size() int64 {
	return fi.size
}

// Mode return file mode
func (fi bindataFileInfo) Mode() os.FileMode {
	return fi.mode
}

// Mode return file modify time
func (fi bindataFileInfo) ModTime() time.Time {
	return fi.modTime
}

// IsDir return file whether a directory
func (fi bindataFileInfo) IsDir() bool {
	return fi.mode&os.ModeDir != 0
}

// Sys return file is sys mode
func (fi bindataFileInfo) Sys() interface{} {
	return nil
}

var _ReleaseLinuxConfContractsCnsinitregeventJson = []byte("\x1f\x8b\x08\x00\x00\x00\x00\x00\x00\xff\x8a\xe6\x52\x50\x50\x50\xa8\x06\x93\x20\xa0\x94\x97\x98\x9b\xaa\x64\xa5\xa0\x14\xed\xec\x17\x1c\xab\xe0\x97\x5f\x92\x99\x56\xa9\xa4\x83\x90\xcf\xcc\x2b\x28\x2d\x29\x56\xb2\x52\x88\x86\x8b\xa1\x9a\x00\x57\x59\x52\x59\x00\x36\xa9\x34\x33\xaf\xc4\xcc\x44\x09\x45\x41\xad\x0e\xb1\xba\x8b\x4b\x8a\x32\xf3\xd2\xd1\x74\xc3\x79\xb1\x48\x2e\x83\xe9\x48\x2d\x4b\xcd\x2b\x81\x68\xa8\xe5\x8a\xe5\x02\x04\x00\x00\xff\xff\x62\x56\xb5\x6c\xe2\x00\x00\x00")

func ReleaseLinuxConfContractsCnsinitregeventJsonBytes() ([]byte, error) {
	return bindataRead(
		_ReleaseLinuxConfContractsCnsinitregeventJson,
		"../../release/linux/conf/contracts/cnsInitRegEvent.json",
	)
}

func ReleaseLinuxConfContractsCnsinitregeventJson() (*asset, error) {
	bytes, err := ReleaseLinuxConfContractsCnsinitregeventJsonBytes()
	if err != nil {
		return nil, err
	}

	info := bindataFileInfo{name: "../../release/linux/conf/contracts/cnsInitRegEvent.json", size: 226, mode: os.FileMode(420), modTime: time.Unix(1624259103, 0)}
	a := &asset{bytes: bytes, info: info}
	return a, nil
}

var _ReleaseLinuxConfContractsCnsinvokeeventJson = []byte("\x1f\x8b\x08\x00\x00\x00\x00\x00\x00\xff\x8a\xe6\x52\x50\x50\x50\xa8\x06\x93\x20\xa0\x94\x97\x98\x9b\xaa\x64\xa5\xa0\xe4\x9c\x57\xec\x99\x57\x96\x9f\x9d\xaa\xa4\x83\x90\xcc\xcc\x2b\x28\x2d\x29\x56\xb2\x52\x88\x86\x8b\xa1\x6a\x87\xab\x2c\xa9\x2c\x00\x1b\x53\x9a\x99\x57\x62\x66\xa2\x84\xa2\xa0\x56\x87\x58\xdd\xc5\x25\x45\x99\x79\xe9\x68\xba\xe1\xbc\x58\x24\x97\xc1\x74\xa4\x96\xa5\xe6\x95\x40\x34\xd4\x72\xc5\x72\x01\x02\x00\x00\xff\xff\xb1\xe8\x47\x80\xdf\x00\x00\x00")

func ReleaseLinuxConfContractsCnsinvokeeventJsonBytes() ([]byte, error) {
	return bindataRead(
		_ReleaseLinuxConfContractsCnsinvokeeventJson,
		"../../release/linux/conf/contracts/cnsInvokeEvent.json",
	)
}

func ReleaseLinuxConfContractsCnsinvokeeventJson() (*asset, error) {
	bytes, err := ReleaseLinuxConfContractsCnsinvokeeventJsonBytes()
	if err != nil {
		return nil, err
	}

	info := bindataFileInfo{name: "../../release/linux/conf/contracts/cnsInvokeEvent.json", size: 223, mode: os.FileMode(420), modTime: time.Unix(1624259103, 0)}
	a := &asset{bytes: bytes, info: info}
	return a, nil
}

var _ReleaseLinuxConfContractsCnsmanagerCppAbiJson = []byte("\x1f\x8b\x08\x00\x00\x00\x00\x00\x00\xff\xec\x56\xb1\x6a\xc3\x30\x10\xdd\xfd\x15\xc2\x73\xa6\xb6\x74\xc8\xd6\xa4\x14\x3a\xd4\x81\x66\x0c\x19\x84\x7d\x16\x07\xf1\xc9\x48\xa7\x80\x5b\xfa\xef\xc5\xa1\x4e\xec\x04\x82\x5d\xb9\x6d\x2a\x92\x21\xc6\x96\x74\x8f\xf7\xde\xe9\x49\xab\x48\x08\x21\xde\x77\xff\xf5\x2f\x26\x59\x40\x3c\x15\x71\x4a\xf6\x15\x14\x5a\x06\xf3\x64\x74\xf1\x4c\xc8\xf1\xe4\x30\x0d\xa9\x74\x6c\xe3\xa9\x58\xed\xbf\x75\x0b\x9d\x14\xdc\x3d\x27\xa7\xe3\x5c\x95\xbb\x71\xcb\x06\x49\xc5\x9d\x09\x1f\x93\xbe\xd5\xb7\x60\x2c\x6a\x1a\x0c\xb0\x7f\x5b\xb7\xc8\x69\xc7\x43\xd9\x9d\x03\x46\xe2\xdb\x9b\x3e\xb8\xa9\x26\xcb\x92\xb8\x5e\x94\xcb\x8d\x6d\xeb\xb5\xaf\x96\x3b\x4a\xb9\xa6\x1a\xb5\x14\x3a\x6f\x60\x88\xc6\xf5\x06\x90\x59\x66\xc0\xda\x6b\x67\x74\x3b\x23\x43\x03\xe9\x75\x4b\xff\x2b\xe3\x14\xf0\x5c\x13\x1b\x99\xf2\xc3\x49\x57\x5f\xfd\x3b\xc2\x1f\x07\xb8\x6d\x20\x1b\xe7\xeb\x5f\x13\xc9\x90\x35\x4e\x7a\x7a\x58\x4a\x05\x89\x2b\x86\x36\x6b\x6f\x17\xeb\xfa\x4b\x7c\x3b\xdb\x27\xbd\x77\x43\xc0\x2e\xce\xaa\x85\x41\x85\xe4\xe7\xa6\x3e\xae\xe1\xc1\x39\x08\xb1\x31\x3f\x68\x3d\xab\x92\x6e\x60\xfd\x56\xe4\x5d\xec\x89\x32\xae\xba\xa3\x9c\x29\xc1\x5d\xb7\x7e\x2a\x2e\x82\x57\xfb\xb2\xd2\x39\xe0\xe8\xf8\xa3\x64\x2e\x4a\x6d\x78\xb1\xc9\xe6\x64\x5f\x24\x49\x05\xe6\x51\xb2\xf4\xd3\x78\x7c\x9a\xde\x97\xee\xd5\x3c\x59\xae\x45\xa2\x19\xf3\xea\x7b\xe4\x1a\x44\x87\xc4\xf7\x77\x43\x2f\x62\xc3\xe9\x37\x2b\x60\x0b\xc4\x5f\xf4\xa2\x75\xf4\x19\x00\x00\xff\xff\x9e\x68\xc1\x12\xeb\x11\x00\x00")

func ReleaseLinuxConfContractsCnsmanagerCppAbiJsonBytes() ([]byte, error) {
	return bindataRead(
		_ReleaseLinuxConfContractsCnsmanagerCppAbiJson,
		"../../release/linux/conf/contracts/cnsManager.cpp.abi.json",
	)
}

func ReleaseLinuxConfContractsCnsmanagerCppAbiJson() (*asset, error) {
	bytes, err := ReleaseLinuxConfContractsCnsmanagerCppAbiJsonBytes()
	if err != nil {
		return nil, err
	}

	info := bindataFileInfo{name: "../../release/linux/conf/contracts/cnsManager.cpp.abi.json", size: 4587, mode: os.FileMode(420), modTime: time.Unix(1624259103, 0)}
	a := &asset{bytes: bytes, info: info}
	return a, nil
}

var _ReleaseLinuxConfContractsContractdataCppAbiJson = []byte("\x1f\x8b\x08\x00\x00\x00\x00\x00\x00\xff\x8a\xe6\x52\x50\x50\x50\xa8\x06\x93\x20\xa0\x94\x97\x98\x9b\xaa\x64\xa5\xa0\x94\x9b\x99\x5e\x94\x58\x92\xaa\xa4\x83\x90\xca\xcc\x2b\x28\x2d\x29\x56\xb2\x52\x88\x86\x8b\xa1\x6a\xc6\x30\xa4\xb8\x28\x19\xc9\x00\xb8\x74\x49\x65\x01\x44\xba\xa4\x28\x33\x2f\x5d\x09\x45\x41\xad\x0e\xb1\x86\xa7\xa4\x16\x97\x90\x6c\x3a\x9c\x17\x8b\xe4\xb1\xfc\xd2\x12\x52\x7d\x86\xcf\xe2\xcc\xbc\x12\x63\x23\x62\xec\x4d\xce\xcf\x2b\x2e\x49\xcc\x2b\x01\x69\x4a\x4b\xcc\x29\x46\x09\x6d\x98\x69\x69\xa5\x79\xc9\x25\x99\xf9\x79\x10\x03\x6b\xb9\x62\x01\x01\x00\x00\xff\xff\xd1\xd0\x6b\xd0\xb3\x01\x00\x00")

func ReleaseLinuxConfContractsContractdataCppAbiJsonBytes() ([]byte, error) {
	return bindataRead(
		_ReleaseLinuxConfContractsContractdataCppAbiJson,
		"../../release/linux/conf/contracts/contractdata.cpp.abi.json",
	)
}

func ReleaseLinuxConfContractsContractdataCppAbiJson() (*asset, error) {
	bytes, err := ReleaseLinuxConfContractsContractdataCppAbiJsonBytes()
	if err != nil {
		return nil, err
	}

	info := bindataFileInfo{name: "../../release/linux/conf/contracts/contractdata.cpp.abi.json", size: 435, mode: os.FileMode(420), modTime: time.Unix(1624259103, 0)}
	a := &asset{bytes: bytes, info: info}
	return a, nil
}

var _ReleaseLinuxConfContractsFirewallAbiJson = []byte("\x1f\x8b\x08\x00\x00\x00\x00\x00\x00\xff\xec\x58\x4f\x6b\xbc\x30\x10\xbd\xfb\x29\x42\xce\x9e\x7e\xbf\xd2\xc3\xde\x4a\xff\x40\x2f\xed\xa1\xc7\x65\x91\xb0\x8e\x25\xe0\x4e\x24\x99\xb4\x95\xb2\xdf\xbd\xec\x2e\x5a\x6d\x97\x56\xb3\x5b\x34\x21\x1e\x04\x35\x79\x8f\xf7\xe6\x0d\x64\x5c\x26\x8c\x31\xf6\xbe\xbf\xef\x2e\x8e\x62\x03\x7c\xc1\x78\x96\x99\xda\x64\x77\xaf\x8f\x15\x20\x4f\x3f\xbf\x4b\xac\x2c\x19\xbe\x60\xcb\xf6\x5d\x1f\xe1\x1b\x92\xc8\x73\x0d\xc6\x74\x40\xda\x25\x54\x57\xfb\x25\x86\xb4\xc4\x67\xde\x5b\xb0\x6d\x9f\x56\x1d\x7a\x65\x69\x2c\xff\x4f\xc4\x12\xe9\xff\xbf\x21\xbc\x6b\x85\x86\x04\xd2\x6e\x53\x21\x4a\x03\x5d\x4f\x1a\xb4\xc2\xe2\x9a\xa4\xc2\x03\xe0\x36\xfd\xcd\xdb\xeb\x52\xf5\x81\xa2\xb9\xe7\x33\xf7\x2a\xcf\xa7\xb1\x36\x1d\x4c\x70\xd0\xf3\x67\xf8\xda\x96\x10\x93\x71\xac\xed\x40\xe8\x30\xb3\x11\x7e\xf1\x6e\xa0\x0c\xb3\x74\xb1\xad\xbf\xa0\x8d\x4d\xc6\x13\x50\x4c\x46\x4c\xc6\xb1\x64\x90\x20\x6b\xc2\x3d\x68\x0d\x27\xee\xda\x4b\xda\x9e\xc5\xdd\xfb\x4d\xa5\xf4\xdc\x5b\x6f\x96\x75\x9b\xb6\x2b\x6e\xdf\xa6\xab\x5b\xb8\x5d\xf1\xa0\x48\x16\xb5\x9b\xab\x0d\x97\x95\x48\x97\x17\x63\xa3\x3e\x5e\x77\xb3\x03\x5e\x00\x69\x68\x6c\xdc\x7f\x08\x78\x21\xef\x84\x99\xdc\x0b\x7d\xce\x63\xb1\x17\xea\x9c\xa7\x03\x2f\xd4\x39\x9f\x70\xbd\x50\x77\xc2\x58\x3e\x07\x7d\xc9\x2a\xf9\x08\x00\x00\xff\xff\x02\x96\xf5\xb6\x4d\x15\x00\x00")

func ReleaseLinuxConfContractsFirewallAbiJsonBytes() ([]byte, error) {
	return bindataRead(
		_ReleaseLinuxConfContractsFirewallAbiJson,
		"../../release/linux/conf/contracts/fireWall.abi.json",
	)
}

func ReleaseLinuxConfContractsFirewallAbiJson() (*asset, error) {
	bytes, err := ReleaseLinuxConfContractsFirewallAbiJsonBytes()
	if err != nil {
		return nil, err
	}

	info := bindataFileInfo{name: "../../release/linux/conf/contracts/fireWall.abi.json", size: 5453, mode: os.FileMode(420), modTime: time.Unix(1624259103, 0)}
	a := &asset{bytes: bytes, info: info}
	return a, nil
}

var _ReleaseLinuxConfContractsGroupmanagerCppAbiJson = []byte("\x1f\x8b\x08\x00\x00\x00\x00\x00\x00\xff\xec\x97\x4f\x4f\x84\x30\x10\xc5\xef\xfb\x29\x9a\x9e\x39\xa9\xf1\xb0\x37\x37\x9b\x98\xf5\x80\x26\x1e\x37\x7b\x68\x60\xc0\x26\x30\x25\x9d\xa9\x09\x31\xfb\xdd\x4d\x31\x20\x1b\x75\xff\x75\x15\xd4\xe5\x40\x02\xb4\xaf\xfd\xbd\xd7\xa1\xb0\x9c\x08\x21\xc4\x4b\x73\xf6\x87\x44\x55\x82\x9c\x0a\x99\x58\x50\x0c\xb7\xd6\xb8\x4a\x46\xef\x8f\x35\x56\x8e\x49\x4e\xc5\xb2\xbb\xb7\x29\xf0\x41\x28\xf7\x12\x0b\xcc\x4c\x4f\xa6\x6b\xc4\x75\xd5\x34\x22\xb6\x1a\x73\xb9\xd1\x60\xdd\x5d\xad\x7a\x13\x30\x8e\x0f\x9d\xc1\xb6\x81\x35\xf2\xe5\xc5\x3e\xe3\x26\x06\x89\x15\xb2\xef\x94\xa9\x82\xa0\xef\x4a\xab\x96\x39\x4c\x58\x1b\x7c\x13\x5c\x47\x5f\x99\x9b\x03\xdf\x14\x45\x63\x2e\x7d\xee\xee\x37\x22\xef\xef\x75\x9f\x99\xad\x0b\x42\x7e\x52\xd4\xf0\xde\x57\x0f\x60\x4b\x4d\xe4\xfb\xfc\x34\xfa\x51\x69\x87\x92\xa3\x49\x81\x62\x57\x86\x95\x91\x57\xb9\x23\x83\x8f\x6c\xff\x4e\x21\x85\x5a\x9b\x03\x37\x8b\x6a\x56\x2f\xe6\xa7\x78\x4b\xcd\xb7\x11\x3a\x8d\x7c\x7d\x35\x80\xb5\xc3\x14\xac\xab\x52\xc5\x30\x33\x86\x63\xbf\x80\x87\xb1\x37\xfa\xd7\xe5\x11\xbc\xcf\xa8\x34\x6d\x03\x3c\xe7\xb7\x73\x0e\xe3\xcb\x2f\x85\xe2\x9c\xdf\x2f\xce\x2f\x36\xac\xb3\xfa\xb8\xe8\xc2\x72\x39\xdc\xf0\xb6\x07\x3c\x03\xf2\x2e\xb0\xe0\xbf\x83\x51\xd3\x05\x97\xdd\xa8\xe9\x82\x37\x85\x51\xd3\x9d\xe4\xb3\x65\x0c\x84\x13\xb1\x7a\x0d\x00\x00\xff\xff\xdf\x65\x30\x4d\x9d\x0f\x00\x00")

func ReleaseLinuxConfContractsGroupmanagerCppAbiJsonBytes() ([]byte, error) {
	return bindataRead(
		_ReleaseLinuxConfContractsGroupmanagerCppAbiJson,
		"../../release/linux/conf/contracts/groupManager.cpp.abi.json",
	)
}

func ReleaseLinuxConfContractsGroupmanagerCppAbiJson() (*asset, error) {
	bytes, err := ReleaseLinuxConfContractsGroupmanagerCppAbiJsonBytes()
	if err != nil {
		return nil, err
	}

	info := bindataFileInfo{name: "../../release/linux/conf/contracts/groupManager.cpp.abi.json", size: 3997, mode: os.FileMode(420), modTime: time.Unix(1624259103, 0)}
	a := &asset{bytes: bytes, info: info}
	return a, nil
}

var _ReleaseLinuxConfContractsNodemanagerCppAbiJson = []byte("\x1f\x8b\x08\x00\x00\x00\x00\x00\x00\xff\xec\x97\x41\x6b\x83\x30\x14\xc7\xef\xfd\x14\x21\x67\x4f\xdb\xd8\xa1\xb7\xb1\xee\xd2\x81\x3b\x0c\x76\x29\x3d\x64\xe6\x29\x81\xf8\x22\xc9\x4b\x41\x46\xbf\xfb\xd0\xa1\x53\xb6\x75\xb5\xe9\xa8\xe8\x3c\x28\x86\xbc\xf7\xf2\xfb\xcb\xff\x3d\xdc\x2c\x18\x63\xec\xad\xbe\x57\x17\x47\x91\x03\x5f\x32\x2e\xa4\xe4\xd1\xe7\xb2\xc2\xc2\x93\xe3\x4b\xb6\x69\xd7\xfa\x81\x5f\x12\xa0\x91\xb0\x76\x06\x9f\xc9\x76\x12\xb5\xdb\xa8\x2c\xea\x6d\x8e\xac\xc2\x8c\xf7\x36\xec\xdb\xb7\x6d\xe7\x08\xc6\xd3\xd0\x33\x1c\x2a\xac\x90\xae\xaf\x8e\xa9\x9b\x18\x74\x24\x90\xaa\xa0\x54\x68\x07\x5d\x5d\x9a\x6c\xa9\xc7\x84\x94\xc1\x8f\x84\xfb\xe8\x27\x59\x33\xa0\x3b\xad\x63\x23\xc1\x7d\x2f\xef\x1f\x12\x1f\x2f\x75\x17\x99\xac\x0f\x25\x7e\xb1\xe9\xbd\x41\x07\xe8\xbc\x9b\x13\xfa\x4e\x68\x25\xd7\x46\x61\x05\x1d\xe6\xa6\xc2\xbf\x6a\x95\x3c\x42\x39\x1d\x2f\x85\xaa\x5b\x35\x18\x17\xfb\xfc\xbf\x4d\x9d\x5d\xda\x0c\xe8\x80\x4f\xa7\x22\xed\x65\x9a\x82\x2f\xa4\xa0\xc0\x6e\x50\x3f\x87\x92\x45\x53\xf9\x6e\x17\x9b\xdc\xb1\xb1\xb9\xd0\x0f\x95\x3e\x73\x1a\x63\x19\xd0\x0a\x34\x10\xc8\xd9\xa1\xab\xbc\x30\x96\x9e\xb4\xac\xa1\x57\x82\x44\x98\x71\x65\x3f\x43\x00\xe5\x24\x3c\x15\x1b\x52\x69\x79\x9a\xa4\x4d\x2d\xaf\x90\x6e\x6f\x86\xf6\xba\xe1\x82\x37\x11\xb0\x03\xa4\xdf\xc0\x4e\xfe\x7b\x1a\x35\x55\xc8\xe8\x1a\x35\xd8\xb9\x6c\x3e\x06\xc8\xc5\xf6\x3d\x00\x00\xff\xff\xc5\x7f\xc4\x82\xd4\x0f\x00\x00")

func ReleaseLinuxConfContractsNodemanagerCppAbiJsonBytes() ([]byte, error) {
	return bindataRead(
		_ReleaseLinuxConfContractsNodemanagerCppAbiJson,
		"../../release/linux/conf/contracts/nodeManager.cpp.abi.json",
	)
}

func ReleaseLinuxConfContractsNodemanagerCppAbiJson() (*asset, error) {
	bytes, err := ReleaseLinuxConfContractsNodemanagerCppAbiJsonBytes()
	if err != nil {
		return nil, err
	}

	info := bindataFileInfo{name: "../../release/linux/conf/contracts/nodeManager.cpp.abi.json", size: 4052, mode: os.FileMode(420), modTime: time.Unix(1624259103, 0)}
	a := &asset{bytes: bytes, info: info}
	return a, nil
}

var _ReleaseLinuxConfContractsParammanagerCppAbiJson = []byte("\x1f\x8b\x08\x00\x00\x00\x00\x00\x00\xff\xdc\x58\x41\x6f\x9b\x30\x14\xbe\xf3\x2b\x10\x67\x4e\xdb\xb4\x43\x6f\xac\xdb\xb2\x49\x53\x15\x6d\xe9\x2e\x55\x0f\x9e\xeb\xa4\x56\xc1\x46\xf6\xa3\x2b\x9a\xf2\xdf\x27\x20\x64\xc9\x6a\xc0\x60\xc7\x46\xe4\x10\x29\xc4\x7e\xfe\xbe\xef\x7d\xf8\x3d\xfb\x2e\x08\xc3\x30\xfc\x53\x7f\x57\x9f\x88\xa1\x8c\x44\x57\x61\x24\x09\xac\x90\xbc\xe6\x0c\x04\xc2\x70\x53\x3d\x8d\xff\x8d\xa2\x2c\x2f\x40\x46\x57\xe1\xdd\xf1\xd9\x79\x9c\x57\xf1\xb0\x3a\xd2\x71\x1c\x94\x79\xb3\x2e\x08\xca\x76\xd1\xd9\x80\xfd\xf1\xd7\xfd\x09\x06\x5e\x40\x0b\xe2\xf4\x31\xe6\x4c\x02\x62\x50\x05\xdb\xa2\x54\x9e\xe1\x6e\x57\xd9\x16\x0c\x03\xe5\xac\x59\x67\x1f\x77\xa9\xb0\xd3\x54\xa1\x03\x97\xae\x38\x76\x04\x39\x65\x0e\xa2\x30\x22\x2e\x09\x7c\x95\x6b\xc1\x1f\x0a\x4c\x3e\x65\x39\x94\x1f\x52\x8e\x9f\xcc\x2c\x40\x7b\x03\xbe\x82\x5a\x50\x06\x6f\xdf\xcc\xc8\x09\xda\x82\x5c\xd0\x0d\xfa\xa2\x58\x76\xc3\xe6\x65\x85\xe4\x37\x9a\x51\x30\x33\x01\xa8\xe2\x28\x69\xbe\x7f\x37\xa3\xdc\x0f\xd1\xbf\x70\xca\xf5\xb4\xb0\x9c\xf2\xda\xe1\x76\xb2\xfe\xab\x23\x94\x01\x59\x47\x89\xd7\x10\x61\x89\xb9\x4f\xd2\x94\xff\x4e\x58\x99\x60\xcc\x0b\x06\x1f\x49\x9e\xf2\xb2\xad\x83\xa6\x65\x40\x37\xb8\x52\x8e\xb9\x78\x43\x12\xb8\x7e\x24\xf8\xa9\xc5\xdd\xb0\x58\x13\x91\x51\x29\xab\xf9\x66\xed\x52\x15\x5a\x19\x4c\x29\xca\xac\xea\xe4\x04\x59\x16\x57\x31\x77\x93\x5e\xa1\xc5\xc9\x50\xb7\x91\x49\x9e\x0b\xfe\x4c\x1a\x01\xc8\x83\xb5\x5d\x64\x30\xac\x81\x08\xce\x5a\xca\x51\xe2\x2c\xd3\x1f\x9b\x97\x5b\x49\x56\x48\x9a\x1a\x42\x11\xc7\x80\xa6\x33\x07\xf4\xd3\x5f\x66\xca\xeb\xae\xea\x56\x92\x8d\xa0\xe4\x0b\x92\x8f\xa6\xa9\xef\x89\x67\x40\xdb\x99\x05\xf4\xe4\x58\x98\x15\x22\x8d\x9b\x95\xff\xc8\x1d\x56\x68\xe1\xee\x63\xf5\xdf\x87\x2b\x13\x35\x81\xc3\x18\xf2\x4c\xda\x5e\xb2\x13\xa0\xe6\x81\xdf\x2f\xc8\xfe\x93\xa9\x5f\x6c\x83\x87\x27\xdf\xf9\x1d\xdd\xaa\xfa\x06\x3c\xa2\x5d\xf0\x0d\xb5\xaf\xae\xb9\xc1\x16\x76\xee\xba\x37\x1c\xe8\xb6\x9c\x56\x75\x7a\xf7\xcc\x58\x77\xb6\xfe\xb5\x6e\x3b\x43\x43\xf5\xba\x9c\xfc\xfc\xfe\x79\x8d\x04\xca\xdc\x37\x14\xde\xae\xaa\x07\x28\xbb\xe5\x76\xf9\xae\x21\xea\xe5\xeb\xfb\xbd\xd7\x69\x66\xbc\x62\x6c\x9a\x2e\x06\xb5\x84\xa3\x1d\xd3\xe5\x16\x33\xa7\xd8\x6a\xe2\xbc\x5c\x16\x4e\xd6\x52\x9b\x9a\x52\xd3\xc1\xbd\xd6\xae\x68\x8e\x4e\x03\x3f\x40\xf8\x10\xd3\x85\x41\xbd\xd5\x07\x4f\x9a\x5a\x31\xa8\x97\xc2\x13\xdc\x07\xc1\xdf\x00\x00\x00\xff\xff\x11\xd4\xfa\x25\x1b\x20\x00\x00")

func ReleaseLinuxConfContractsParammanagerCppAbiJsonBytes() ([]byte, error) {
	return bindataRead(
		_ReleaseLinuxConfContractsParammanagerCppAbiJson,
		"../../release/linux/conf/contracts/paramManager.cpp.abi.json",
	)
}

func ReleaseLinuxConfContractsParammanagerCppAbiJson() (*asset, error) {
	bytes, err := ReleaseLinuxConfContractsParammanagerCppAbiJsonBytes()
	if err != nil {
		return nil, err
	}

	info := bindataFileInfo{name: "../../release/linux/conf/contracts/paramManager.cpp.abi.json", size: 8219, mode: os.FileMode(420), modTime: time.Unix(1624259103, 0)}
	a := &asset{bytes: bytes, info: info}
	return a, nil
}

var _ReleaseLinuxConfContractsPermissiondeniedeventJson = []byte("\x1f\x8b\x08\x00\x00\x00\x00\x00\x00\xff\x8a\xe6\x52\x50\x50\x50\xa8\x06\x93\x20\xa0\x94\x97\x98\x9b\xaa\x64\xa5\xa0\x94\x9c\x9f\x57\x52\x94\x98\x5c\xa2\x50\x90\x5a\x94\x9b\x59\x5c\x9c\x99\x9f\xa7\xa4\x83\x50\x96\x99\x57\x50\x5a\x52\xac\x64\xa5\x10\x0d\x17\x43\x35\x08\xae\xb2\xa4\xb2\x00\x6c\x60\x71\x49\x51\x66\x5e\xba\x12\x8a\x82\x5a\x38\x2f\x16\xc9\x6c\x98\x8e\xd4\xb2\xd4\xbc\x12\x88\x86\x5a\xae\x58\x2e\x40\x00\x00\x00\xff\xff\xde\xc4\x5a\x41\xab\x00\x00\x00")

func ReleaseLinuxConfContractsPermissiondeniedeventJsonBytes() ([]byte, error) {
	return bindataRead(
		_ReleaseLinuxConfContractsPermissiondeniedeventJson,
		"../../release/linux/conf/contracts/permissionDeniedEvent.json",
	)
}

func ReleaseLinuxConfContractsPermissiondeniedeventJson() (*asset, error) {
	bytes, err := ReleaseLinuxConfContractsPermissiondeniedeventJsonBytes()
	if err != nil {
		return nil, err
	}

	info := bindataFileInfo{name: "../../release/linux/conf/contracts/permissionDeniedEvent.json", size: 171, mode: os.FileMode(420), modTime: time.Unix(1624259103, 0)}
	a := &asset{bytes: bytes, info: info}
	return a, nil
}

var _ReleaseLinuxConfContractsUsermanagerCppAbiJson = []byte("\x1f\x8b\x08\x00\x00\x00\x00\x00\x00\xff\xec\x9a\xcf\x6e\x9c\x30\x10\xc6\xef\x79\x0a\x8b\x33\xa7\xb6\xea\x21\xb7\x34\x91\xaa\x4a\xd5\x56\x6a\xd5\x53\x94\x83\xb5\x1e\x36\x48\xc4\x46\xf6\x50\x69\x55\xe5\xdd\x2b\x36\xbb\x2c\x7f\x6c\x16\x62\xd8\x35\xdb\xc9\x21\x07\x43\xcc\x37\x3f\x0f\x9f\xc9\x78\x1e\x6f\x18\x63\xec\xef\xee\x77\xf9\x13\x49\xfe\x02\xd1\x2d\x8b\x0c\xe0\xaf\x22\x07\x7d\x27\x5e\x52\x19\xc5\xc7\x1b\x52\x99\x17\x68\xa2\x5b\xf6\xf8\x54\x1b\x55\x05\xda\x86\xd7\x4a\x1a\xe4\x12\xcb\x19\x13\x9e\x19\xa8\xcf\x84\xdb\x7c\xf7\xa8\xa4\x90\x6b\x4c\x95\x8c\x76\x97\x5e\x63\x97\x24\xd4\x5c\x9a\x04\xf4\x51\xd7\x97\xed\x9d\x10\x1a\x8c\xb1\x0b\xac\xc6\x9a\x13\x76\x26\xe6\x42\xe8\xda\x0c\x1d\x81\x06\x75\x2a\x37\x51\x75\xa1\x71\xe7\x6b\x7d\xe0\xdc\x48\x36\x80\x25\x81\xef\xa9\xc1\x1f\xc9\x4f\x95\x81\x1f\x08\xe4\x7a\x03\xd8\x9a\xc7\x85\xa3\x49\xe1\x14\x82\xa1\x12\xa6\x79\x70\x1d\x32\xea\xc2\x97\x71\x49\xc4\x9c\x37\xd9\xae\x8d\xee\xfe\x25\xe9\xe5\xbb\x2a\x47\xbc\xe0\xca\xe6\x0c\xff\x0d\x5c\x17\x5b\x2e\xc4\xfd\x33\x4f\xe5\x05\xdc\x72\x29\x8c\x07\x7b\xf0\x50\xc8\xfe\x59\x4c\x84\x5b\x9a\x04\x64\x94\xc6\x53\x41\x1e\x4a\x99\xdc\x78\x0e\xa7\xf8\xaa\x55\x91\x13\xe1\x33\x11\x26\xa7\xf0\x81\xdc\xe7\x14\x44\x79\x22\xca\x03\x19\x93\x57\x4c\x0d\x98\x0b\xb1\x52\x02\x88\xef\x8c\x5e\x5c\x03\x4c\x26\xe1\xc3\xb8\xcf\x8a\x09\xf2\x34\x90\x87\x21\x26\xa3\x98\xc1\x88\xef\x95\x44\xcd\xd7\x48\x69\x3c\xa7\x57\x74\x41\x53\x32\x7b\x52\xee\xf1\x0c\x42\x3d\x19\xea\x31\x9c\xc9\x3b\x66\x42\x5d\x73\x8f\x07\xc8\x33\xb5\x05\x4d\xb4\xfd\x68\x8f\x42\x4d\x75\xe5\x79\x4a\x9e\x04\x7a\x3a\xd0\xa3\x38\x93\x77\x78\xa0\xee\xb1\x8e\xdf\x06\xb4\x1f\xd4\xc2\x80\xfe\x26\x13\x45\x60\x6b\x4c\x72\xc1\x11\x4a\xb6\x0f\x60\xd6\x2d\x3a\xef\x4c\xdc\x66\xf6\x0f\x8b\x34\x1e\xfa\x00\xd1\x95\xe9\xc1\xf2\x3a\x16\x71\x03\x78\x97\x65\xe5\x22\x3a\x6c\x27\xb8\x90\x27\xe8\x5e\x29\xc3\x3d\xe1\xb6\xee\xf0\x4e\xe4\xab\x33\xba\xd7\xeb\x25\xe9\xfe\x3e\x20\x8c\xa7\x31\x3e\x73\xe3\x6e\x54\x9b\x1e\x20\x1b\xe1\x99\x91\x56\x19\xac\x7a\xff\xcb\xbe\x9c\x63\xa6\x12\x3f\x7e\x38\xef\x62\xb5\x9b\x40\x5b\xd2\xba\x0b\xd8\x0a\xf4\xc0\xac\x78\xd3\xde\x5e\x86\x16\x52\x7b\x34\xfb\x7b\xe0\x0f\x48\x3c\xa1\xb6\xaf\x3f\x94\xb1\x78\x11\x7a\x77\xd9\x57\xde\x14\x87\xc7\xd7\xd1\x4c\xc6\xe2\x40\xf3\xc1\xd6\x97\xc5\xd8\x31\x8b\x03\xd4\x6b\x3b\x55\xaf\x24\x07\xae\xf7\x90\xba\x71\xc0\x7c\x2d\x47\x65\x4b\xd1\x5b\x39\x43\xd0\xef\x9b\xbd\x16\xfb\xa6\x39\x7c\xbd\x6f\x8c\x17\xc0\xb7\x5b\x44\x09\x70\x7f\x73\x57\x31\x59\x1c\x22\x5f\x47\x97\x69\xf5\xc6\x05\xae\xb7\xfe\xe9\x10\xaa\x5e\xfb\xfe\x16\x2e\x5f\xdb\xfe\x16\x32\xdf\x45\xed\x6f\x96\xbe\x0a\x16\x38\x5f\xe7\xfe\x16\xe2\xf7\xba\xe3\x0c\xba\x42\x1c\xb0\xde\x45\xec\x6f\xee\xc3\xa3\x30\xf3\xe1\x70\x34\xc0\x02\xfa\x48\x6f\x28\xd8\xeb\xb4\x55\xda\x2f\xad\xb3\x2d\x92\x45\x2b\x85\x69\xb2\x7d\xdf\x19\xc0\xa1\x0c\x53\x6a\xfd\xfc\x69\x6c\xb1\x6a\x7c\xfd\xee\xf0\x17\xf5\xc0\x6e\x9e\xfe\x05\x00\x00\xff\xff\x02\x83\x21\x9b\xfc\x3b\x00\x00")

func ReleaseLinuxConfContractsUsermanagerCppAbiJsonBytes() ([]byte, error) {
	return bindataRead(
		_ReleaseLinuxConfContractsUsermanagerCppAbiJson,
		"../../release/linux/conf/contracts/userManager.cpp.abi.json",
	)
}

func ReleaseLinuxConfContractsUsermanagerCppAbiJson() (*asset, error) {
	bytes, err := ReleaseLinuxConfContractsUsermanagerCppAbiJsonBytes()
	if err != nil {
		return nil, err
	}

	info := bindataFileInfo{name: "../../release/linux/conf/contracts/userManager.cpp.abi.json", size: 15356, mode: os.FileMode(420), modTime: time.Unix(1624259103, 0)}
	a := &asset{bytes: bytes, info: info}
	return a, nil
}

// Asset loads and returns the asset for the given name.
// It returns an error if the asset could not be found or
// could not be loaded.
func Asset(name string) ([]byte, error) {
	cannonicalName := strings.Replace(name, "\\", "/", -1)
	if f, ok := _bindata[cannonicalName]; ok {
		a, err := f()
		if err != nil {
			return nil, fmt.Errorf("Asset %s can't read by error: %v", name, err)
		}
		return a.bytes, nil
	}
	return nil, fmt.Errorf("Asset %s not found", name)
}

// MustAsset is like Asset but panics when Asset would return an error.
// It simplifies safe initialization of global variables.
func MustAsset(name string) []byte {
	a, err := Asset(name)
	if err != nil {
		panic("asset: Asset(" + name + "): " + err.Error())
	}

	return a
}

// AssetInfo loads and returns the asset info for the given name.
// It returns an error if the asset could not be found or
// could not be loaded.
func AssetInfo(name string) (os.FileInfo, error) {
	cannonicalName := strings.Replace(name, "\\", "/", -1)
	if f, ok := _bindata[cannonicalName]; ok {
		a, err := f()
		if err != nil {
			return nil, fmt.Errorf("AssetInfo %s can't read by error: %v", name, err)
		}
		return a.info, nil
	}
	return nil, fmt.Errorf("AssetInfo %s not found", name)
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
var _bindata = map[string]func() (*asset, error){
	"../../release/linux/conf/contracts/cnsInitRegEvent.json":       ReleaseLinuxConfContractsCnsinitregeventJson,
	"../../release/linux/conf/contracts/cnsInvokeEvent.json":        ReleaseLinuxConfContractsCnsinvokeeventJson,
	"../../release/linux/conf/contracts/cnsManager.cpp.abi.json":    ReleaseLinuxConfContractsCnsmanagerCppAbiJson,
	"../../release/linux/conf/contracts/contractdata.cpp.abi.json":  ReleaseLinuxConfContractsContractdataCppAbiJson,
	"../../release/linux/conf/contracts/fireWall.abi.json":          ReleaseLinuxConfContractsFirewallAbiJson,
	"../../release/linux/conf/contracts/groupManager.cpp.abi.json":  ReleaseLinuxConfContractsGroupmanagerCppAbiJson,
	"../../release/linux/conf/contracts/nodeManager.cpp.abi.json":   ReleaseLinuxConfContractsNodemanagerCppAbiJson,
	"../../release/linux/conf/contracts/paramManager.cpp.abi.json":  ReleaseLinuxConfContractsParammanagerCppAbiJson,
	"../../release/linux/conf/contracts/permissionDeniedEvent.json": ReleaseLinuxConfContractsPermissiondeniedeventJson,
	"../../release/linux/conf/contracts/userManager.cpp.abi.json":   ReleaseLinuxConfContractsUsermanagerCppAbiJson,
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
	for childName := range node.Children {
		rv = append(rv, childName)
	}
	return rv, nil
}

type bintree struct {
	Func     func() (*asset, error)
	Children map[string]*bintree
}

var _bintree = &bintree{nil, map[string]*bintree{
	"..": &bintree{nil, map[string]*bintree{
		"..": &bintree{nil, map[string]*bintree{
			"release": &bintree{nil, map[string]*bintree{
				"linux": &bintree{nil, map[string]*bintree{
					"conf": &bintree{nil, map[string]*bintree{
						"contracts": &bintree{nil, map[string]*bintree{
							"cnsInitRegEvent.json":       &bintree{ReleaseLinuxConfContractsCnsinitregeventJson, map[string]*bintree{}},
							"cnsInvokeEvent.json":        &bintree{ReleaseLinuxConfContractsCnsinvokeeventJson, map[string]*bintree{}},
							"cnsManager.cpp.abi.json":    &bintree{ReleaseLinuxConfContractsCnsmanagerCppAbiJson, map[string]*bintree{}},
							"contractdata.cpp.abi.json":  &bintree{ReleaseLinuxConfContractsContractdataCppAbiJson, map[string]*bintree{}},
							"fireWall.abi.json":          &bintree{ReleaseLinuxConfContractsFirewallAbiJson, map[string]*bintree{}},
							"groupManager.cpp.abi.json":  &bintree{ReleaseLinuxConfContractsGroupmanagerCppAbiJson, map[string]*bintree{}},
							"nodeManager.cpp.abi.json":   &bintree{ReleaseLinuxConfContractsNodemanagerCppAbiJson, map[string]*bintree{}},
							"paramManager.cpp.abi.json":  &bintree{ReleaseLinuxConfContractsParammanagerCppAbiJson, map[string]*bintree{}},
							"permissionDeniedEvent.json": &bintree{ReleaseLinuxConfContractsPermissiondeniedeventJson, map[string]*bintree{}},
							"userManager.cpp.abi.json":   &bintree{ReleaseLinuxConfContractsUsermanagerCppAbiJson, map[string]*bintree{}},
						}},
					}},
				}},
			}},
		}},
	}},
}}

// RestoreAsset restores an asset under the given directory
func RestoreAsset(dir, name string) error {
	data, err := Asset(name)
	if err != nil {
		return err
	}
	info, err := AssetInfo(name)
	if err != nil {
		return err
	}
	err = os.MkdirAll(_filePath(dir, filepath.Dir(name)), os.FileMode(0755))
	if err != nil {
		return err
	}
	err = ioutil.WriteFile(_filePath(dir, name), data, info.Mode())
	if err != nil {
		return err
	}
	err = os.Chtimes(_filePath(dir, name), info.ModTime(), info.ModTime())
	if err != nil {
		return err
	}
	return nil
}

// RestoreAssets restores an asset under the given directory recursively
func RestoreAssets(dir, name string) error {
	children, err := AssetDir(name)
	// File
	if err != nil {
		return RestoreAsset(dir, name)
	}
	// Dir
	for _, child := range children {
		err = RestoreAssets(dir, filepath.Join(name, child))
		if err != nil {
			return err
		}
	}
	return nil
}

func _filePath(dir, name string) string {
	cannonicalName := strings.Replace(name, "\\", "/", -1)
	return filepath.Join(append([]string{dir}, strings.Split(cannonicalName, "/")...)...)
}
