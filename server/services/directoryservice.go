package services

import (
	"../models"
	"../repository"
	"../services/filemessage"
	"encoding/json"
	"fmt"
	"github.com/google/uuid"
	"net/http"
	"strings"
	"time"
)

func MainDirectoryService(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Access-Control-Allow-Methods", "POST, GET, OPTIONS, PUT, DELETE")
	w.Header().Set("Access-Control-Allow-Headers", "Accept, Content-Type, Content-Length, Accept-Encoding, X-CSRF-Token, Authorization")

	fmt.Println(r.Method)
	switch r.Method {
	case "POST":
		AddDirectory(w, r)
		return
	case "DELETE":
		RemoveDirectory(w, r)
		return
	case "GET":
		GetDirectories(w,r)
		return
	case "PUT":
		PutDirectories(w,r)
		return
	case "OPTIONS":
		w.WriteHeader(http.StatusOK)
		return
	default:
		http.Error(w, r.Method+" Method Not supported In Directory Service", 405)
	}
}

func PutDirectories(w http.ResponseWriter, r *http.Request) {
	var serverFiles []models.ClientFileModel
	decoder := json.NewDecoder(r.Body)
	jsonErr := decoder.Decode(&serverFiles)
	if jsonErr != nil {
		fmt.Println("Couldnt read from json")
		return
	}
	const layout = "Mon Jan 02 15:04:05 -0700 2006"
	for _, serverFile := range serverFiles {
		repository.FileMutex.Lock()
		repo := repository.GetFileRepository()
		if repo[serverFile.Path] == nil || repo[serverFile.Path].VersionNumber < serverFile.VersionNumber {
			timeParse, _ := time.Parse(layout, serverFile.Modified)
			newFile := &models.FileModel{
				ID:            serverFile.ID,
				IsDirectory:   serverFile.IsDirectory,
				Path:          serverFile.Path,
				Name:          serverFile.Name,
				VersionNumber: serverFile.VersionNumber,
				LastModified:  timeParse,
				SizeInBytes:   serverFile.SizeInBytes,
			}
			repo[serverFile.Path] = newFile
		}
	}
	repository.FileMutex.Unlock()
	w.Write([]byte("success"))
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
		repository.FileMutex.Unlock()
		m := &models.FileModel{Name: paths[len(paths)-1],
			Path:        qpath[0],
			ID:          id.String(),
			//Children:    []*models.FileModel{},
			IsDirectory: true}
		requestId, err := uuid.NewUUID()
		if err != nil {
			fmt.Println("Failure in generating new id")
		}
		FileChannel <- &filemessage.FileMessageRequest{RequestId: requestId.String(), File:m , FileContents: nil, MessageType: filemessage.CREATE}
	}

}

func RemoveDirectory(w http.ResponseWriter, r *http.Request) {
	qpath, ok := r.URL.Query()["path"]
	if !ok || len(qpath[0]) < 1 {
		http.Error(w, "Url Param 'path' is missing", 400)
		return
	}

	fmt.Println("Got a request to remove " + qpath[0])

	repository.FileMutex.RLock()
	file := repository.GetFileRepository()[qpath[0]]
	repository.FileMutex.RUnlock()
	requestId, err := uuid.NewUUID()
	if err != nil {
		fmt.Println("Failure in generating new id")
	}
	FileChannel <- &filemessage.FileMessageRequest{
		RequestId: requestId.String(),
		File:        file,
		FileContents: nil,
		MessageType:  filemessage.DELETE,
	}
}
