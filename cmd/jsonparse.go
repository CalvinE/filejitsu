package cmd

import (
	"encoding/json"
	"errors"

	"github.com/spf13/cobra"
)

type JsonParseArgs struct {
	Input []byte `json:"input"`
}

var jsonParseArgs = JsonParseArgs{}

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
	var data map[string]interface{}
	json.Unmarshal(jsonParseArgs.Input, &data)
	return errors.New("need to implement")
}
