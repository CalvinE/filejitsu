package cmd

import (
	"context"
	"errors"
	"fmt"
	"regexp"
	"text/template"

	"github.com/calvine/filejitsu/bulkrename"
	"github.com/calvine/filejitsu/util"
	"github.com/spf13/cobra"
	"golang.org/x/exp/slog"
)

// TODO: add flag to save job to a file and also resume job from a file

type BulkRenameArgs struct {
	RootPath                  string `json:"rootPath"`
	TargetRegexString         string `json:"targetRegexString"`
	DestinationTemplateString string `json:"destinationTemplateString"`
	Recursive                 bool   `json:"recursive"`
	IsTest                    bool   `json:"isTest"`
}

type bkrnProcessingFunction func(*slog.Logger, bulkrename.ResultEntry) error

const bulkRenameCommandName = "bulk-rename"

var bulkRenameCommand = &cobra.Command{
	Use:     bulkRenameCommandName,
	Aliases: []string{"bkrn"},
	Short:   "rename files in bulk",
	Long:    "rename files in bulk based on regex named capture groups",
	RunE:    bulkRenameRun,
}

var bkrnArgs = BulkRenameArgs{}

func bulkRenameInit() {
	bulkRenameCommand.PersistentFlags().StringVarP(&bkrnArgs.RootPath, "rootPath", "p", "", "The root path to perform the bulk rename in")
	bulkRenameCommand.PersistentFlags().StringVarP(&bkrnArgs.TargetRegexString, "targetRegex", "r", "", "The target regex to use for renaming with named capture groups present in the destination regex")
	bulkRenameCommand.PersistentFlags().StringVarP(&bkrnArgs.DestinationTemplateString, "destinationTemplate", "d", "", "The destination template to use for renaming with named capture groups present in the target regex")
	bulkRenameCommand.PersistentFlags().BoolVarP(&bkrnArgs.Recursive, "recursive", "s", false, "If present the bulk rename will work recursively")
	bulkRenameCommand.PersistentFlags().BoolVarP(&bkrnArgs.IsTest, "test", "t", false, "If present rename will not happen, but the rename mapping will be put out to stdout")
	rootCmd.AddCommand(bulkRenameCommand)
}

func bulkRenameRun(cmd *cobra.Command, args []string) error {
	params, err := validateBulkrenameArgs(cmd.Context(), bkrnArgs)
	if err != nil {
		commandLogger.Error("failed to validate command arguments",
			slog.String("error", err.Error()),
		)
		return fmt.Errorf("failed to validate bulk rename arguments: %v", err)
	}
	commandLogger.Debug("args validated successfully",
		slog.Any("params", params),
	)
	results, err := bulkrename.CalculateJobs(cmd.Context(), commandLogger, params)
	if err != nil {
		commandLogger.Error("failed to perform bulk rename",
			slog.String("error", err.Error()),
		)
		return fmt.Errorf("failed to perform bulk rename: %v", err)
	}
	if len(results) == 0 {
		commandLogger.Warn("no files found with name matching input regex",
			slog.String("regex", bkrnArgs.TargetRegexString),
		)
		fmt.Println("no files in root dir matched target regex")
		return nil
	}
	var processingFunction bkrnProcessingFunction
	if params.IsTest {
		commandLogger.Info("command run in test mode")
		fmt.Println("test flag found, printing potential renames")
		processingFunction = bulkrename.ProcessTestResults
	} else {
		commandLogger.Debug("bulk rename being run in normal mode")
		processingFunction = bulkrename.ProcessResult
	}
	for i, result := range results {
		commandLogger.Debug("result set item",
			slog.Int("index", i),
			slog.String("oldName", result.Original.Name),
			slog.String("newName", result.New.Name),
		)
		if err := processingFunction(commandLogger, result); err != nil {
			commandLogger.Error("failed to rename file",
				slog.String("error", err.Error()),
				slog.Any("failedOperation", result),
			)
			return err
		}
	}
	return nil
}

func validateBulkrenameArgs(ctx context.Context, args BulkRenameArgs) (bulkrename.Params, error) {
	params := bulkrename.Params{}
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
	destinationTemplate, err := template.New("filePart").Option("missingkey=error").Funcs(template.FuncMap{
		"padLeft":  util.PadLeft,
		"padRight": util.PadRight,
	}).Parse(args.DestinationTemplateString)
	if err != nil {
		return params, fmt.Errorf("failed to parse destination template: %w", err) // errors.New("failed to parse destination template")
	}
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
