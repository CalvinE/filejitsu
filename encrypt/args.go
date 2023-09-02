package encrypt

import "io"

type Operation int8

//go:generate stringer -type=Operation
const (
	OpEncrypt Operation = iota + 1
	OpDecrypt
)

type Params struct {
	Input      io.Reader
	Output     io.Writer
	Operation  Operation
	Passphrase []byte `json:"passphrase"`
}
