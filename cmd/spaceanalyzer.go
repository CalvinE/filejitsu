package cmd

import (
	"encoding/json"
	"os"

	"github.com/calvine/filejitsu/util"
	"github.com/spf13/cobra"
	"golang.org/x/exp/slog"
)

type SpaceAnalyzerArgs struct {
	RootPath            string `json:"rootPath"`
	MaxRecursion        int    `json:"maxRecursion"`
	CalculateFileHashes bool   `json:"calculateFileHashes"`
}

const spaceAnalyzerCommandName = "space-analyzer"

var spaceAnalyzerCommand = &cobra.Command{
	Use:     spaceAnalyzerCommandName,
	Aliases: []string{"sa"},
	Short:   "Analyze storage usage in a given directory",
	Long:    "Analyze storage usage in a given directory. Outputs a JSON object with data on all of the content.",
	RunE:    spaceAnalyzerRun,
}

var spaceAnalyzerArgs = SpaceAnalyzerArgs{}

func spaceAnalyzerInit() {
	spaceAnalyzerCommand.PersistentFlags().StringVarP(&spaceAnalyzerArgs.RootPath, "rootPath", "p", ".", "The root path to analyze. Default is current directory.")
	spaceAnalyzerCommand.PersistentFlags().IntVarP(&spaceAnalyzerArgs.MaxRecursion, "maxRecursion", "m", -1, "Max  number of recursive calls allowed. -1 means no limit")
	spaceAnalyzerCommand.PersistentFlags().BoolVarP(&spaceAnalyzerArgs.CalculateFileHashes, "calculateFileHashes", "c", false, "If present file hashes will be calculated on files")
	rootCmd.AddCommand(spaceAnalyzerCommand)
}

func spaceAnalyzerRun(cmd *cobra.Command, args []string) error {
	commandLogger := logger.With(slog.String("commandName", spaceAnalyzerCommandName))
	commandLogger.Debug("starting command",
		slog.Any("args", spaceAnalyzerArgs),
	)
	defer commandLogger.Debug("ending command")
	info, err := util.GetDirContentDetails(commandLogger, spaceAnalyzerArgs.RootPath, "", spaceAnalyzerArgs.CalculateFileHashes, spaceAnalyzerArgs.MaxRecursion, 0)
	if err != nil {
		commandLogger.Error("failed to get dir content details", slog.String("errorMessage", err.Error()), slog.String("rootPath", spaceAnalyzerArgs.RootPath))
		return err
	}
	infoString, err := json.MarshalIndent(info, "", "  ")
	if err != nil {
		commandLogger.Error("failed to marshal data to JSON", slog.String("errorMessage", err.Error()))
		return err
	}
	output := os.Stdout
	_, err = output.WriteString(string(infoString))
	if err != nil {
		commandLogger.Error("failed to write data to stdout", slog.String("errorMessage", err.Error()))
	}
	return nil
}
