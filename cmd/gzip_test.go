package cmd

import (
	"bytes"
	"io"
	"testing"
)

func TestGZIPRoundTrip(t *testing.T) {
	inputString := "hey there"
	gzipCommand := SetupCommand("", "")
	output := bytes.NewBuffer([]byte{})
	gzipCommand.SetOut(output)
	gzipCommand.SetArgs([]string{
		"gzip",
		"-t",
		inputString,
	})
	if err := gzipCommand.Execute(); err != nil {
		t.Errorf("failed to execute gzip command: %s", err.Error())
		return
	}
	out, err := io.ReadAll(output)
	if err != nil {
		t.Errorf("failed to read gzip command output: %s", err.Error())
		return
	}
	gunzipCommand := SetupCommand("", "")
	output2 := bytes.NewBuffer([]byte{})
	gunzipCommand.SetOut(output2)
	gunzipCommand.SetArgs([]string{
		"gunzip",
		"-t",
		string(out),
	})
	if err := gunzipCommand.Execute(); err != nil {
		t.Errorf("failed to execute gzip command: %s", err.Error())
		return
	}
	out2, err := io.ReadAll(output2)
	if err != nil {
		t.Errorf("failed to read gzip command output: %s", err.Error())
		return
	}
	if inputString != string(out2) {
		t.Error("input text is not equal to output text")
	}
}
