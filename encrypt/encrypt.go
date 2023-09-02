package encrypt

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"crypto/sha256"
	"errors"
	"fmt"
	"io"

	"golang.org/x/exp/slog"
)

var (
	ErrPassphraseTooLongCannotTrim = errors.New("passphrase length is longer than desiredLength and allowTrim is false")
)

func CMSPadOrTrim(passphrase []byte, desiredLength int, allowTrim bool) ([]byte, error) {
	passphraseLen := len(passphrase)
	if passphraseLen == 0 {
		return nil, errors.New("passphrase cannot be empty")
	}
	if passphraseLen == desiredLength {
		return passphrase, nil
	}
	if passphraseLen > desiredLength {
		if !allowTrim {
			return nil, ErrPassphraseTooLongCannotTrim
		}
		return passphrase[:desiredLength], nil
	}
	missingCount := desiredLength - passphraseLen
	newBytes := make([]byte, desiredLength)
	for i := 0; i < passphraseLen; i++ {
		newBytes[i] = passphrase[i]
	}
	for i := passphraseLen; i < desiredLength; i++ {
		newBytes[i] = byte(missingCount)
	}
	return newBytes, nil
}

func ValidateArgs(logger *slog.Logger, args Args) (Params, error) {
	params := Params{}

	return params, nil
}

func Encrypt(logger *slog.Logger, params Params) error {
	iv := make([]byte, aes.BlockSize)
	if _, err := io.ReadFull(rand.Reader, iv); err != nil {
		logger.Error("failed to generate random nonce", slog.String("errorMessage", err.Error()))
		return err
	}
	ivLen := len(iv)
	logger.Debug("writing iv to encrypted file", slog.Int("ivLen", ivLen))
	n, err := params.Output.Write(iv)
	if err != nil {
		logger.Error("failed to write iv to encrypted file", slog.String("errorMessage", err.Error()))
		return err
	}
	if n != ivLen {
		err := fmt.Errorf("number of bytes written does not equal iv length: wrote %d - expected %d", n, ivLen)
		logger.Error("failed to write iv to encrypted file", slog.String("errorMessage", err.Error()))
		return err
	}
	hashedPassword := sha256.Sum256(params.Passphrase)
	block, err := aes.NewCipher(hashedPassword[:])
	if err != nil {
		logger.Error("creating cipher block failed", slog.String("errorMessage", err.Error()))
		return err
	}
	stream := cipher.NewOFB(block, iv)
	cipherStream := cipher.StreamWriter{
		S: stream,
		W: params.Output,
	}
	err = processStreams(logger, params.Input, cipherStream)
	if err != nil {
		logger.Error("failed to encrypt data", slog.String("errorMessage", err.Error()))
		return err
	}
	logger.Debug("done encrypting input")
	return nil
}

func Decrypt(logger *slog.Logger, params Params) error {
	iv := make([]byte, aes.BlockSize)
	n, err := params.Input.Read(iv)
	if err != nil {
		logger.Error("failed to read iv from input", slog.String("errorMessage", err.Error()))
		return err
	}
	ivLen := len(iv)
	if n != ivLen {
		err := fmt.Errorf("iv read from input not expected length: got %d - expected %d", n, ivLen)
		logger.Error("failed to read iv from input", slog.String("errorMessage", err.Error()), slog.Int("got", n), slog.Int("expected", ivLen))
		return err
	}
	hashedPassphrase := sha256.Sum256(params.Passphrase)
	block, err := aes.NewCipher(hashedPassphrase[:])
	if err != nil {
		logger.Error("failed to create new cipher block", slog.String("errorMessage", err.Error()))
		return err
	}
	stream := cipher.NewOFB(block, iv)
	cipherStream := cipher.StreamReader{S: stream, R: params.Input}
	err = processStreams(logger, cipherStream, params.Output)
	if err != nil {
		logger.Error("failed to decrypt data", slog.String("errorMessage", err.Error()))
		return err
	}
	logger.Debug("done decrypting input")
	return nil
}

// processStreams writes data from an io.Reader to and io.Writer
func processStreams(logger *slog.Logger, input io.Reader, output io.Writer) error {
	inputBuffer := make([]byte, 64)
	bytesRead, bytesWritten := 0, 0
	done := false
	for !done {
		rn, err := input.Read(inputBuffer)
		bytesRead += rn
		if err != nil {
			if err == io.EOF {
				// we are done here
				done = true
			} else {
				logger.Error("failed to read data from input buffer", slog.String("errorMessage", err.Error()))
				return err
			}
		}
		if rn > 0 {
			wn, err := output.Write(inputBuffer[:rn])
			bytesWritten += wn
			if err != nil {
				logger.Error("failed to write to output stream", slog.String("errorMessage", err.Error()))
				return err
			}
		}
	}
	logger.Debug("done processing stream", slog.Int("bytesWritten", bytesWritten), slog.Int("bytesRead", bytesRead))
	return nil
}
