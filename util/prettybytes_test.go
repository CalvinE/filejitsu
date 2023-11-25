package util

import (
	"testing"
)

func TestGetPrettyBytesSize(t *testing.T) {
	type testCase struct {
		Name          string
		Bytes         int64
		ExpectedValue string
	}
	testCases := []testCase{
		{
			Name:          "KB test",
			Bytes:         16570,
			ExpectedValue: "16.18 KB",
		},
		{
			Name:          "MB test",
			Bytes:         97208320,
			ExpectedValue: "92.71 MB",
		},
		{
			Name:          "GB test",
			Bytes:         15229071494,
			ExpectedValue: "14.18 GB",
		},
	}
	for _, tc := range testCases {
		t.Run(tc.Name, func(t *testing.T) {
			val := GetPrettyBytesSize(tc.Bytes)
			if val != tc.ExpectedValue {
				t.Errorf("%d returned %s but expected %s", tc.Bytes, val, tc.ExpectedValue)
				return
			}

		})
	}
}
