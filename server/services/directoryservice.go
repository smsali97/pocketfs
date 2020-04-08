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
		AddDirectory(w,r)
	case "DELETE":
		RemoveDirectory(w,r)
	default:
		http.Error(w,  r.Method + " Method Not supported In Directory Service", 405)
	}
}

func AddDirectory (w http.ResponseWriter, r *http.Request) {
	qpath, ok := r.URL.Query()["path"]
	if !ok || len(qpath[0]) < 1 {
		http.Error(w,"Url Param 'path' is missing",400)
		return
	}
	fileRepository := repository.GetRepository()


	path := strings.Split(qpath[0],"/")
	directoryName := path[len(path)-1]
	id, err := uuid.NewUUID()
	if err != nil {
		fmt.Println(err)
		return
	}
	if len(path) > 1 {
		parent := fileRepository[path[len(path)-2]]
		// parent directory doesnt exist
		if parent == nil {
			http.Error(w,"Parent directory doesnt exist",404)
			return
		}
		for _, dir := range  parent.Children {
			if dir.Name == directoryName {
				http.Error(w, "Name directory already exists in parent",400)
				return
			}
		}
		newDirectory := models.FileModel{Name: directoryName,
										Path: qpath[0],
										ID: id.String(),
										Children: []*models.FileModel{},
										IsDirectory:true}
		parent.Children = append(parent.Children, &newDirectory)
	} else {
		// insert at root
		if fileRepository[path[0]]	!= nil {
			http.Error(w, "Name directory already exists in parent",400)
		} else {
			newDirectory := models.FileModel{Name: path[0],
				Path: qpath[0],
				ID: id.String(),
				Children: []*models.FileModel{},
				IsDirectory:true}
			fileRepository[path[0]]	= &newDirectory
		}
	}

}


func RemoveDirectory (w http.ResponseWriter, r *http.Request) {

}