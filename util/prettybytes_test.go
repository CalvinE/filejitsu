package util

import "testing"

func TestGetPrettyBytesSize(t *testing.T) {
	x := int64(279370246)
	val := GetPrettyBytesSize(x)
	t.Log(val)
}
