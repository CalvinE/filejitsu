package spaceanalyzer

import (
	"errors"
	"os"
	"path"
	"path/filepath"
	"sync"

	"github.com/google/uuid"
	"golang.org/x/exp/slog"
)

const (
	concurrencyLimit = 16
)

var (
	ErrRecursionLimit = errors.New("recursion limit")
	emptyFSJob        = FSJob{}
)

type ConcurrentFSScanner interface {
	Scan(logger *slog.Logger, entityPath, rootID string, calculateFileHashes bool, maxRecursion int) (FSEntity, error)
}

type concurrentFSScanner struct {
}

func NewConcurrentFSScanner() ConcurrentFSScanner {
	return &concurrentFSScanner{}
}

func (cfs *concurrentFSScanner) Scan(logger *slog.Logger, entityPath, rootID string, shouldCalculateFileHashes bool, maxRecursion int) (FSEntity, error) {
	files := make(map[string][]FSEntity)
	dirs := make(map[string]FSEntity)
	wg := sync.WaitGroup{}
	if len(rootID) == 0 {
		logger.Debug("creating new id because one provided was blank")
		rootID = "root"
	}
	jobsChan := enumerateScanTargets(logger, &wg, entityPath, "", rootID, maxRecursion)
	mutex := sync.Mutex{}
	go func() {
		working := true
		for working {
			// TODO: go routine this
			// TODO: mutex log the map write...
			j, ok := <-jobsChan
			wg.Add(1)
			if !ok {
				// TODO: if the item is not empty process it...
				working = false
				logger.Warn("jobs channel closed")
				if j == emptyFSJob {
					wg.Done()
					break
				}
			}
			go func() {
				if j.IsDir {
					dir := FileInfoToFSEntry(logger, j.Info, j.ParentID, j.ID, j.FullPath, shouldCalculateFileHashes)
					logger.Debug("received dir job", slog.Any("dirEntity", dir))
					mutex.Lock()
					dirs[dir.ID] = dir
					mutex.Unlock()
				} else {
					file := FileInfoToFSEntry(logger, j.Info, j.ParentID, j.ID, j.FullPath, shouldCalculateFileHashes)
					mutex.Lock()
					files[j.ParentID] = append(files[j.ParentID], file)
					mutex.Unlock()
				}
				wg.Done()
			}()
		}
	}()
	wg.Wait()
	rootEntity := dirs[rootID]
	delete(dirs, rootID)
	entity := collateEntities(rootEntity, files, dirs)

	// logger.Debug("running concurrent scan", slog.Any("fsJobs", jobsChan))
	// entity, err := processJobs(logger, jobsChan, shouldCalculateFileHashes)
	// entity, err := cfs.scan(logger, entityPath, rootID, shouldCalculateFileHashes, maxRecursion, 0)
	// if err != nil {
	// 	logger.Error("failed to scan entity", slog.String("errorMessage", err.Error()))
	// 	return FSEntity{}, err
	// }
	logger.Debug("concurrent scan complete")
	return entity, nil
}

func collateEntities(entity FSEntity, files map[string][]FSEntity, dirs map[string]FSEntity) FSEntity {
	// add files
	entityFiles := files[entity.ID]
	if len(entityFiles) > 0 {
		delete(files, entity.ID)
		entity.Children = append(entity.Children, entityFiles...)
	}
	// add directories
	for _, v := range dirs {
		if v.ParentID == entity.ID {
			childDir := collateEntities(v, files, dirs)
			entityFiles = append(entityFiles, childDir)
		}
	}
	entity.Children = append(entity.Children, entityFiles...)
	return entity
}

// func processJobs(logger *slog.Logger, job FSJob, shouldCalculateFileHashes bool) (FSEntity, error) {
// }

func enumerateScanTargets(logger *slog.Logger, wg *sync.WaitGroup, entityPath, parentID, id string, maxRecursion int) <-chan FSJob {
	jobsChan := make(chan FSJob, concurrencyLimit)
	wg.Add(1)
	go func() {
		err := recursiveEnumerateScanTargets(logger, jobsChan, entityPath, parentID, id, maxRecursion, 0)
		if err != nil {
			// TODO: handle recursion error
			logger.Error("failed to enumerate fully", slog.String("errorMessage", err.Error()))
		}
		// close(jobsChan)
		wg.Done()
	}()
	return jobsChan
}

