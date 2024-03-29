package spaceanalyzer

import (
	"errors"
	"os"
	"path/filepath"

	"log/slog"

	"github.com/google/uuid"
)

// THIS WHOLE FILE IS DEPRECATED... Keeping for posterity... for now

type NonConcurrentFSSCanner interface {
	Scan(logger *slog.Logger, currentPath, parentID, id string, calculateFileHashes bool, maxRecursion, recursionCount int) (FSEntity, error)
}

type nonConcurrentFSScanner struct {
}

func NewNonConcurrentFSScanner() NonConcurrentFSSCanner {
	return &nonConcurrentFSScanner{}
}

func (ncs nonConcurrentFSScanner) Scan(logger *slog.Logger, currentPath, parentID, id string, calculateFileHashes bool, maxRecursion, recursionCount int) (FSEntity, error) {
	logger.Debug("attempting to list contents of provided path", slog.String("currentPath", currentPath), slog.Int("recursionCount", recursionCount))
	if !filepath.IsAbs(currentPath) {
		var err error
		currentPath, err = filepath.Abs(currentPath)
		if err != nil {
			logger.Error("failed to get absolute path for non absolute input", slog.String("currentPath", currentPath), slog.String("errorMessage", err.Error()))
			return FSEntity{}, err
		}
	}
	currentStat, err := os.Stat(currentPath)
	if err != nil {
		logger.Error("failed to get stat on current path", slog.String("errorMessage", err.Error()))
		return FSEntity{}, err
	}
	if len(id) == 0 {
		id = uuid.New().String()
	}
	rootEntity := FileInfoToFSEntry(logger, currentStat, parentID, id, currentPath, calculateFileHashes, recursionCount)
	if recursionCount == 0 {
		// TODO: This is a hack until I feel like working out the issue here. first run of this adds the name to the path again, so removing the last bit for first pass only...
		rootEntity.FullPath = currentPath
	}
	logger.Info("processing file", slog.String("path", rootEntity.FullPath))
	if rootEntity.EntityType == DirectoryType {
		entries, err := os.ReadDir(currentPath)
		if err != nil {
			logger.Error("failed to read directory", slog.String("errorMessage", err.Error()))
			return FSEntity{}, err
		}
		numItems := len(entries)
		logger.Debug("directory contents read", slog.Int("numItems", numItems))
		dirContents := make([]FSEntity, 0, numItems)
		for _, e := range entries {
			name := e.Name()
			isDir := e.IsDir()
			eType := e.Type()
			isRegular := eType.IsRegular()
			entityType := getEntityType(isDir, isRegular)
			entry := FSEntity{
				Name:       name,
				IsDir:      isDir,
				EntityType: entityType,
				Type:       uint32(eType),
			}
			if isDir {
				logger.Debug("child item is dir", slog.String("name", name))
				dInfo := DirInfoToFSEntry(e, parentID, id, currentPath)
				if maxRecursion != -1 && maxRecursion > recursionCount {
					logger.Debug("skipping call to get dir children info due to max recursion setting", slog.Int("maxRecursion", maxRecursion), slog.Int("recursionCount", recursionCount))
				} else {
					// populate children
					newPath := filepath.Join(currentPath, dInfo.Name)
					childID := uuid.New().String()
					subDInfo, err := ncs.Scan(logger, newPath, dInfo.ID, childID, calculateFileHashes, maxRecursion, recursionCount+1)
					if err != nil {
						logger.Error("failed to get details for child directory", slog.String("path", newPath), slog.String("errorMessage", err.Error()))
						return FSEntity{}, err
					}
					numChildren := len(subDInfo.Children)
					logger.Debug("finished processing sub directory", slog.Int("numChildren", numChildren))
					if numChildren > 0 {
						dInfo.Children = make([]FSEntity, 0, numChildren)
						dInfo.Children = append(dInfo.Children, subDInfo.Children...)
					}
				}
				dirContents = append(dirContents, dInfo)
			} else if isRegular {
				logger.Debug("child item is regular file", slog.String("fileName", name))
				fileStat, err := e.Info()
				if err != nil {
					logger.Error("failed to get stat on file", slog.Any("file", entry))
					return FSEntity{}, err
				}

				fInfo := FileInfoToFSEntry(logger, fileStat, parentID, id, currentPath, calculateFileHashes, recursionCount)
				dirContents = append(dirContents, fInfo)
			} else {
				logger.Debug("found non regular file / dir... skipping...", slog.Any("file", entry))
				continue
			}

			// logger.Debug("child items is not a dir", slog.String("name", name))
		}
		rootEntity.Children = dirContents
		return rootEntity, nil
	}
	logger.Error("path provided was not a directory")
	return FSEntity{}, errors.New("path provided was not a directory")
}
