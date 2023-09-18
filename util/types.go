package util

import "time"

type File struct {
	Name      string `json:"name"`
	Extension string `json:"extension"`
	Path      string `json:"path"`
	Size      int64  `json:"size"`
}

type EntityType string

const (
	FileType      EntityType = "file"
	DirectoryType EntityType = "directory"
	OtherType     EntityType = "other"
)

type FSEntity struct {
	ID           string     `json:"id,omitempty"`
	ParentID     string     `json:"parentID,omitempty"`
	Name         string     `json:"name,omitempty"`
	Extension    string     `json:"extension,omitempty"`
	FullPath     string     `json:"fullPath,omitempty"`
	Size         int64      `json:"size"`
	PrettySize   string     `json:"prettySize"`
	FileHash     string     `json:"fileHash,omitempty"`
	IsDir        bool       `json:"isDir"`
	EntityType   EntityType `json:"entityType"`
	Mode         uint32     `json:"mode,omitempty"`
	Type         uint32     `json:"type"`
	Permissions  uint32     `json:"permissions"`
	LastModified time.Time  `json:"lastModified"`
	Children     []FSEntity `json:"children,omitempty"`
}
