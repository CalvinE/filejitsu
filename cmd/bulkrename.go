package main

import (
	"errors"
	"fmt"
	"regexp"
	"text/template"

	"github.com/spf13/cobra"
)

type bulkRenameArgs struct {
	RootPath                  string
	TargetRegexString         string
	DestinationTemplateString string
	Recursive                 bool
}

var bulkRenameCommand = &cobra.Command{
	Use:     "bulk-rename",
	Aliases: []string{"bkrn"},
	Short:   "rename files in bulk",
	Long:    "rename files in bulk based on regex named capture groups",
	RunE:    bulkRenameRun,
}

var bkrnArgs = bulkRenameArgs{}

func bulkRenameInit() {
	rootCmd.AddCommand(bulkRenameCommand)
	bulkRenameCommand.PersistentFlags().StringVarP(&bkrnArgs.RootPath, "rootPath", "p", "", "The root path to perform the bulk rename in")
	bulkRenameCommand.PersistentFlags().StringVarP(&bkrnArgs.TargetRegexString, "targetRegex", "t", "", "the target regex to use for renaming with named capture groups present in the destination regex")
	bulkRenameCommand.PersistentFlags().StringVarP(&bkrnArgs.DestinationTemplateString, "destinationTemplate", "d", "", "the destination template to use for renaming with named capture groups present in the target regex")
	bulkRenameCommand.PersistentFlags().BoolVarP(&bkrnArgs.Recursive, "recursive", "r", false, "if present the bulk rename will work recursively")
}

func bulkRenameRun(cmd *cobra.Command, args []string) error {
	// validate root path
	if len(bkrnArgs.RootPath) == 0 {
		return errors.New("root path not specified")
	}
	// validate target regex
	if len(bkrnArgs.TargetRegexString) == 0 {
		return errors.New("target regex not provided")
	}
	targetRegex, err := regexp.Compile(bkrnArgs.TargetRegexString)
	if err != nil {
		return fmt.Errorf("target regex failed to compile: %v", err)
	}
	targetCaptureGroupNames := targetRegex.SubexpNames()
	if len(targetCaptureGroupNames) <= 1 {
		return errors.New("no named capture groups found in target regex")
	}
	// validate destination regex
	if len(bkrnArgs.DestinationTemplateString) == 0 {
		return errors.New("destination template not provided")
	}
	// using .Option("missingkey=error") so that if a named capture group from the regex is missing in the template we should get an error?
	// want to think on this more... Id like to find a way to know ahead of running the template
	// it looks like the template struct contains a lot of data we can possibly use to find the variable names used in the template
	// need to look more into it, but NodeAction (1) nodes in the tree can be navigated and within them there are Pipe.Cmd.Args that
	// contain NodeField(8) that hold the value of the variable name in the template...
	t, err := template.New("filePart").Option("missingkey=error").Parse(bkrnArgs.DestinationTemplateString)
	if err != nil {
		return errors.New("failed to parse destination template")
	}
	fmt.Println(t)
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

	fmt.Println(bkrnArgs)
	return nil
}
