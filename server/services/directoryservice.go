package services

import (
	"../models"
	"../repository"
	"encoding/json"
	"fmt"
	"github.com/google/uuid"
	"net/http"
	"strings"
)

func MainDirectoryService(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Access-Control-Allow-Methods", "POST, GET, OPTIONS, PUT, DELETE")
	w.Header().Set("Access-Control-Allow-Headers", "Accept, Content-Type, Content-Length, Accept-Encoding, X-CSRF-Token, Authorization")

	fmt.Println(r.Method)
	switch r.Method {
	case "POST":
		AddDirectory(w, r)
	case "DELETE":
		RemoveDirectory(w, r)
	case "GET":
		GetDirectories(w,r)
	case "OPTIONS":
		w.WriteHeader(http.StatusOK)
		return
	default:
		http.Error(w, r.Method+" Method Not supported In Directory Service", 405)
	}
}

func GetDirectories(w http.ResponseWriter, r *http.Request) {
	const layout = "Mon Jan 02 15:04:05 -0700 2006"


	repository.FileMutex.RLock()
	defer repository.FileMutex.RUnlock()
	fileRepository := repository.GetFileRepository()
	keys := make([]models.ClientFileModel, 0, len(fileRepository))
	for k, v := range fileRepository {
		var s string
		if v.IsDirectory {
			s = "/"
		} else {
			s = ""
		}

		newModel := models.ClientFileModel{
			Key:      k + s,
			Size:     v.SizeInBytes,
			Modified: v.LastModified.Format(layout),
		}
		keys = append(keys, newModel)
	}
	json.NewEncoder(w).Encode(keys)
	fmt.Println(keys)

	return
}

func AddDirectory(w http.ResponseWriter, r *http.Request) {
	qpath, ok := r.URL.Query()["path"]
	if !ok || len(qpath[0]) < 1 {
		http.Error(w, "Url Param 'path' is missing", 400)
		return
	}

	fmt.Println("Got a request to add " + qpath[0])

	repository.FileMutex.Lock()
	fileRepository := repository.GetFileRepository()

	paths := strings.Split(qpath[0], "/")

	// check all parent directories for correctly formulated path
	for i, path := range paths {
		if i != len(paths)-1 && fileRepository[path] == nil {
			fmt.Println(i)
			fmt.Println(len(path) - 1)
			repository.FileMutex.Unlock()
			http.Error(w, "Parent directory "+path+" doesnt exist", 404)
			return
		}
	}

	if fileRepository[qpath[0]] != nil {
		repository.FileMutex.Unlock()
		http.Error(w, qpath[0] + "directory already exists in parent", 400)
		return
	} else {
		id, err := uuid.NewUUID()
		if err != nil {
			repository.FileMutex.Unlock()
			fmt.Println(err)
			return
		}
		fileRepository[qpath[0]] = &models.FileModel{Name: paths[len(paths)-1],
			Path:        qpath[0],
			ID:          id.String(),
			//Children:    []*models.FileModel{},
			IsDirectory: true}
		repository.FileMutex.Unlock()
		fmt.Println(fileRepository)
	}

}

func RemoveDirectory(w http.ResponseWriter, r *http.Request) {
	qpath, ok := r.URL.Query()["path"]
	if !ok || len(qpath[0]) < 1 {
		http.Error(w, "Url Param 'path' is missing", 400)
		return
	}

	fmt.Println("Got a request to remove " + qpath[0])

	repository.FileMutex.Lock()
	//paths := strings.Split(qpath[0], "/")

	fileRepository := repository.GetFileRepository()
	// check all parent directories for correctly formulated path
	if fileRepository[qpath[0]] == nil {
		repository.FileMutex.Unlock()
		http.Error(w, "Directory "+qpath[0]+" doesnt exist", 404)
		return
	}

	// foo/bar/baz <--- foo/bar
	isDeleted := false
	for key := range fileRepository {
		if len(qpath[0]) <= len(key) && key[:len(qpath[0])] == qpath[0] {
			delete(fileRepository, qpath[0])
			isDeleted = true
		}
	}
	repository.FileMutex.Unlock()
	if !isDeleted {
		http.Error(w, "Couldn't find " + qpath[0] + " to delete", 400)
	}
	fmt.Println(fileRepository)
}
