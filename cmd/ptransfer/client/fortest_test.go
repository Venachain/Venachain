package main

import (
	"fmt"
	"testing"

	utl "github.com/Venachain/Venachain/cmd/utils"
)

func TestAdd(t *testing.T) {
	name := ""
	if name != `^[a-z0-9A-Z\p{Han}]+(_[a-z0-9A-Z\p{Han}]+)*$` {
		utl.Fatalf(fmt.Sprintf("filename is illegal: %v\n"))
	}
	if name == "null" {
		utl.Fatalf(fmt.Sprintf("filename is null: %v\n"))
	}
}
