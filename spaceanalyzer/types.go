package spaceanalyzer

import (
	"io/fs"
	"time"
)

type EntityType string

const (
	FileType      EntityType = "file"
	DirectoryType EntityType = "directory"
	OtherType     EntityType = "other"
)

// TODO: add errors hash and scan to struct...
type FSEntity struct {
	ID           string     `json:"id,omitempty"`
	ParentID     string     `json:"parentID,omitempty"`
	Name         string     `json:"name,omitempty"`
	Extension    string     `json:"extension,omitempty"`
	FullPath     string     `json:"fullPath,omitempty"`
	Size         int64      `json:"size"`
	PrettySize   string     `json:"prettySize"`
	FileHash     string     `json:"fileHash,omitempty"`
	IsDir        bool       `json:"isDir"` // TODO: remove and have client calculate based on entityType?
	Depth        int        `json:"depth"`
	EntityType   EntityType `json:"entityType"`
	Mode         uint32     `json:"mode,omitempty"`
	Type         uint32     `json:"type"`
	Permissions  uint32     `json:"permissions"`
	LastModified time.Time  `json:"lastModified"`
	Children     []FSEntity `json:"children,omitempty"`
	ErrorMessage string     `json:"errorMessage,omitempty"`
}

type FSJob struct {
	ID         string
	ParentID   string
	FullPath   string
	Info       fs.FileInfo
	IsDir      bool
	Depth      int
	FailedScan bool
	Error      error
}
