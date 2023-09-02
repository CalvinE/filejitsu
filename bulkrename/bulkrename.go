package bulkrename

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"os"
	"path"
	"path/filepath"
	"regexp"
	"text/template"

	"github.com/calvine/filejitsu/util"
	"golang.org/x/exp/slog"
)

func CalculateJobs(ctx context.Context, logger *slog.Logger, params Params) ([]ResultEntry, error) {
	recursionCount := util.GetContextRecursionCount(ctx)
	logger = logger.With("recursionCount", recursionCount)
	results := make([]ResultEntry, 0)
	absPath, err := filepath.Abs(params.RootPath)
	if err != nil {
		logger.Error("failed to get absolute path for root path... terminating run",
			slog.String("rootPath", params.RootPath),
			slog.String("error", err.Error()),
		)
		return nil, err
	}
	// enumerate root dir
	content, err := os.ReadDir(params.RootPath)
	if err != nil {
		logger.Error("failed to enumerate root directory",
			slog.String("error", err.Error()),
			slog.String("rootDir", params.RootPath),
		)
		return nil, fmt.Errorf("failed to enumerate root dir: %v", err)
	}
	// look at root dir contents and try to find files that match the target regex
	for _, c := range content {
		i, err := c.Info()
		iName := i.Name()
		if err != nil {
			logger.Error("failed to get info on item... terminating run",
				slog.String("fileName", c.Name()),
				slog.String("error", err.Error()),
			)
			return nil, err
		}
		if !i.IsDir() {
			fName := iName
			matchesTarget := params.TargetRegex.Match([]byte(fName))
			if !matchesTarget {
				logger.Debug("file name does not match target regex",
					slog.String("fileName", fName),
				)
				// file does not match target regex so we can move forward
				continue
			}
			originalFile := util.File{
				Name:      fName,
				Size:      i.Size(),
				Extension: path.Ext(fName),
				Path:      absPath, // may want to convert to absolute path?
			}
			// TODO: process file name to make new file name
			result, err := calculateRename(logger, originalFile, params.TargetRegex, params.DestinationTemplate)
			if err != nil {
				logger.Error("failed to calculate rename for file",
					slog.Any("file", originalFile),
					slog.String("error", err.Error()),
				)
				return nil, err
			}
			results = append(results, result)
		} else if params.Recursive {
			// it is a dir and we may need to recurse
			recursiveParams := params
			recursiveParams.RootPath = path.Join(absPath, iName)
			logger.Info("directory found and recursion is activated. performing recursive dir scan",
				slog.String("newRootPath", recursiveParams.RootPath),
				slog.Int("recursionCount", recursionCount),
			)
			ctx = util.SetContextRecursionCount(ctx, recursionCount+1)
			recursiveResults, err := CalculateJobs(ctx, logger, recursiveParams)
			if err != nil {
				logger.Error("failed in recursive run",
					slog.Any("recursiveParams", recursiveParams),
					slog.String("error", err.Error()),
				)
				return nil, err
			}
			results = append(results, recursiveResults...)
		}
	}

	return results, nil
}

func calculateRename(logger *slog.Logger, original util.File, captureRegex *regexp.Regexp, template *template.Template) (ResultEntry, error) {
	valueMap := make(map[string]string)
	subMatches := captureRegex.FindStringSubmatch(original.Name)
	numSubMatches := len(subMatches)
	if numSubMatches == 0 {
		err := errors.New("no matches found for file name")
		logger.Error(err.Error(),
			slog.String("fileName", original.Name),
			slog.String("regex", captureRegex.String()),
		)
		return ResultEntry{}, err
	}
	numRegexSubExpressions := captureRegex.NumSubexp()
	if numSubMatches != numRegexSubExpressions+1 {
		err := errors.New("number of regex sub expressions not match sub matches + 1")
		logger.Error(err.Error(),
			slog.Int("numRegexSubExpressions", numRegexSubExpressions),
			slog.Int("numSubMatches", numSubMatches),
		)
		return ResultEntry{}, err
	}
	for i, s := range captureRegex.SubexpNames() {
		logger.Debug("setting value map item",
			slog.String("key", s),
			slog.String("value", subMatches[i]),
		)
		valueMap[s] = subMatches[i]
	}
	var buffer bytes.Buffer
	if err := template.Execute(&buffer, valueMap); err != nil {
		logger.Error("failed to execute template with captured data",
			slog.Any("data", valueMap),
			slog.String("errorMessage", err.Error()),
		)
		return ResultEntry{}, err
	}
	newFileName := captureRegex.ReplaceAll([]byte(original.Name), buffer.Bytes())
	logger.Debug("new file name created",
		slog.String("originalFileName", original.Name),
		slog.String("newFileName", string(newFileName)),
	)
	newFile := util.File{
		Name:      string(newFileName),
		Path:      original.Path,
		Extension: path.Ext(string(newFileName)),
		Size:      original.Size,
	}
	didChange := original.Name != newFile.Name
	return ResultEntry{
		Original:  original,
		New:       newFile,
		DidChange: didChange,
	}, nil

}

func ProcessResult(logger *slog.Logger, result ResultEntry) error {
	origPath := path.Join(result.Original.Path, result.Original.Name)
	newPath := path.Join(result.New.Path, result.New.Name)
	if err := os.Rename(origPath, newPath); err != nil {
		if errors.Is(err, os.ErrNotExist) {
			logger.Warn("file does not exist... could the file have already been renamed?",
				slog.String("file", origPath),
			)
		}
		logger.Error("rename operation encountered error",
			slog.String("originalFile", origPath),
			slog.String("newFile", newPath),
			slog.String("errorMessage", err.Error()),
		)
		return err
	}
	return nil
}

func ProcessTestResults(logger *slog.Logger, r ResultEntry) error {
	fmt.Printf("%s => %s\n\n", r.Original.Name, r.New.Name)
	return nil
}