func recursiveEnumerateScanTargets(logger *slog.Logger, jobsChan chan<- FSJob, entityPath, parentID, id string, maxRecursion, recursionCount int) error {
	logger = logger.With(slog.String("parentID", parentID), slog.String("entityPath", entityPath))
	currentPath := entityPath
	var job FSJob
	if !filepath.IsAbs(currentPath) {
		var err error
		currentPath, err = filepath.Abs(currentPath)
		if err != nil {
			logger.Error("failed to get absolute path for non absolute input", slog.String("entityPath", entityPath), slog.String("errorMessage", err.Error()))
			return err
		}
	}
	if maxRecursion > -1 && maxRecursion >= recursionCount {
		logger.Warn("skipping call to enumerate entity due to max recursion setting", slog.Int("maxRecursion", maxRecursion), slog.Int("recursionCount", recursionCount), slog.String("currentPath", currentPath))
		return ErrRecursionLimit
	}
	currentStat, err := os.Stat(currentPath)
	if err != nil {
		logger.Error("failed to get stat on current path", slog.String("errorMessage", err.Error()))
		return err
	}
	job.ID = id
	if len(job.ID) == 0 {
		job.ID = uuid.New().String()
	}
	job.ParentID = parentID
	job.Info = currentStat
	job.FullPath = currentPath
	logger = logger.With("id", id)
	if currentStat.IsDir() {
		job.IsDir = true
		dirContents, err := os.ReadDir(currentPath)
		if err != nil {
			logger.Error("failed to get contents of directory", slog.String("errorMessage", err.Error()))
		}
		numChildren := len(dirContents)
		logger.Debug("dir contents retrieved", slog.Int("numChildren", numChildren))
		for _, d := range dirContents {
			childName := d.Name()
			childPath := path.Join(currentPath, childName)
			childID := uuid.New().String()
			err := recursiveEnumerateScanTargets(logger, jobsChan, childPath, job.ID, childID, maxRecursion, recursionCount+1)
			if err != nil {
				// Should we die here or return that there is an error?
				logger.Error("failed to enumerate child entity", slog.String("childPath", childPath), slog.String("errorMessage", err.Error()))
				return err
			}
		}
	}
	jobsChan <- job
	return nil
}

