package main

import (
	"errors"
	"fmt"
	"regexp"

	"github.com/spf13/cobra"
)

type bulkRenameArgs struct {
	RootPath               string
	TargetRegexString      string
	DestinationRegexString string
	Recursive              bool
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
	bulkRenameCommand.PersistentFlags().StringVarP(&bkrnArgs.DestinationRegexString, "destinationRegex", "d", "", "the destination regex to use for renaming with named capture groups present in the target regex")
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
	if len(bkrnArgs.DestinationRegexString) == 0 {
		return errors.New("destination regex not provided")
	}
	destinationRegex, err := regexp.Compile(bkrnArgs.DestinationRegexString)
	if err != nil {
		return fmt.Errorf("destination regex failed to compile: %v", err)
	}
	destinationCaptureGroupNames := destinationRegex.SubexpNames()
	if len(destinationCaptureGroupNames) <= 1 {
		return errors.New("no named capture groups found in destination regex")
	}
	// we need to make sure for each named destination capture group we have a comparable named capture group in the target regex
	for i := 0; i < len(destinationCaptureGroupNames); i++ {
		destName := destinationCaptureGroupNames[i]
		found := false
		for j := 0; j < len(targetCaptureGroupNames); j++ {
			if destName == targetCaptureGroupNames[j] {
				found = true
				break
			}
		}
		if !found {
			return fmt.Errorf("target regex is missing named capture group found in destination regex: %s", destName)
		}
	}

	fmt.Println(bkrnArgs)
	return nil
}
