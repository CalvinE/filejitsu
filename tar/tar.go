package tar

import (
	"archive/tar"
	"compress/gzip"
	"errors"
	"io"
	"io/fs"
	"log/slog"
	"os"
	"path/filepath"
	"strings"

	"github.com/calvine/filejitsu/encrypt"
	fgzip "github.com/calvine/filejitsu/gzip"
	"github.com/calvine/filejitsu/util"
)

const (
	DefaultPermission = 0754
)

type GZIPOptions struct {
	Header           gzip.Header
	CompressionLevel fgzip.GZipCompressionLevel
}

type EncryptionOptions struct {
	Passphrase []byte
}

type TarPackageParams struct {
	InputPaths        []string
	Output            io.Writer
	UseGzip           bool
	GZIPOptions       GZIPOptions
	UseEncryption     bool
	EncryptionOptions EncryptionOptions
}

type TarUnpackageParams struct {
	Input             io.Reader
	OutputPath        string
	UseGzip           bool
	UseEncryption     bool
	EncryptionOptions EncryptionOptions
}

func TarPackage(logger *slog.Logger, params TarPackageParams) error {
	logger.Debug("attempting to tar package the target path", slog.Any("params", params))
	out := params.Output
	// if use encryption then make encrypted writer
	if params.UseEncryption {
		logger.Debug("encryption enabled")
		encryptedOut, err := encrypt.NewAESEncryptionWriter(logger, out, params.EncryptionOptions.Passphrase)
		if err != nil {
			logger.Error("failed to create encrypted stream writer", slog.String("errorMessage", err.Error()))
			return err
		}
		out = encryptedOut
		defer func() {
			logger.Debug("closing encryption writer")
			if err := encryptedOut.Close(); err != nil {
				logger.Warn("encryption writer failed to close", slog.String("errorMessage", err.Error()))
			}
		}()
	}
	// if use gzip then make gzip writer
	if params.UseGzip {
		logger.Debug("gzip compression enabled", slog.Any("gzipOptions", params.GZIPOptions))
		compressionLevel, err := fgzip.GZipCompressionLevelToLevel(params.GZIPOptions.CompressionLevel)
		if err != nil {
			logger.Error("invalid gzip compression level provided", slog.String("gzipCompressionLevel", string(params.GZIPOptions.CompressionLevel)))
			return err
		}
		gzipOut, err := fgzip.NewGZIPWriter(logger, out, compressionLevel, gzip.Header{})
		if err != nil {
			logger.Error("failed to construct gzip writer", slog.String("errorMessage", err.Error()))
			return err
		}
		out = gzipOut
		defer func() {
			logger.Debug("closing gzip writer")
			if err := gzipOut.Flush(); err != nil {
				logger.Warn("gzip writer failed to flush", slog.String("errorMessage", err.Error()))
			}
			if err := gzipOut.Close(); err != nil {
				logger.Warn("gzip writer failed to close", slog.String("errorMessage", err.Error()))
			}
		}()
	}

	if len(params.InputPaths) == 0 {
		errMsg := "no input paths provided for tar archive"
		logger.Error(errMsg)
		return errors.New(errMsg)
	}

	// make the item to contain the tar data
	tarWriter := tar.NewWriter(out)
	defer func() {
		logger.Debug("closing tar writer")
		if err := tarWriter.Flush(); err != nil {
			logger.Warn("tar writer failed to flush", slog.String("errorMessage", err.Error()))
		}
		if err := tarWriter.Close(); err != nil {
			logger.Warn("tar writer failed to close", slog.String("errorMessage", err.Error()))
		}
	}()
	for _, ip := range params.InputPaths {
		logger.Info("processing input path", slog.String("path", ip))
		filepath.Walk(ip, func(path string, info fs.FileInfo, err error) (returnErr error) {
			walkLogger := logger.With(slog.String("path", path))
			if err != nil {
				walkLogger.Error("failed to walk entity",
					slog.String("errorMessage", err.Error()),
				)
				return err
			}
			fMode := info.Mode()
			isRegular := fMode.IsRegular()
			isDir := fMode.IsDir()
			name := info.Name()
			if !isRegular && !isDir {
				walkLogger.Debug("skipping entity because its not a regular file or directory")
				return nil
			}
			tarHeader, returnErr := tar.FileInfoHeader(info, name)
			if returnErr != nil {
				walkLogger.Error("failed to create tar header for file",
					slog.String("errorMessage", err.Error()),
				)
				return returnErr
			}

			tarHeader.Name = strings.TrimPrefix(strings.Replace(path, ip, "", -1), string(filepath.Separator))

			if tarHeader.Name == "" {
				if isRegular {
					walkLogger.Debug("got input path that is a file and not a directory, changing the header name to compensate")
					tarHeader.Name = filepath.Base(path)
				} else {
					return nil
				}
			}

			returnErr = tarWriter.WriteHeader(tarHeader)
			if returnErr != nil {
				walkLogger.Error("failed to write tar header for file", slog.String("errorMessage", err.Error()))
				return returnErr
			}

			if fMode.IsRegular() {
				logger.Debug("item is regular file, so writing file to tar package")
				f, returnErr := os.Open(path)
				if returnErr != nil {
					walkLogger.Error("failed to open file",
						slog.String("errorMessage", returnErr.Error()),
					)
					return returnErr
				}

				defer func() {
					returnErr = f.Close()
					if returnErr != nil {
						walkLogger.Error("failed to close file",
							slog.String("errorMessage", returnErr.Error()),
						)
					}
				}()

				bytesWritten, returnErr := io.Copy(tarWriter, f)
				logger.Debug("bytes written to tar writer", slog.Int64("bytesWritten", bytesWritten))
				if returnErr != nil {
					walkLogger.Error("failed to copy file to tar writer",
						slog.String("errorMessage", err.Error()),
					)
					return returnErr
				}
			}

			return nil
		})
	}
	if err := tarWriter.Close(); err != nil {
		logger.Warn("failed to close tar writer", slog.String("errorMessage", err.Error()))
	}
	return nil
}

