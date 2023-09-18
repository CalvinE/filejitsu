package spaceanalyzer

import (
	"fmt"

	"github.com/calvine/filejitsu/util"
	"golang.org/x/exp/slog"
)

const (
	KB = 1 << 10
	MB = 1 << (10 * (1 + iota))
	GB
	TB
	PB
	EB
)

func populateExtraSizeInfo(item *util.FSEntity) {
	var size int64 = 0
	if len(item.Children) > 0 {
		for index, childItem := range item.Children {
			populateExtraSizeInfo(&childItem)
			item.Children[index] = childItem
			size += childItem.Size
		}
	} else {
		size += item.Size
	}
	item.Size = size
	var divSize float64
	var unit string
	if item.Size >= EB {
		divSize = float64(item.Size) / float64(EB)
		unit = "EB"
	} else if item.Size >= PB {
		divSize = float64(item.Size) / float64(PB)
		unit = "PB"
	} else if item.Size >= TB {
		divSize = float64(item.Size) / float64(TB)
		unit = "TB"
	} else if item.Size >= GB {
		divSize = float64(item.Size) / float64(GB)
		unit = "GB"
	} else if item.Size >= MB {
		divSize = float64(item.Size) / float64(MB)
		unit = "MB"
	} else if item.Size >= KB {
		divSize = float64(item.Size) / float64(KB)
		unit = "KB"
	} else {
		divSize = float64(item.Size)
		unit = "B"
	}
	item.PrettySize = fmt.Sprintf("%.2f %s", divSize, unit)
}

func Scan(logger *slog.Logger, params ScanParams) (util.FSEntity, error) {
	info, err := util.GetDirContentDetails(logger, params.RootPath, "", params.CalculateFileHashes, params.MaxRecursion, 0)
	if err != nil {
		logger.Error("failed to get dir content details", slog.String("errorMessage", err.Error()), slog.String("rootPath", params.RootPath))
		return util.FSEntity{}, err
	}
	populateExtraSizeInfo(&info)
	return info, nil
}
