package encrypt

import (
	"bytes"
	"fmt"
	"os"
	"testing"

	"golang.org/x/exp/slices"
	"golang.org/x/exp/slog"
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
			err := Encrypt(logger, Params{
				Input:      inputBuffer,
				Output:     outputBuffer,
				Passphrase: tc.passphrase,
			})
			if err != nil {
				t.Errorf("encryption failed with error: %v", err)
			}
			encryptedData := slices.Clone(outputBuffer.Bytes())
			t.Logf("encrypted data len %d", len(encryptedData))
			err = Decrypt(logger, Params{
				Input:      outputBuffer,
				Output:     inputBuffer,
				Passphrase: tc.passphrase,
			})
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
