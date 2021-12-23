package main

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func Test_curDir(t *testing.T) {
	dir := curDir()
	t.Log("cur dir:", dir)
	assert.NotEqual(t, "", dir)
}

func Test_calcAbsPath(t *testing.T) {
	abspath := calcAbsPath("venachain")
	assert.Contains(t, abspath, "Venachain/cmd/localtestnet/venachain")
}
