package cmd

import (
	"context"
	"encoding/json"
	"io"

	"log/slog"

	"github.com/calvine/filejitsu/spaceanalyzer"
	"github.com/calvine/filejitsu/util"
	"github.com/calvine/filejitsu/util/streamingjson"
	"github.com/spf13/cobra"
)

type SpaceAnalyzerArgs struct {
	RootPath            string `json:"rootPath"`
	MaxRecursion        int    `json:"maxRecursion"`
	CalculateFileHashes bool   `json:"calculateFileHashes"`
	OutputFormat        string `json:"outputFormat"`
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
	spaceAnalyzerCommand.PersistentFlags().StringVarP(&spaceAnalyzerArgs.OutputFormat, "outputFormat", "f", "json", "Output format for scan data. Options are 'json' or 'sjson' for streaming json")
	// spaceAnalyzerCommand.PersistentFlags().BoolVarP(&spaceAnalyzerArgs.ExistingAnalysisFile, "existingAnalyzerFile", "e", "", "An existing analysis file from a previous")
	rootCmd.AddCommand(spaceAnalyzerCommand)
}

func spaceAnalyzerScanRun(cmd *cobra.Command, args []string) error {
	commandLogger.Debug("args provided", slog.Any("args", spaceAnalyzerArgs))
	info, err := spaceanalyzer.Scan(commandLogger, spaceanalyzer.ScanParams{
		RootPath:            spaceAnalyzerArgs.RootPath,
		MaxRecursion:        spaceAnalyzerArgs.MaxRecursion,
		CalculateFileHashes: spaceAnalyzerArgs.CalculateFileHashes,
	})
	if err != nil {
		return err
	}
	output := outputFile
	var bytesWritten int
	// TODO: add a option to enable streaming JSON and if not present do normal ish.
	streamingOutput := spaceAnalyzerArgs.OutputFormat == "sjson"
	if streamingOutput {
		commandLogger.Info("writing output as streaming JSON")
		jsonStreamer := streamingjson.NewLengthPrefixStreamJSONHandler[spaceanalyzer.FSEntity]()
		bw, err := WriteOutputAsStreamingJSON(cmd.Context(), info, output, jsonStreamer)
		bytesWritten += bw
		if err != nil {
			prettySize := util.GetPrettyBytesSize(int64(bytesWritten))
			commandLogger.Error("failed to write data to stdout", slog.Int("bytesWritten", bytesWritten), slog.String("prettyBytesWritten", prettySize), slog.String("errorMessage", err.Error()))
			return err
		}
	} else {
		commandLogger.Info("writing output as JSON")
		infoString, err := json.MarshalIndent(info, "", "  ")
		if err != nil {
			commandLogger.Error("failed to marshal data to JSON", slog.String("errorMessage", err.Error()))
			return err
		}
		bw, err := output.WriteString(string(infoString))
		bytesWritten += bw
		if err != nil {
			prettySize := util.GetPrettyBytesSize(int64(bytesWritten))
			commandLogger.Error("failed to write data to stdout", slog.Int("bytesWritten", bytesWritten), slog.String("prettyBytesWritten", prettySize), slog.String("errorMessage", err.Error()))
			return err
		}
	}
	prettySize := util.GetPrettyBytesSize(int64(bytesWritten))
	commandLogger.Info("finished writing output", slog.Int("bytesWritten", bytesWritten), slog.String("prettyBytesWritten", prettySize))
	return nil
}

func WriteOutputAsStreamingJSON(ctx context.Context, rootInfo spaceanalyzer.FSEntity, writer io.Writer, streamingHandler streamingjson.StreamingJSONWriter[spaceanalyzer.FSEntity]) (int, error) {
	var bytesWritten int
	var err error
	if rootInfo.IsDir {
		for _, i := range rootInfo.Children {
			bw, err := WriteOutputAsStreamingJSON(ctx, i, writer, streamingHandler)
			bytesWritten += bw
			if err != nil {
				return bytesWritten, err
			}
		}
		rootInfo.Children = nil
	}
	bw, err := streamingHandler.WriteObject(ctx, rootInfo, writer)
	bytesWritten += bw
	return bytesWritten, err
}
