package spaceanalyzer

import (
	"fmt"
	"io/fs"
	"os"
	"path"

	"github.com/calvine/filejitsu/util"
	"golang.org/x/exp/slog"
)

func populateExtraSizeInfo(item *FSEntity) {
	if len(item.Children) > 0 {
		var size int64
		for index, childItem := range item.Children {
			populateExtraSizeInfo(&childItem)
			item.Children[index] = childItem
			size += childItem.Size
		}
		item.Size = size
	}
	item.PrettySize = util.GetPrettyBytesSize(int64(item.Size))
}

func Scan(logger *slog.Logger, params ScanParams) (FSEntity, error) {
	// ncs := NewNonConcurrentFSScanner()
	// TODO: parameterize concurrency limit
	cfs := NewConcurrentFSScanner(0)
	info, err := cfs.Scan(logger, params.RootPath, "base", params.CalculateFileHashes, params.MaxRecursion)
	logger.Info("finished scan")
	if err != nil {
		logger.Error("failed to get dir content details", slog.String("errorMessage", err.Error()), slog.String("rootPath", params.RootPath))
		return FSEntity{}, err
	}
	logger.Info("populating extra size info")
	populateExtraSizeInfo(&info)
	logger.Info("finished populating extra size info")
	return info, nil
}

func FileInfoToFSEntry(logger *slog.Logger, fi fs.FileInfo, parentID, id, fullPath string, shouldCalculateFileHash bool, depth int) FSEntity {
	name := fi.Name()
	mode := fi.Mode()
	eType := mode.Type()
	isDir := fi.IsDir()
	isRegular := eType.IsRegular()
	size := int64(0)
	lastModified := fi.ModTime()
	extension := ""
	fileHash := ""
	if isRegular {
		size = fi.Size()
		extension = path.Ext(name)
		if shouldCalculateFileHash {
			var err error
			fileHash, err = calculateFileHash(logger, fullPath)
			if err != nil {
				logger.Warn("failed to calculate file hash", slog.Any("fullPath", fullPath), slog.String("errorMessage", err.Error()))
			}
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
		Depth:        depth,
	}
	return e
}

func calculateFileHash(logger *slog.Logger, fullPath string) (string, error) {
	var fileHash string
	fd, err := os.Open(fullPath)
	if err != nil {
		// logger.Warn("failed to open file for hashing", slog.String("fullPath", fullPath), slog.String("errorMessage", err.Error()))
		return fileHash, fmt.Errorf("failed to open file for hashing: %w", err)
	} else {
		fileHash, err = util.Sha512HashData(logger, fd)
		if err != nil {
			// logger.Warn("failed to calculate file hash after open", slog.String("fullPath", fullPath), slog.String("errorMessage", err.Error()))
			return fileHash, fmt.Errorf("failed to calculate file hash after open: %w", err)
		}
	}
	err = fd.Close()
	if err != nil {
		// logger.Error("failed to close file", slog.String("fullPath", fullPath), slog.String("errorMessage", err.Error()))
		return fileHash, fmt.Errorf("failed to close file: %w", err)
	}
	return fileHash, nil
}

func DirInfoToFSEntry(di fs.DirEntry, parentID, id, fullPath string) FSEntity {
	eType := di.Type()
	permissions := eType.Perm()
	name := di.Name()
	isDir := di.IsDir()
	isRegular := eType.IsRegular()
	entityType := getEntityType(isDir, isRegular)
	e := FSEntity{
		ID:          id,
		ParentID:    parentID,
		Name:        name,
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
