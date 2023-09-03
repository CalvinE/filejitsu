package cmd

import (
	"errors"

	"github.com/spf13/cobra"
)

const jsonParseCommandName = "jsonparse"

var jsonParseCommand = &cobra.Command{
	Use:     jsonParseCommandName,
	Aliases: []string{"jp"},
	Short:   "",
	Long:    "",
	RunE:    jsonParseRun,
}

func jsonParseInit() {
	rootCmd.AddCommand(jsonParseCommand)
	// TODO: add flags...
}

func jsonParseRun(cmd *cobra.Command, args []string) error {
	return errors.New("need to implement")
}
