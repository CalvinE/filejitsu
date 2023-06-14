package bulkrename

import (
	"regexp"
	"text/template"

	"github.com/calvine/filejitsu/util"
)

type Args struct {
	RootPath                  string `json:"rootPath"`
	TargetRegexString         string `json:"targetRegexString"`
	DestinationTemplateString string `json:"destinationTemplateString"`
	Recursive                 bool   `json:"recursive"`
	IsTest                    bool   `json:"isTest"`
}

type Params struct {
	RootPath            string             `json:"rootPath"`
	TargetRegex         *regexp.Regexp     `json:"targetRegex"`
	DestinationTemplate *template.Template `json:"destinationTemplate"`
	Recursive           bool               `json:"recursive"`
	IsTest              bool               `json:"isTest"`
}

type ResultEntry struct {
	Original  util.File `json:"original"`
	New       util.File `json:"new"`
	DidChange bool      `json:"didChange"`
}
