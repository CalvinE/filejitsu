package encrypt

import (
	"bytes"
	"fmt"
	"os"
	"testing"

	"log/slog"

	"golang.org/x/exp/slices"
)

func TestEncryptDecrypt(t *testing.T) {
	logger := slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{
		AddSource: true,
		Level:     slog.LevelDebug.Level(),
	}))
	type testCase struct {
		data       []byte
		passphrase []byte
	}
	testCases := []testCase{
		{
			data:       []byte("This is a test string"),
			passphrase: []byte("testpass"),
		},
	}
	for i, tc := range testCases {
		t.Run(fmt.Sprintf("test case #%d", i+1), func(t *testing.T) {
			originalData := slices.Clone(tc.data)
			inputBuffer := bytes.NewBuffer(tc.data)
			outputBuffer := bytes.NewBuffer([]byte{})
			encryptionWriter, err := NewAESEncryptionWriter(logger, outputBuffer, tc.passphrase)
			if err != nil {
				t.Errorf("failed to create encryption writer: %v", err)
			}
			err = Encrypt(logger, inputBuffer, encryptionWriter)
			if err != nil {
				t.Errorf("encryption failed with error: %v", err)
			}
			encryptedData := slices.Clone(outputBuffer.Bytes())
			t.Logf("encrypted data len %d", len(encryptedData))
			decryptionReader, err := NewAESDecryptionReader(logger, outputBuffer, tc.passphrase)
			if err != nil {
				t.Errorf("failed to create decryption reader: %v", err)
			}
			err = Decrypt(logger, decryptionReader, inputBuffer)
			if err != nil {
				t.Errorf("decryption failed with error: %v", err)
			}
			decryptedData := slices.Clone(inputBuffer.Bytes())
			if !slices.Equal(originalData, decryptedData) {
				t.Error("original data and decrypted data are not the same...")
			}
		})
	}
}
