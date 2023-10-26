package encrypt

import "io"

// Operation is a type alias for indicating the operation desired.
type Operation int8

//go:generate stringer -type=Operation
const (
	// OpEncrypt indicates that encryption is the desired operation.
	OpEncrypt Operation = iota + 1
	// OpDecrypt indicates that decryption is the desired operation.
	OpDecrypt
	// OpPassThrough indicates that passthrough is the desired operation.
	OpPassThrough
)

// Represents the parameters for the encrypt / decrypt command.
type Params struct {
	// Input is a reader that will be used as input for the operation.
	Input io.Reader
	// Output is a writer that will be used as the output for the operation.
	Output io.Writer
	// Operation is the desired operation.
	Operation Operation
	// Passphrase the passphrase used for encryption or decryption per the Operation.
	Passphrase []byte `json:"passphrase"`
}
