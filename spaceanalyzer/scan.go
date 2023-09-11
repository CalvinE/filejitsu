package spaceanalyzer

import (
	"github.com/calvine/filejitsu/util"
	"golang.org/x/exp/slog"
)

func Scan(logger *slog.Logger, params ScanParams) (util.FSEntity, error) {
	info, err := util.GetDirContentDetails(logger, params.RootPath, "", params.CalculateFileHashes, params.MaxRecursion, 0)
	if err != nil {
		logger.Error("failed to get dir content details", slog.String("errorMessage", err.Error()), slog.String("rootPath", params.RootPath))
		return util.FSEntity{}, err
	}
	return info, nil
}
