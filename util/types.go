package util

type File struct {
	Name      string `json:"name"`
	Extension string `json:"extension"`
	Path      string `json:"path"`
	Size      int64  `json:"size"`
}
