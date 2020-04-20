package services

import (
	"../models"
	"../repository"
	"fmt"
	"github.com/google/uuid"
	"net/http"
	"strings"
)

func MainDirectoryService(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case "POST":
		AddDirectory(w, r)
	case "DELETE":
		RemoveDirectory(w, r)
	default:
		http.Error(w, r.Method+" Method Not supported In Directory Service", 405)
	}
}

func AddDirectory(w http.ResponseWriter, r *http.Request) {
	qpath, ok := r.URL.Query()["path"]
	if !ok || len(qpath[0]) < 1 {
		http.Error(w, "Url Param 'path' is missing", 400)
		return
	}
	fileRepository := repository.GetFileRepository()

	paths := strings.Split(qpath[0], "/")

	// check all parent directories for correctly formulated path
	for i, path := range paths {
		if i != len(path)-1 && fileRepository[path] == nil {
			http.Error(w, "Parent directory "+path+" doesnt exist", 404)
			return
		}
	}

	if fileRepository[qpath[0]] == nil {
		http.Error(w, "Name directory already exists in parent", 400)
		return
	} else {
		id, err := uuid.NewUUID()
		if err != nil {
			fmt.Println(err)
			return
		}
		fileRepository[qpath[0]] = &models.FileModel{Name: paths[len(paths)-1],
			Path:        qpath[0],
			ID:          id.String(),
			Children:    []*models.FileModel{},
			IsDirectory: true}
	}

}

func RemoveDirectory(w http.ResponseWriter, r *http.Request) {
	qpath, ok := r.URL.Query()["path"]
	if !ok || len(qpath[0]) < 1 {
		http.Error(w, "Url Param 'path' is missing", 400)
		return
	}
	fileRepository := repository.GetFileRepository()
	paths := strings.Split(qpath[0], "/")

	// check all parent directories for correctly formulated path
	for _, path := range paths {
		if fileRepository[path] == nil {
			http.Error(w, "Directory "+path+" doesnt exist", 404)
			return
		}
	}

	// foo/bar/baz <--- foo/bar

	for key := range fileRepository {
		if len(qpath[0]) <= len(key) && key[:len(qpath[0])] == qpath[0] {
			delete(fileRepository,qpath[0])
		}
	}

	fmt.Println(fileRepository)
}
