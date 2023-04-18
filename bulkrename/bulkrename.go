package bulkrename

import (
	"context"
	"errors"
	"fmt"
	"os"
	"path"
	"path/filepath"
	"regexp"
	"text/template"

	"github.com/calvine/filejitsu/util"
)

func ValidateArgs(ctx context.Context, args Args) (Params, error) {
	params := Params{}
	// validate root path
	if len(args.RootPath) == 0 {
		return params, errors.New("root path not specified")
	}
	params.RootPath = args.RootPath
	// validate target regex
	if len(args.TargetRegexString) == 0 {
		return params, errors.New("target regex not provided")
	}
	targetRegex, err := regexp.Compile(args.TargetRegexString)
	if err != nil {
		return params, fmt.Errorf("target regex failed to compile: %v", err)
	}
	targetCaptureGroupNames := targetRegex.SubexpNames()
	if len(targetCaptureGroupNames) <= 1 {
		return params, errors.New("no named capture groups found in target regex")
	}
	params.TargetRegex = targetRegex
	// validate destination template
	if len(args.DestinationTemplateString) == 0 {
		return params, errors.New("destination template not provided")
	}
	// using .Option("missingkey=error") so that if a named capture group from the regex is missing in the template we should get an error?
	// want to think on this more... Id like to find a way to know ahead of running the template
	// it looks like the template struct contains a lot of data we can possibly use to find the variable names used in the template
	// need to look more into it, but NodeAction (1) nodes in the tree can be navigated and within them there are Pipe.Cmd.Args that
	// contain NodeField(8) that hold the value of the variable name in the template...
	destinationTemplate, err := template.New("filePart").Option("missingkey=error").Parse(args.DestinationTemplateString)
	if err != nil {
		return params, errors.New("failed to parse destination template")
	}
	fmt.Println(destinationTemplate)
	params.DestinationTemplate = destinationTemplate
	// we need to make sure for each named destination capture group we have a comparable named capture group in the target regex
	// for i := 0; i < len(destinationCaptureGroupNames); i++ {
	// 	destName := destinationCaptureGroupNames[i]
	// 	found := false
	// 	for j := 0; j < len(targetCaptureGroupNames); j++ {
	// 		if destName == targetCaptureGroupNames[j] {
	// 			found = true
	// 			break
	// 		}
	// 	}
	// 	if !found {
	// 		return fmt.Errorf("target regex is missing named capture group found in destination regex: %s", destName)
	// 	}
	// }
	params.Recursive = args.Recursive
	params.IsTest = args.IsTest
	return params, nil
}

func Run(ctx context.Context, params Params) ([]ResultEntry, error) {
	if params.IsTest {
		fmt.Println("test mode flag present. changes will not be made")
	}
	targetFiles := make([]util.File, 0)
	// enumerate root dir
	content, err := os.ReadDir(params.RootPath)
	if err != nil {
		return nil, fmt.Errorf("failed to enumerate root dir: %v", err)
	}
	// look at root dir contents and try to find files that match the target regex
	for _, c := range content {
		i, err := c.Info()
		if err != nil {
			// ?
		}
		if !i.IsDir() {
			fName := i.Name()
			matchesTarget := params.TargetRegex.FindStringSubmatch(fName)
			if len(matchesTarget) == 0 {
				// file does not match target regex so we can move forward
				continue
			}
			absPath, err := filepath.Abs(params.RootPath)
			if err != nil {
				// ?
			}
			// file name matches target regex so add to target files
			f := util.File{
				Name:      fName,
				Size:      i.Size(),
				Extension: path.Ext(fName),
				Path:      absPath, // may want to convert to absolute path?
			}
			targetFiles = append(targetFiles, f)
		}
	}
	if len(targetFiles) == 0 {
		// no files were found with target regex... just return nothing?
		return nil, nil
	}

	return nil, nil
}
