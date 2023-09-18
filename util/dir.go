package util

import (
	"crypto/sha512"
	"encoding/hex"
	"errors"
	"io/fs"
	"os"
	"path"
	"path/filepath"

	"github.com/google/uuid"
	"golang.org/x/exp/slog"
)

func GetDirContentDetails(logger *slog.Logger, currentPath, currentID string, calculateFileHashes bool, maxRecursion, recursionCount int) (FSEntity, error) {
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
	rootID := currentID
	if len(rootID) == 0 {
		logger.Debug("creating new id because one provided was blank")
		rootID = uuid.New().String()
	}
	rootEntity := FileInfoToFSEntry(logger, currentStat, rootID, currentPath, calculateFileHashes)
	if recursionCount == 0 {
		// TODO: This is a hack until I feel like working out the issue here. first run of this adds the name to the path again, so removing the last bit for first pass only...
		rootEntity.FullPath = currentPath
	}
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
				dInfo := DirInfoToFSEntry(e, rootID, currentPath)
				if maxRecursion != -1 && maxRecursion > recursionCount {
					logger.Debug("skipping call to get dir children info due to max recursion setting", slog.Int("maxRecursion", maxRecursion), slog.Int("recursionCount", recursionCount))
				} else {
					// populate children
					newPath := path.Join(currentPath, dInfo.Name)
					subDInfo, err := GetDirContentDetails(logger, newPath, dInfo.ID, calculateFileHashes, maxRecursion, recursionCount+1)
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
				fInfo := FileInfoToFSEntry(logger, fileStat, rootID, currentPath, calculateFileHashes)
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

func FileInfoToFSEntry(logger *slog.Logger, fi fs.FileInfo, parentID, ePath string, calculateFilePath bool) FSEntity {
	hasher := sha512.New()
	id := uuid.New().String()
	name := fi.Name()
	mode := fi.Mode()
	eType := mode.Type()
	isDir := fi.IsDir()
	isRegular := eType.IsRegular()
	size := int64(0)
	lastModified := fi.ModTime()
	extension := ""
	fileHash := ""
	fullPath := path.Join(ePath, name)
	if isRegular {
		size = fi.Size()
		extension = path.Ext(name)
		data, err := os.ReadFile(fullPath)
		if err != nil {
			logger.Warn("failed to calculate file hash", slog.String("fullPath", fullPath), slog.String("errorMessage", err.Error()))
		} else {
			hasher.Write(data)
			fileHash = hex.EncodeToString(hasher.Sum(nil))
			hasher.Reset()
		}
	}
	permissions := fi.Mode().Perm()
	entityType := getEntityType(isDir, isRegular)

	e := FSEntity{
		Name:         name,
		Size:         size,
		FullPath:     fullPath,
		IsDir:        isDir,
		EntityType:   entityType,
		Extension:    extension,
		FileHash:     fileHash,
		Mode:         uint32(mode),
		Type:         uint32(eType),
		Permissions:  uint32(permissions),
		LastModified: lastModified,
		ParentID:     parentID,
		ID:           id,
	}
	return e
}

func DirInfoToFSEntry(di fs.DirEntry, parentID, ePath string) FSEntity {
	eType := di.Type()
	permissions := eType.Perm()
	name := di.Name()
	fullPath := path.Join(ePath, name)
	isDir := di.IsDir()
	isRegular := eType.IsRegular()
	entityType := getEntityType(isDir, isRegular)
	e := FSEntity{
		ID:          uuid.New().String(),
		ParentID:    parentID,
		Name:        di.Name(),
		IsDir:       isDir,
		EntityType:  entityType,
		FullPath:    fullPath,
		Type:        uint32(eType),
		Permissions: uint32(permissions),
	}
	return e
}

func getEntityType(isDir, isRegular bool) EntityType {
	if isDir {
		return DirectoryType
	} else if isRegular {
		return FileType
	}
	return OtherType
}
