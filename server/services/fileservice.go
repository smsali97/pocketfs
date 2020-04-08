package services

import (
	"../models"
	"../repository"
	"fmt"
	"github.com/google/uuid"
	"io"
	"io/ioutil"
	"net/http"
	"os"
	"strconv"
	"strings"
)


func DownloadFile(w http.ResponseWriter, r *http.Request) {
	fmt.Println("File Download Endpoint Hit")


	names, ok := r.URL.Query()["name"]
	if !ok || len(names[0]) < 1 {
		http.Error(w,"Url Param 'name' is missing",400)
		return
	}

	name := names[0]
	fileRepository := repository.GetRepository()
	filePath := strings.Split(name, "/")
	var file* models.FileModel

	if len(filePath) > 1 {

		for _, path := range filePath {
			// first time
			if file == nil {
				file = fileRepository[path]
			}
			if file != nil {

			}

		}

	} else {
		file = fileRepository[filePath[0]]
	}

	fmt.Println(fileRepository)
	if file == nil {
		//File not found, send 404
		http.Error(w, "File not found in repository .", 404)
		return
	}


	//Check if file exists and open
	Openfile, err := os.Open("file-server/" + file.ID)
	defer Openfile.Close() //Close after function return

	if err != nil {
		//File not found, send 404
		http.Error(w, "File not found on server.", 404)
		return
	}

	fmt.Println("Sending file to client " + file.Name)
	//File is found, create and send the correct headers

	//Get the Content-Type of the file
	//Create a buffer to store the header of the file in
	FileHeader := make([]byte, 512)
	//Copy the headers into the FileHeader buffer
	Openfile.Read(FileHeader)
	//Get content type of file
	FileContentType := http.DetectContentType(FileHeader)

	//Get the file size
	FileStat, _ := Openfile.Stat()                     //Get info from file
	FileSize := strconv.FormatInt(FileStat.Size(), 10) //Get file size as a string

	//Send the headers
	w.Header().Set("Content-Disposition", "attachment; filename="+file.Name)
	w.Header().Set("Content-Type", FileContentType)
	w.Header().Set("Content-Length", FileSize)

	// Send the file
	// We read 512 bytes from the file already, so we reset the offset back to 0
	Openfile.Seek(0, 0)
	io.Copy(w, Openfile) //'Copy' the file to the client
	return

}


func UploadFile(w http.ResponseWriter, r *http.Request) {
	fmt.Println("File Upload Endpoint Hit")

	// Retrieve path of where to upload the file
	qpath, ok := r.URL.Query()["path"]
	fileRepository := repository.GetRepository()
	if !ok || len(qpath[0]) < 1 {
		http.Error(w,"Url Param 'path' is missing",400)
		return
	}

	path := strings.Split(qpath[0],"/")
	var directory *models.FileModel
	var tempDirectory *models.FileModel
	if len(path) == 0 {
		directory = nil
	} else {

		for _, parentDir := range path {
			if tempDirectory == nil {
				tempDirectory = fileRepository[parentDir]
			} else {
				var childDirectory *models.FileModel
				for _, currDir := range tempDirectory.Children {
					if currDir.IsDirectory && currDir.Name == parentDir {
						childDirectory = currDir
					}
				}
				tempDirectory = childDirectory
			}
			if tempDirectory == nil {
				http.Error(w, "Directory does not exist. Cannot upload file there",400)
				return
			}
		}
		directory = tempDirectory
	}


	// upload of 50 MB files.
	// TODO: Add File Size Check
	r.ParseMultipartForm(50 << 20)

	file, handler, err := r.FormFile("myFile")
	if err != nil {
		fmt.Println("Error Retrieving the File")
		fmt.Println(err)
		return
	}
	defer file.Close()
	fmt.Printf("Uploaded File: %+v\n", handler.Filename)
	fmt.Printf("File Size: %+v\n", handler.Size)
	fmt.Printf("MIME Header: %+v\n", handler.Header)

	id, err := uuid.NewUUID()
	if err != nil {
		fmt.Println(err)
		return
	}

	tempFile, err := os.Create("file-server/" + id.String())
	if err != nil {
		fmt.Println(err)
		return
	}
	defer tempFile.Close()

	// read all of the contents of our uploaded file into a
	// byte array
	fileBytes, err := ioutil.ReadAll(file)
	if err != nil {
		fmt.Println(err)
		return
	}
	// write this byte array to our temporary file
	tempFile.Write(fileBytes)
	// return that we have successfully uploaded our file!
	fmt.Fprintf(w, "Successfully Uploaded File\n")

	newFile := models.FileModel{ID: id.String(), IsDirectory: false, Name: handler.Filename}
	if directory == nil {
		fileRepository[handler.Filename] = &newFile
	} else {
		directory.Children = append(directory.Children,&newFile)
	}

	fmt.Println(len(fileRepository))

}