// func (cfs *concurrentFSScanner) scan(logger *slog.Logger, entityPath, parentID string, shouldCalculateFileHashes bool, maxRecursion, recursionCount int) (FSEntity, error) {
// 	// wg := sync.WaitGroup{}
// 	logger = logger.With(slog.String("parentID", parentID))
// 	currentPath := entityPath
// 	if !filepath.IsAbs(currentPath) {
// 		var err error
// 		currentPath, err = filepath.Abs(currentPath)
// 		if err != nil {
// 			logger.Error("failed to get absolute path for non absolute input", slog.String("entityPath", entityPath), slog.String("errorMessage", err.Error()))
// 			return FSEntity{}, err
// 		}
// 	}
// 	currentStat, err := os.Stat(currentPath)
// 	if err != nil {
// 		logger.Error("failed to get stat on current path", slog.String("errorMessage", err.Error()))
// 		return FSEntity{}, err
// 	}
// 	entity := FileInfoToFSEntry(logger, currentStat, parentID, currentPath, shouldCalculateFileHashes)
// 	logger = logger.With("id", entity.ID)
// 	// if recursionCount == 0 {
// 	// TODO: This is a hack until I feel like working out the issue here. first run of this adds the name to the path again, so removing the last bit for first pass only...
// 	entity.FullPath = currentPath
// 	// }
// 	logger.Info("processing file", slog.String("path", entity.FullPath))
// 	if entity.EntityType == DirectoryType {
// 		logger.Debug("entity is a directory", slog.Any("entity", entity))
// 		fsChan := make(chan FSEntity, 10)
// 		errorChan := make(chan error, 10)
// 		defer func() {
// 			close(fsChan)
// 			close(errorChan)
// 		}()
// 		entries, err := os.ReadDir(currentPath)
// 		if err != nil {
// 			logger.Error("failed to read directory", slog.String("errorMessage", err.Error()))
// 			return FSEntity{}, err
// 		}
// 		numItems := len(entries)
// 		logger.Debug("directory contents read", slog.Int("numItems", numItems))
// 		dirContents := make([]FSEntity, 0, numItems)
// 		for _, e := range entries {
// 			name := e.Name()
// 			isDir := e.IsDir()
// 			eType := e.Type()
// 			isRegular := eType.IsRegular()
// 			entityType := getEntityType(isDir, isRegular)
// 			entry := FSEntity{
// 				Name:       name,
// 				IsDir:      isDir,
// 				EntityType: entityType,
// 				Type:       uint32(eType),
// 			}
// 			if isDir {
// 				logger.Debug("child item is dir", slog.String("name", name))
// 				dInfo := DirInfoToFSEntry(e, parentID, currentPath)
// 				if maxRecursion != -1 && maxRecursion > recursionCount {
// 					logger.Debug("skipping call to get dir children info due to max recursion setting", slog.Int("maxRecursion", maxRecursion), slog.Int("recursionCount", recursionCount))
// 				} else {
// 					// populate children
// 					newPath := path.Join(currentPath, dInfo.Name)
// 					// wg.Add(1)
// 					go func() {
// 						subDInfo, err := cfs.scan(logger, newPath, dInfo.ID, shouldCalculateFileHashes, maxRecursion, recursionCount+1)
// 						if err != nil {
// 							logger.Error("failed to scan child directory", slog.Any("directoryInfo", dInfo), slog.String("errorMessage", err.Error()))
// 							errorChan <- err
// 							return
// 						}
// 						numChildren := len(subDInfo.Children)
// 						logger.Debug("finished processing sub directory", slog.Int("numChildren", numChildren))
// 						if numChildren > 0 {
// 							dInfo.Children = make([]FSEntity, 0, numChildren)
// 							dInfo.Children = append(dInfo.Children, subDInfo.Children...)
// 						}
// 						// dirContents = append(dirContents, dInfo)
// 						fsChan <- dInfo
// 						// wg.Done()
// 					}()
// 				}
// 				// fsChan <- dInfo
// 			} else if isRegular {
// 				logger.Debug("child item is regular file", slog.String("fileName", name))
// 				fileStat, err := e.Info()
// 				if err != nil {
// 					logger.Error("failed to get stat on file", slog.Any("file", entry))
// 					return FSEntity{}, err
// 				}
// 				childEntity := FileInfoToFSEntry(logger, fileStat, parentID, entityPath, shouldCalculateFileHashes)
// 				dirContents = append(dirContents, childEntity)
// 				// wg.Add(1)
// 				// go func() {
// 				// 	defer func() {
// 				// 		if r := recover(); r != nil {
// 				// 			logger.Error("panic for file", slog.Any("reason", r))
// 				// 		}
// 				// 	}()
// 				// 	childEntity := FileInfoToFSEntry(logger, fileStat, parentID, entityPath, shouldCalculateFileHashes)
// 				// 	fsChan <- childEntity
// 				// 	// dirContents = append(dirContents, childEntity)
// 				// 	// wg.Done()
// 				// }()
// 				// fInfo := FileInfoToFSEntry(logger, fileStat, parentID, currentPath, calculateFileHashes)
// 				// fsChan <- fInfo
// 			} else {
// 				logger.Debug("found non regular file / dir... skipping...", slog.Any("file", entry))
// 				continue
// 			}

// 			// logger.Debug("child items is not a dir", slog.String("name", name))
// 		}
// 		// fsChan <- entity
// 		// cfs.WaitGroup.Done()
// 		// wg.Wait()
// 		messagesReceived := 0
// 		for {
// 			if messagesReceived == numItems {
// 				logger.Info("received number of expected messages", slog.Int("messagesReceived", messagesReceived), slog.Int("numItems", numItems))
// 				break
// 			}
// 			select {
// 			case fse, ok := <-fsChan:
// 				logger.Debug("received fs entity message on fsChan", slog.Any("childEntity", fse))
// 				if !ok {
// 					logger.Warn("fsChan closed, done with scan for this item")
// 					break
// 				}
// 				dirContents = append(dirContents, fse)
// 				messagesReceived++
// 			case err := <-errorChan:
// 				return entity, err
// 			}
// 		}
// 		logger.Info("adding children to entity", slog.String("entityPath", entity.FullPath), slog.Int("numChildren", messagesReceived))
// 		entity.Children = dirContents
// 	}
// 	slog.Debug("returning entity", slog.Any("entity", entity))
// 	return entity, nil
// }
