package spaceanalyzer

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"sync"

	"log/slog"

	"github.com/google/uuid"
)

const (
	jobsChannelBufferSize = 32
	rootParentID          = ""
	defaultRootID         = "root"
)

var (
	ErrRecursionLimit = errors.New("recursion limit")
	emptyFSJob        = FSJob{}
)

type ConcurrentFSScanner interface {
	Scan(logger *slog.Logger, entityPath, rootID string, calculateFileHashes bool, maxRecursion int) (FSEntity, error)
}

type concurrentFSScanner struct {
	concurrencyLimit int
}

func NewConcurrentFSScanner(concurrencyLimit int) ConcurrentFSScanner {
	if concurrencyLimit <= 0 {
		concurrencyLimit = runtime.NumCPU()
	}
	return &concurrentFSScanner{
		concurrencyLimit: concurrencyLimit,
	}
}

func (cfs *concurrentFSScanner) Scan(logger *slog.Logger, entityPath, rootID string, shouldCalculateFileHashes bool, maxRecursion int) (FSEntity, error) {
	logger.Info("starting concurrent scan",
		slog.Int("concurrencyLimit", cfs.concurrencyLimit),
		slog.String("entityPath", entityPath),
		slog.String("rootID", rootID),
		slog.Bool("shouldCalculateFileHashes", shouldCalculateFileHashes),
		slog.Int("maxRecursion", maxRecursion),
	)
	files := make(map[string][]FSEntity)
	dirs := make(map[string][]FSEntity)
	limiter := make(chan bool, cfs.concurrencyLimit)
	wg := sync.WaitGroup{}
	if len(rootID) == 0 {
		logger.Warn("creating new id because one provided was blank")
		rootID = defaultRootID
	}
	jobsChan := enumerateScanTargets(logger, entityPath, rootParentID, rootID, maxRecursion)
	mutex := sync.Mutex{}
	wg.Add(1)
	go func() {
		for {
			j, ok := <-jobsChan
			limiter <- true
			if !ok {
				logger.Warn("jobs channel closed")
				if j == emptyFSJob {
					logger.Debug("last item from channel was empty")
					break
				}
			}
			wg.Add(1)
			go func() {
				defer func() {
					<-limiter
					wg.Done()
				}()
				if j.FailedScan {
					logger.Warn("failed scan job",
						slog.String("errorMessage", j.Error.Error()),
						slog.String("id", j.ID),
						slog.String("parentID", j.ParentID),
						slog.String("fullPath", j.FullPath),
					)
				}
				entity := FileInfoToFSEntry(logger, j.Info, j.ParentID, j.ID, j.FullPath, shouldCalculateFileHashes, j.Depth)
				entity.Depth = j.Depth
				if j.Error != nil {
					if len(entity.ErrorMessage) > 0 {
						entity.ErrorMessage = fmt.Sprintf("%s|||||%s", j.Error.Error(), entity.ErrorMessage)
					} else {
						entity.ErrorMessage = j.Error.Error()
					}
				}
				if j.IsDir {
					logger.Debug("processed dir job", slog.Any("dirEntity", entity))
					mutex.Lock()
					dirs[entity.ParentID] = append(dirs[entity.ParentID], entity)
					mutex.Unlock()
				} else {
					mutex.Lock()
					files[entity.ParentID] = append(files[j.ParentID], entity)
					mutex.Unlock()
				}
				// <-limiter
				// wg.Done()
			}()
		}
		wg.Done()
	}()
	wg.Wait()
	close(limiter)
	rootEntity := dirs[rootParentID][0]
	delete(dirs, rootParentID)
	logger.Info("collating entities")
	entity := collateEntities(logger, rootEntity, files, dirs)
	logger.Info("finished collating entities")
	return entity, nil
}

func collateEntities(logger *slog.Logger, entity FSEntity, files map[string][]FSEntity, dirs map[string][]FSEntity) FSEntity {
	// add files
	entityFiles := files[entity.ID]
	numFiles := 0
	numDirs := 0
	if len(entityFiles) > 0 {
		delete(files, entity.ID)
		numFiles += len(entityFiles)
		entity.Children = append(entity.Children, entityFiles...)
	}
	// add directories
	entityChildDirs := dirs[entity.ID]
	delete(dirs, entity.ID)
	for _, v := range entityChildDirs {
		childDir := collateEntities(logger, v, files, dirs)
		entity.Children = append(entity.Children, childDir)
		numDirs++
	}
	logger.Debug("entity collated", slog.String("entityID", entity.ID), slog.Int("numFiles", numFiles), slog.Int("numDirs", numDirs))
	return entity
}

func enumerateScanTargets(logger *slog.Logger, entityPath, parentID, id string, maxRecursion int) <-chan FSJob {
	logger.Info("enumerating targets")
	jobsChan := make(chan FSJob, jobsChannelBufferSize)
	go func() {
		recursiveEnumerateScanTargets(logger, jobsChan, entityPath, parentID, id, maxRecursion, 0)
		logger.Info("finished enumerating targets")
		close(jobsChan)
	}()
	return jobsChan
}

func recursiveEnumerateScanTargets(logger *slog.Logger, jobsChan chan<- FSJob, entityPath, parentID, id string, maxRecursion, recursionCount int) {
	logger = logger.With(slog.String("parentID", parentID), slog.String("entityPath", entityPath))
	currentPath := entityPath
	var job FSJob
	defer func() {
		jobsChan <- job
	}()
	if !filepath.IsAbs(currentPath) {
		var err error
		currentPath, err = filepath.Abs(currentPath)
		if err != nil {
			job.FailedScan = true
			job.Error = fmt.Errorf("failed to get absolute path for non absolute input: %w", err)
			return
		}
	}
	if maxRecursion > -1 && maxRecursion >= recursionCount {
		logger.Warn("skipping call to enumerate entity due to max recursion setting", slog.Int("maxRecursion", maxRecursion), slog.Int("recursionCount", recursionCount), slog.String("currentPath", currentPath))
		job.FailedScan = true
		job.Error = ErrRecursionLimit
		return
	}
	currentStat, err := os.Stat(currentPath)
	if err != nil {
		job.FailedScan = true
		job.Error = fmt.Errorf("failed to get stat on current path: %w", err)
		return
	}
	job.ID = id
	if len(job.ID) == 0 {
		job.ID = uuid.New().String()
	}
	job.ParentID = parentID
	job.Info = currentStat
	job.FullPath = currentPath
	job.Depth = recursionCount
	logger = logger.With("id", id)
	if currentStat.IsDir() {
		job.IsDir = true
		dirContents, err := os.ReadDir(currentPath)
		if err != nil {
			job.FailedScan = true
			job.Error = fmt.Errorf("failed to get contents of directory: %w", err)
			return
		}
		numChildren := len(dirContents)
		logger.Debug("dir contents retrieved", slog.Int("numChildren", numChildren))
		for _, d := range dirContents {
			childName := d.Name()
			childPath := filepath.Join(currentPath, childName)
			childID := uuid.New().String()
			recursiveEnumerateScanTargets(logger, jobsChan, childPath, job.ID, childID, maxRecursion, recursionCount+1)
		}
	}
}
