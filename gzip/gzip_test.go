package gzip

import (
	"bytes"
	"log/slog"
	"os"
	"slices"
	"testing"
)

func TestCompress(t *testing.T) {
	logger := slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{
		AddSource: true,
		Level:     slog.LevelDebug.Level(),
	}))
	testString := "This is a test string"
	input := bytes.NewBuffer([]byte(testString))
	output := bytes.NewBuffer([]byte{})
	if err := Compress(logger, Params{
		Input:  input,
		Output: output,
	}); err != nil {
		t.Error(err)
	}
	outputData := slices.Clone(output.Bytes())
	outputString := string(outputData)
	t.Log(outputString)
	// if testString != outputString {
	// 	t.Error("test string does not equal output")
	// 	return
	// }
	output2 := bytes.NewBuffer([]byte{})
	header, err := Decompress(logger, Params{
		Input:  output,
		Output: output2,
	})
	if err != nil {
		t.Error("failed to decompress gzip")
		return
	}
	t.Log(header)
	output2Data := slices.Clone(output2.Bytes())
	output2String := string(output2Data)
	if testString != output2String {
		t.Error("test string does not equal output")
		return
	}
}
