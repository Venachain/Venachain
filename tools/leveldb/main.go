/*
Copyright (C) BABEC. All rights reserved.
Copyright (C) THL A29 Limited, a Tencent company. All rights reserved.

SPDX-License-Identifier: Apache-2.0
*/

package main

import (
	"fmt"
	"os"
	"strings"

	"github.com/Venachain/Venachain/common"
	"github.com/syndtr/goleveldb/leveldb"
	"github.com/syndtr/goleveldb/leveldb/opt"
)

var toAddr = common.BytesToAddress

//args: dbPath, dbName, update/query, address, value
func main() {
	args := os.Args
	if len(args) < 6 {
		panic(fmt.Sprintf("invalid args %v\n", args))
	}
	dbOpts := &opt.Options{}
	dbPath := args[1]
	dbName := args[2]
	action := args[3]
	//address := args[4]
	keyName := args[4]
	value := args[5]
	if dbPath[len(dbPath)-1] != '/' {
		dbPath = strings.Join([]string{dbPath, "/"}, "")
	}
	db, err := leveldb.OpenFile(dbPath+dbName, dbOpts)
	if err != nil {
		panic(fmt.Sprintf("Error opening leveldbprovider: %s", err))
	}
	defer db.Close()
	//key := makeKeyWithDbName("", []byte(strings.Join([]string{"asset_new7", address, "balance"}, "#")))
	//address1 := common.HexToAddress(address)
	//keystr := GetKeyTrie(address1, []byte(keyName))
	key := []byte(keyName)
	//key := []byte(keystr)
	//fmt.Printf("address %s, keystr %s, key %v\n", address, key, key)
	fmt.Printf("keystr %s, key %v\n", key, key)
	beforeValue, err := db.Get(key, nil)
	if err != nil && err != leveldb.ErrNotFound {
		panic(fmt.Sprintf("Error getting key: %s", err))
	}
	//if action == "query" {
	//	fmt.Printf("the value of address %s is %s\n", address, string(beforeValue))
	//	return
	//}
	if action == "query" {
		fmt.Printf("the value of is %v\n", beforeValue)
		return
	}
	fmt.Printf("Before modified, the value of  is %v\n", beforeValue)
	err = db.Put(key, []byte(value), nil)
	if err != nil {
		panic(fmt.Sprintf("Error writting key %s: %s", key, err))
	}
	afterValue, err := db.Get(key, nil)
	if err != nil {
		panic(fmt.Sprintf("Error getting key %s: %s", key, err))
	}
	fmt.Printf("After modified, the value is %v\n", afterValue)
}

//func GetKeyTrie(address common.Address, key []byte) string {
//	var buffer bytes.Buffer
//	buffer.WriteString(address.String())
//	buffer.WriteString(string(key))
//	keyTrie := buffer.String()
//	return keyTrie
//}
