package encrypt

import "io"

type Args struct {
	// RootPath                  string `json:"rootPath"`
	// TargetRegexString         string `json:"targetRegexString"`
	// DestinationTemplateString string `json:"destinationTemplateString"`
	// Recursive                 bool   `json:"recursive"`
	// IsTest                    bool   `json:"isTest"`

	FilePath   string `json:"filePath"`
	Passphrase string `json:"passphrase"`
	OutputFile string `json:"outputFile"`
	Encrypt    bool   `json:"encrypt"`
	Decrypt    bool   `json:"decrypt"`
}

type Params struct {
	// RootPath            string             `json:"rootPath"`
	// TargetRegex         *regexp.Regexp     `json:"targetRegex"`
	// DestinationTemplate *template.Template `json:"destinationTemplate"`
	// Recursive           bool               `json:"recursive"`
	// IsTest              bool               `json:"isTest"`
	Input      io.Reader
	Output     io.Writer
	Passphrase []byte `json:"passphrase"`
}
