package encrypt

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"crypto/sha256"
	"fmt"
	"io"

	"github.com/calvine/filejitsu/util"
	"golang.org/x/exp/slog"
)

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
	err = util.ProcessStreams(logger, params.Input, cipherStream)
	if err != nil {
		logger.Error("failed to encrypt data", slog.String("errorMessage", err.Error()))
		return err
	}
	logger.Debug("done encrypting input")
	return nil
}
