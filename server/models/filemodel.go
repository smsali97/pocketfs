package models

// FileModel is a model of a file
type FileModel struct {
	ID          string         `json:"id"`
	IsDirectory bool        `json:"isDirectory"`
	Path        string      `json:"path"`
	Name        string      `json:"name"`
	VersionNumber int		`json:"versionNumber"`
	Children    []*FileModel `json:"children"`
}
