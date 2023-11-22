package tar

import (
	"compress/gzip"
	"io"

	fgzip "github.com/calvine/filejitsu/gzip"
)

type GZIPOptions struct {
	Header           gzip.Header                `json:"header"`
	CompressionLevel fgzip.GZipCompressionLevel `json:"compressionLevel"`
}

type EncryptionOptions struct {
}

type TarPackageParams struct {
	TargetPath    string
	Output        io.Writer
	UseGzip       bool `json:"useGzip"`
	UseEncryption bool `json:"useEncryption"`
}

type TarPackageOutput struct {
}

// func makeTarWriter(logger *slog.Logger, params TarPackageParams) (io.Writer, error) {
// 	// make the base file for writing the data to.
// 	outputStat, err := os.Lstat(params.OutputPath)
// 	if err != nil {
// 		logger.Error("failed to get info on output path", slog.String("errorMessage", err.Error()))
// 	}
// 	outputMode := outputStat.Mode()
// 	if !outputMode.IsRegular() {

// 	}
// 	// if use encryption then make encrypted writer

// 	// if use gzip then make gzip writer

// 	return nil, nil

// }

// func TarPackage(logger *slog.Logger, params TarPackageParams) (TarPackageOutput error) {
// 	logger.Debug("attempting to tar package the target path", slog.Any("params", params))
// 	// make the item to contain the tar data
// 	tarWriter, err := makeTarWriter(logger, params)
// 	if err != nil {
// 		logger.Error("failed to make tar writer", slog.String("errorMessage", err.Error()))
// 		return TarPackageOutput{}, err
// 	}
// 	//
// }
