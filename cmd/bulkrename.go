package main

import (
	"fmt"

	"github.com/calvine/filejitsu/bulkrename"
	"github.com/spf13/cobra"
	"golang.org/x/exp/slog"
)

// TODO: add flag to save job to a file and also resume job from a file

type bkrnProcessingFunction func(*slog.Logger, bulkrename.ResultEntry) error

const bulkRenameCommandName = "bulk-rename"

var bulkRenameCommand = &cobra.Command{
	Use:     bulkRenameCommandName,
	Aliases: []string{"bkrn"},
	Short:   "rename files in bulk",
	Long:    "rename files in bulk based on regex named capture groups",
	RunE:    bulkRenameRun,
}

var bkrnArgs = bulkrename.Args{}

func bulkRenameInit() {
	rootCmd.AddCommand(bulkRenameCommand)
	bulkRenameCommand.PersistentFlags().StringVarP(&bkrnArgs.RootPath, "rootPath", "p", "", "The root path to perform the bulk rename in")
	bulkRenameCommand.PersistentFlags().StringVarP(&bkrnArgs.TargetRegexString, "targetRegex", "r", "", "the target regex to use for renaming with named capture groups present in the destination regex")
	bulkRenameCommand.PersistentFlags().StringVarP(&bkrnArgs.DestinationTemplateString, "destinationTemplate", "d", "", "the destination template to use for renaming with named capture groups present in the target regex")
	bulkRenameCommand.PersistentFlags().BoolVarP(&bkrnArgs.Recursive, "recursive", "s", false, "if present the bulk rename will work recursively")
	bulkRenameCommand.PersistentFlags().BoolVarP(&bkrnArgs.IsTest, "test", "t", false, "if present rename will not happen, but the rename mapping will be put out to stdout")
}

func bulkRenameRun(cmd *cobra.Command, args []string) error {
	commandLogger := logger.With(slog.String("commandName", bulkRenameCommandName))
	commandLogger.Debug("starting command",
		slog.String("name", bulkRenameCommandName),
		slog.Any("args", bkrnArgs),
	)
	defer commandLogger.Debug("ending command",
		slog.String("name", bulkRenameCommandName),
	)
	params, err := bulkrename.ValidateArgs(cmd.Context(), bkrnArgs)
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
			logger.Error("failed to rename file",
				slog.String("error", err.Error()),
				slog.Any("failedOperation", result),
			)
			return err
		}
	}
	return nil
}
