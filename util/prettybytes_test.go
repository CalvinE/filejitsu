package util

import (
	"runtime"
	"testing"
)

func TestGetPrettyBytesSize(t *testing.T) {
	x := int64(1023624856) // int64(279370246)
	val := GetPrettyBytesSize(x)
	// TODO: Clean up
	numCPUs := runtime.NumCPU()
	t.Log(val, numCPUs)
}