func TarUnpackage(logger *slog.Logger, params TarUnpackageParams) error {
	in := params.Input

	if params.UseEncryption {
		logger.Debug("using decryption for tar unpack")
		decryptionReader, err := encrypt.NewAESDecryptionReader(logger, in, params.EncryptionOptions.Passphrase)
		if err != nil {
			logger.Error("failed to create decryption reader", slog.String("errorMessage", err.Error()))
			return err
		}
		in = decryptionReader
	}
	if params.UseGzip {
		logger.Debug("using gzip for tar unpack")
		gzipReader, _, err := fgzip.NewGZIPReader(logger, in)
		if err != nil {
			logger.Error("failed to create gzip reader", slog.String("errorMessage", err.Error()))
			return err
		}
		in = gzipReader
		defer func() {
			logger.Debug("closing gzip reader")
			if err := gzipReader.Close(); err != nil {
				logger.Warn("gzip reader failed to close", slog.String("errorMessage", err.Error()))
			}
		}()
	}

	tarReader := tar.NewReader(in)
	if err := util.MakeAllDirIfNotExists(logger, params.OutputPath, DefaultPermission); err != nil {
		logger.Error("failed to create output directory", slog.String("outputPath", params.OutputPath), slog.String("errorMessage", err.Error()))
	}
	// info, err := os.Lstat(params.OutputPath)
	// if err != nil {
	// 	if os.IsNotExist(err) {
	// 		logger.Debug("output path does not exist, attempting to create it", slog.String("outputPath", params.OutputPath))
	// 		err = os.MkdirAll(params.OutputPath, DefaultPermission)
	// 		if err != nil {
	// 			logger.Error("failed to make output directory", slog.String("outputPath", params.OutputPath), slog.String("errorMessage", err.Error()))
	// 			return fmt.Errorf("failed to create output directory: %w", err)
	// 		}
	// 	} else {
	// 		logger.Error("failed to get info on output path", slog.String("outputPath", params.OutputPath), slog.String("errorMessage", err.Error()))
	// 		return fmt.Errorf("failed to create output directory: %w", err)
	// 	}
	// } else if !info.IsDir() {
	// 	errMsg := "output path exists and is not a directory"
	// 	logger.Error(errMsg, slog.String("outputPath", params.OutputPath))
	// 	return errors.New(errMsg)
	// }
	numFiles := 0
	for {
		nextHeader, err := tarReader.Next()
		switch {
		case err == io.EOF:
			logger.Debug("finished reading tar file", slog.Int("numFiles", numFiles))
			return nil
		case err != nil:
			logger.Error("failed to read next from tar package", slog.String("errorMessage", err.Error()))
			return err
		case nextHeader == nil:
			logger.Warn("encountered nil tar header... continuing...")
			continue
		}
		numFiles++
		target := filepath.Join(params.OutputPath, nextHeader.Name)
		logger.Debug("starting to unpackage item", slog.String("target", target))
		switch nextHeader.Typeflag {
		// if its a dir and it doesn't exist create it
		case tar.TypeDir:
			logger.Debug("got directory from tar", slog.String("target", target))
			if _, err := os.Stat(target); err != nil {
				if err := os.MkdirAll(target, DefaultPermission); err != nil {
					logger.Error("failed to make target directory", slog.String("target", target))
					return err
				}
			}

		// if it's a file create it
		case tar.TypeReg:
			logger.Debug("got regular file from tar", slog.String("target", target))
			pathToFile := filepath.Dir(target)
			if err := util.MakeAllDirIfNotExists(logger, pathToFile, DefaultPermission); err != nil {
				logger.Error("failed to create directory ")
			}
			f, err := os.OpenFile(target, os.O_CREATE|os.O_RDWR, os.FileMode(nextHeader.Mode))
			if err != nil {
				logger.Error("failed to open target file for unpackaging", slog.String("target", target))
				return err
			}

			// copy over contents
			bytesWritten, err := io.Copy(f, tarReader)
			logger.Debug("bytes written to output file", slog.String("target", target), slog.Int64("bytesWritten", bytesWritten))
			if err != nil {
				logger.Error("failed to write tar data to output file", slog.String("target", target), slog.String("errorMessage", err.Error()))
				return err
			}

			// manually close here after each file operation; defering would cause each file close
			// to wait until all operations have completed.
			if err := f.Close(); err != nil {
				logger.Warn("failed to close target file", slog.String("target", target), slog.String("errorMessage", err.Error()))
			}
		}
	}
}
