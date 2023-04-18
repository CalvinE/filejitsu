package main

import (
	"fmt"

	"github.com/calvine/filejitsu/bulkrename"
	"github.com/spf13/cobra"
)

var bulkRenameCommand = &cobra.Command{
	Use:     "bulk-rename",
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
	params, err := bulkrename.ValidateArgs(cmd.Context(), bkrnArgs)
	if err != nil {
		return fmt.Errorf("failed to validate bulk rename arguments: %v", err)
	}
	result, err := bulkrename.Run(cmd.Context(), params)
	if err != nil {
		return fmt.Errorf("failed to perform bulk rename: %v", err)
	}
	if len(result) == 0 {
		fmt.Println("no files in root dir matched target regex")
	}
	if params.IsTest {
		fmt.Println("test flag found, printing potential renames")
		fmt.Println("--------------------------")
		for _, r := range result {
			fmt.Printf("%s => %s\n\n", r.Original.Name, r.New.Name)
		}
		fmt.Println("--------------------------")
	}
	return nil
}
