package bulkrename

import (
	"regexp"
	"text/template"

	"github.com/calvine/filejitsu/util"
)

// Params represents the parameters for the bulk rename command.
type Params struct {
	// RootPath is the path to perform the bulk rename operation.
	RootPath string `json:"rootPath"`
	// TargetRegex is used to match files for rename. Any named capture groups in the regex will be available as values for use in the DestinationTemplate.
	TargetRegex *regexp.Regexp `json:"targetRegex"`
	// DestinationTemplate is a GO text template for replacing the part of the file name matched by the TargetRegex with the variables. Named capture groups in the TargetRegex will be available for use in the template.
	DestinationTemplate *template.Template `json:"destinationTemplate"`
	// Recursive if true directories in the RootPath will also be evaluated by the bulk rename operation.
	Recursive bool `json:"recursive"`
	// IsTest if true does a dry-run.
	IsTest bool `json:"isTest"`
}

// ResultEntry is a struct that represents the end state of the rename operation.
type ResultEntry struct {
	// Original is the original name of the file prior to rename.
	Original util.File `json:"original"`
	// New is the new file name after the rename.
	New util.File `json:"new"`
	// DidChange indicates if the rename changed the file name.
	DidChange bool `json:"didChange"`
}
