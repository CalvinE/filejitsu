package cmd

import (
	"bytes"
	"io"
	"testing"
)

func TestBase64SimplifiedFromStdInRoundTrip(t *testing.T) {
	inputString := "hey there"
	base64ECommand := SetupCommand("", "")
	input := bytes.NewBufferString(inputString)
	base64ECommand.SetIn(input)
	output := bytes.NewBuffer([]byte{})
	base64ECommand.SetOut(output)
	base64ECommand.SetArgs([]string{
		"b64",
		"e",
	})
	if err := base64ECommand.Execute(); err != nil {
		t.Errorf("failed to execute base64 encode command: %s", err.Error())
		return
	}
	out, err := io.ReadAll(output)
	if err != nil {
		t.Errorf("failed to read base64 encode output: %s", err.Error())
		return
	}
	base64DCommand := SetupCommand("", "")
	input2 := bytes.NewBuffer(out)
	base64DCommand.SetIn(input2)
	output2 := bytes.NewBuffer([]byte{})
	base64DCommand.SetOut(output2)
	base64DCommand.SetArgs([]string{
		"b64",
		"d",
		"-e",
	})
	if err := base64DCommand.Execute(); err != nil {
		t.Errorf("failed to execute base64 decode command: %s", err.Error())
		return
	}
	out2, err := io.ReadAll(output2)
	if err != nil {
		t.Errorf("failed to read base64 decode command output: %s", err.Error())
		return
	}
	if inputString != string(out2) {
		t.Error("input text is not equal to output text")
	}
}

func TestBase64FromStdInRoundTrip(t *testing.T) {
	inputString := "hey there"
	base64ECommand := SetupCommand("", "")
	input := bytes.NewBufferString(inputString)
	base64ECommand.SetIn(input)
	output := bytes.NewBuffer([]byte{})
	base64ECommand.SetOut(output)
	base64ECommand.SetArgs([]string{
		"b64",
	})
	if err := base64ECommand.Execute(); err != nil {
		t.Errorf("failed to execute base64 encode command: %s", err.Error())
		return
	}
	out, err := io.ReadAll(output)
	if err != nil {
		t.Errorf("failed to read base64 encode output: %s", err.Error())
		return
	}
	base64DCommand := SetupCommand("", "")
	input2 := bytes.NewBuffer(out)
	base64DCommand.SetIn(input2)
	output2 := bytes.NewBuffer([]byte{})
	base64DCommand.SetOut(output2)
	base64DCommand.SetArgs([]string{
		"b64",
		"-d",
		"-e",
	})
	if err := base64DCommand.Execute(); err != nil {
		t.Errorf("failed to execute base64 decode command: %s", err.Error())
		return
	}
	out2, err := io.ReadAll(output2)
	if err != nil {
		t.Errorf("failed to read base64 decode command output: %s", err.Error())
		return
	}
	if inputString != string(out2) {
		t.Error("input text is not equal to output text")
	}
}

func TestBase64RoundTrip(t *testing.T) {
	inputString := "hey there"
	base64EncodeCommand := SetupCommand("", "")
	output := bytes.NewBuffer([]byte{})
	base64EncodeCommand.SetOut(output)
	base64EncodeCommand.SetArgs([]string{
		"b64",
		"-t",
		inputString,
	})
	if err := base64EncodeCommand.Execute(); err != nil {
		t.Errorf("failed to execute gzip command: %s", err.Error())
		return
	}
	out, err := io.ReadAll(output)
	if err != nil {
		t.Errorf("failed to read gzip command output: %s", err.Error())
		return
	}
	base64DecodeCommand := SetupCommand("", "")
	output2 := bytes.NewBuffer([]byte{})
	base64DecodeCommand.SetOut(output2)
	base64DecodeCommand.SetArgs([]string{
		"b64",
		"-t",
		string(out),
		"-e",
		"-d",
	})
	if err := base64DecodeCommand.Execute(); err != nil {
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
