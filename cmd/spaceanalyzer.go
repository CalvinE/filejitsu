package cmd

import (
	"context"
	"io"
	"os"

	"github.com/calvine/filejitsu/spaceanalyzer"
	"github.com/calvine/filejitsu/util/streamingjson"
	"github.com/spf13/cobra"
)

type SpaceAnalyzerArgs struct {
	RootPath            string `json:"rootPath"`
	MaxRecursion        int    `json:"maxRecursion"`
	CalculateFileHashes bool   `json:"calculateFileHashes"`
	// ExistingAnalysisFile string `json:"existingAnalysisFile"`
}

const spaceAnalyzerCommandName = "space-analyzer"

var spaceAnalyzerCommand = &cobra.Command{
	Use:     spaceAnalyzerCommandName,
	Aliases: []string{"sa"},
	Short:   "Analyze storage usage in a given directory",
	Long:    "Analyze storage usage in a given directory. Outputs a JSON object with data on all of the content.",
	RunE:    spaceAnalyzerScanRun,
}

var spaceAnalyzerArgs = SpaceAnalyzerArgs{}

func spaceAnalyzerInit() {
	spaceAnalyzerCommand.PersistentFlags().StringVarP(&spaceAnalyzerArgs.RootPath, "rootPath", "p", ".", "The root path to analyze. Default is current directory.")
	spaceAnalyzerCommand.PersistentFlags().IntVarP(&spaceAnalyzerArgs.MaxRecursion, "maxRecursion", "m", -1, "Max number of recursive calls allowed. -1 means no limit")
	spaceAnalyzerCommand.PersistentFlags().BoolVarP(&spaceAnalyzerArgs.CalculateFileHashes, "calculateFileHashes", "c", false, "If present file hashes will be calculated on files")
	// spaceAnalyzerCommand.PersistentFlags().BoolVarP(&spaceAnalyzerArgs.ExistingAnalysisFile, "existingAnalyzerFile", "e", "", "An existing analysis file from a previous")
	rootCmd.AddCommand(spaceAnalyzerCommand)
}

func spaceAnalyzerScanRun(cmd *cobra.Command, args []string) error {
	info, err := spaceanalyzer.Scan(commandLogger, spaceanalyzer.ScanParams{
		RootPath:            spaceAnalyzerArgs.RootPath,
		MaxRecursion:        spaceAnalyzerArgs.MaxRecursion,
		CalculateFileHashes: spaceAnalyzerArgs.CalculateFileHashes,
	})
	if err != nil {
		return err
	}
	// infoString, err := json.MarshalIndent(info, "", "  ")
	// if err != nil {
	// 	commandLogger.Error("failed to marshal data to JSON", slog.String("errorMessage", err.Error()))
	// 	return err
	// }
	output := os.Stdout
	// TODO: add a option to enable streaming JSON and if not present do normal ish.
	jsonStreamer := streamingjson.NewLengthPrefixStreamJSONHandler[spaceanalyzer.FSEntity]()
	WriteOutputAsStreamingJSON(cmd.Context(), info, output, jsonStreamer)
	// jsonStreamer.WriteObject(cmd.Context(), info, os.Stdout)
	// _, err = output.WriteString(string(infoString))
	// if err != nil {
	// 	commandLogger.Error("failed to write data to stdout", slog.String("errorMessage", err.Error()))
	// }
	return nil
}

func WriteOutputAsStreamingJSON(ctx context.Context, rootInfo spaceanalyzer.FSEntity, writer io.Writer, streamingHandler streamingjson.StreamingJSONWriter[spaceanalyzer.FSEntity]) error {
	if rootInfo.IsDir {
		for _, i := range rootInfo.Children {
			if err := WriteOutputAsStreamingJSON(ctx, i, writer, streamingHandler); err != nil {
				return err
			}
		}
		rootInfo.Children = nil
	}
	_, err := streamingHandler.WriteObject(ctx, rootInfo, os.Stdout)
	return err
}
