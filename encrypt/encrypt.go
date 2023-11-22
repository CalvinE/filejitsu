package encrypt

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"crypto/sha256"
	"fmt"
	"io"

	"log/slog"

	"github.com/calvine/filejitsu/util"
)

func NewAESEncryptionWriter(logger *slog.Logger, output io.Writer, passphrase []byte) (*cipher.StreamWriter, error) {
	iv := make([]byte, aes.BlockSize)
	if _, err := io.ReadFull(rand.Reader, iv); err != nil {
		logger.Error("failed to generate random nonce", slog.String("errorMessage", err.Error()))
		return nil, err
	}
	ivLen := len(iv)
	logger.Debug("writing iv to encrypted file", slog.Int("ivLen", ivLen))
	n, err := output.Write(iv)
	if err != nil {
		logger.Error("failed to write iv to encrypted file", slog.String("errorMessage", err.Error()))
		return nil, err
	}
	if n != ivLen {
		err := fmt.Errorf("number of bytes written does not equal iv length: wrote %d - expected %d", n, ivLen)
		logger.Error("failed to write iv to encrypted file", slog.String("errorMessage", err.Error()))
		return nil, err
	}
	hashedPassword := sha256.Sum256(passphrase)
	block, err := aes.NewCipher(hashedPassword[:])
	if err != nil {
		logger.Error("creating cipher block failed", slog.String("errorMessage", err.Error()))
		return nil, err
	}
	stream := cipher.NewOFB(block, iv)
	cipherStream := cipher.StreamWriter{
		S: stream,
		W: output,
	}
	return &cipherStream, nil
}

func Encrypt(logger *slog.Logger, input io.Reader, output *cipher.StreamWriter) error {
	// iv := make([]byte, aes.BlockSize)
	// if _, err := io.ReadFull(rand.Reader, iv); err != nil {
	// 	logger.Error("failed to generate random nonce", slog.String("errorMessage", err.Error()))
	// 	return err
	// }
	// ivLen := len(iv)
	// logger.Debug("writing iv to encrypted file", slog.Int("ivLen", ivLen))
	// n, err := params.Output.Write(iv)
	// if err != nil {
	// 	logger.Error("failed to write iv to encrypted file", slog.String("errorMessage", err.Error()))
	// 	return err
	// }
	// if n != ivLen {
	// 	err := fmt.Errorf("number of bytes written does not equal iv length: wrote %d - expected %d", n, ivLen)
	// 	logger.Error("failed to write iv to encrypted file", slog.String("errorMessage", err.Error()))
	// 	return err
	// }
	// hashedPassword := sha256.Sum256(params.Passphrase)
	// block, err := aes.NewCipher(hashedPassword[:])
	// if err != nil {
	// 	logger.Error("creating cipher block failed", slog.String("errorMessage", err.Error()))
	// 	return err
	// }
	// stream := cipher.NewOFB(block, iv)
	// cipherStream := cipher.StreamWriter{
	// 	S: stream,
	// 	W: params.Output,
	// }
	err := util.ProcessStreams(logger, input, output)
	if err != nil {
		logger.Error("failed to encrypt data", slog.String("errorMessage", err.Error()))
		return err
	}
	logger.Debug("done encrypting input")
	return nil
}
