package paillier

/*
#cgo CFLAGS: -I/usr/local/include
#cgo LDFLAGS: -L/usr/local/lib -lgmp -L. -lpaillierChain
#include "./include/paillier_operate.h"
#include <stdlib.h>
*/
import "C"

import (
	"unsafe"
)

// 密文权重加
func PaillierWeightAdd(args []string, arr []uint, pubKey string) string{
	arg := make([]*C.char, 0) //C语言char*指针创建切片
	l := len(args)
	for i, _ := range args {
		char := C.CString(args[i])
		defer C.free(unsafe.Pointer(char)) //释放内存
		strptr := (*C.char)(unsafe.Pointer(char))
		arg = append(arg, strptr) //将char*指针加入到arg切片
	}

	msgPtr := (**C.char)(unsafe.Pointer(&arg[0]))
	var ulongC []C.ulong
	for i, _ := range arr {
		ulongC = append(ulongC, C.ulong(arr[i]))
	}
	cPubKey := C.CString(pubKey)
	defer C.free(unsafe.Pointer(cPubKey))
	CipherWeightAddSum := C.sPaillierWeightAdd(msgPtr, (*C.ulong)(unsafe.Pointer(&ulongC[0])), C.int(l), cPubKey)
	return C.GoString(CipherWeightAddSum)
}

// 密文加法
func PaillierAdd(args []string, pubKey string) string {
	arg := make([]*C.char, 0) //C语言char*指针创建切片
	l := len(args)
	for i, _ := range args {
		char := C.CString(args[i])
		defer C.free(unsafe.Pointer(char)) //释放内存
		strptr := (*C.char)(unsafe.Pointer(char))
		arg = append(arg, strptr) //将char*指针加入到arg切片
	}

	msgPtr := (**C.char)(unsafe.Pointer(&arg[0]))

	cPubKey := C.CString(pubKey)
	defer C.free(unsafe.Pointer(cPubKey))
	CipherAddSum := C.sPaillierAdd(msgPtr, C.int(l), cPubKey)
	return C.GoString(CipherAddSum)
}

// 密文标量乘
func PaillierMul(arg string, scalar uint, pubKey string) string {
	cipher := C.CString(arg)
	defer C.free(unsafe.Pointer(cipher)) //释放内存

	cPubKey := C.CString(pubKey)
	defer C.free(unsafe.Pointer(cPubKey))
	CipherMulPro := C.sPaillierMul(cipher, C.ulong(scalar), cPubKey)
	return C.GoString(CipherMulPro)
}
