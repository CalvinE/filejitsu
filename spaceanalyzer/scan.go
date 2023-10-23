package spaceanalyzer

import (
	"io/fs"
	"os"
	"path"

	"github.com/calvine/filejitsu/util"
	"github.com/google/uuid"
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
	cfs := NewConcurrentFSScanner()
	info, err := cfs.Scan(logger, params.RootPath, "base", params.CalculateFileHashes, params.MaxRecursion)
	if err != nil {
		logger.Error("failed to get dir content details", slog.String("errorMessage", err.Error()), slog.String("rootPath", params.RootPath))
		return FSEntity{}, err
	}
	logger.Debug("populating extra size info recursively")
	populateExtraSizeInfo(&info)
	logger.Debug("finished populating extra size info recursively")
	return info, nil
}

func FileInfoToFSEntry(logger *slog.Logger, fi fs.FileInfo, parentID, ePath string, shouldCalculateFileHash bool) FSEntity {
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
		if shouldCalculateFileHash {
			fileHash = calculateFileHash(logger, fullPath, fileHash)
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

func calculateFileHash(logger *slog.Logger, fullPath, fileHash string) string {
	fd, err := os.Open(fullPath)
	if err != nil {
		logger.Warn("failed to calculate file hash", slog.String("fullPath", fullPath), slog.String("errorMessage", err.Error()))
	} else {
		fileHash, err = util.Sha512HashData(logger, fd)
		if err != nil {
			logger.Warn("failed to calculate file hash after open", slog.String("fullPath", fullPath), slog.String("errorMessage", err.Error()))
		}
	}
	err = fd.Close()
	if err != nil {
		logger.Error("failed to close file", slog.String("fullPath", fullPath), slog.String("errorMessage", err.Error()))
	}
	return fileHash
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
