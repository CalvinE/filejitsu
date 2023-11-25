package cmd

import (
	"encoding/json"
	"fmt"
	"log/slog"

	"github.com/spf13/cobra"
)

const versionCommandName = "version"

type versionInfo struct {
	BuildTime string `json:"buildDate"`
	BuildHash string `json:"commitHash"`
	Version   string `json:"version"`
}

func newVersionCommand(buildTime, buildHash, version string) *cobra.Command {
	return &cobra.Command{
		Use:   versionCommandName,
		Short: "Get the version information about this build of filejitsu",
		RunE: func(cmd *cobra.Command, args []string) error {
			vi := versionInfo{
				BuildTime: buildTime,
				BuildHash: buildHash,
				Version:   version,
			}
			viString, err := json.MarshalIndent(vi, "", "  ")
			if err != nil {
				errMsg := "failed to marshal version info"
				commandLogger.Error(errMsg, slog.String("errorMessage", err.Error()))
				return fmt.Errorf("%s: %w", errMsg, err)
			}

			// viString = append(viString, []byte(util.NewLine)...)

			_, err = outputFile.Write(viString)
			if err != nil {
				errMsg := "failed to write version info to output file"
				commandLogger.Error(errMsg, slog.String("errorMessage", err.Error()))
				return fmt.Errorf("%s: %w", errMsg, err)
			}

			return nil
		},
	}
}

func versionInit(parentCommand *cobra.Command, buildTime, buildHash, version string) {
	versionCommand := newVersionCommand(buildTime, buildHash, version)
	parentCommand.AddCommand(versionCommand)
}
