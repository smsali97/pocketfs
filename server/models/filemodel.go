package models

import "time"

// FileModel is a model of a file
type FileModel struct {
	ID            string       `json:"id"`
	IsDirectory   bool         `json:"isDirectory"`
	Path          string       `json:"path"`
	Name          string       `json:"name"`
	VersionNumber int          `json:"versionNumber"`
	//Children      []*FileModel `json:"children"`
	LastModified  time.Time    `json:"lastModified"`
	SizeInBytes  int64    `json:"sizeInBytes"`
}

type ClientFileModel struct {
	Key string `json:"key"`
	Size int64 `json:"size"`
	Modified string  `json:"modified"`
}
