package cmd

import (
	"bytes"
	"io"
	"testing"
)

// TODO: make a test for --passphraseFile

func TestEncryptDecryptFromStdInRoundTrip(t *testing.T) {
	passphrase := "weaqliugvnjieuwn98738r9o87GPI*UYGYOG\\(*O\\)PG&O"
	inputString := "hey there"
	encryptCommand := SetupCommand("", "", "")
	input := bytes.NewBufferString(inputString)
	encryptCommand.SetIn(input)
	output := bytes.NewBuffer([]byte{})
	encryptCommand.SetOut(output)
	encryptCommand.SetArgs([]string{
		"encr",
		"-p",
		passphrase,
	})
	if err := encryptCommand.Execute(); err != nil {
		t.Errorf("failed to execute gzip command: %s", err.Error())
		return
	}
	out, err := io.ReadAll(output)
	if err != nil {
		t.Errorf("failed to read gzip command output: %s", err.Error())
		return
	}
	decryptCommand := SetupCommand("", "", "")
	input2 := bytes.NewBuffer(out)
	decryptCommand.SetIn(input2)
	output2 := bytes.NewBuffer([]byte{})
	decryptCommand.SetOut(output2)
	decryptCommand.SetArgs([]string{
		"dcry",
		"-p",
		passphrase,
	})
	if err := decryptCommand.Execute(); err != nil {
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

func TestEncryptDecryptRoundTrip(t *testing.T) {
	passphrase := "weaqliugvnjieuwn98738r9o87GPI*UYGYOG\\(*O\\)PG&O"
	inputString := "hey there"
	encryptCommand := SetupCommand("", "", "")
	output := bytes.NewBuffer([]byte{})
	encryptCommand.SetOut(output)
	encryptCommand.SetArgs([]string{
		"encr",
		"-t",
		inputString,
		"-p",
		passphrase,
	})
	if err := encryptCommand.Execute(); err != nil {
		t.Errorf("failed to execute gzip command: %s", err.Error())
		return
	}
	out, err := io.ReadAll(output)
	if err != nil {
		t.Errorf("failed to read gzip command output: %s", err.Error())
		return
	}
	decryptCommand := SetupCommand("", "", "")
	output2 := bytes.NewBuffer([]byte{})
	decryptCommand.SetOut(output2)
	decryptCommand.SetArgs([]string{
		"dcry",
		"-t",
		string(out),
		"-p",
		passphrase,
	})
	if err := decryptCommand.Execute(); err != nil {
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
