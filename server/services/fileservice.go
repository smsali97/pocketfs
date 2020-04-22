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


	qpath, ok := r.URL.Query()["path"]
	if !ok || len(qpath[0]) < 1 {
		http.Error(w,"Url Param 'path' is missing",400)
		return
	}

	path := qpath[0]


	repository.FileMutex.RLock() // START READING FROM FILE REPOSITORY

	fileRepository := repository.GetFileRepository()
	filePath := strings.Split(path, "/")
	var file* models.FileModel

	tempPath := ""
	for i, path := range filePath {
		tempPath += path
		if i != len(filePath) - 1 && fileRepository[tempPath] == nil  {
			http.Error(w, "Directory " + path + " does not exist", 404)
			repository.FileMutex.RUnlock() // FILE REPOSITORY
			return
		}
		tempPath += "/"
	}

	file = fileRepository[path]
	repository.FileMutex.RUnlock() // END READING FROM FILE REPOSITORY
	if file == nil {
		//File not found, send 404
		http.Error(w, "File not found in given path", 404)
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
	qPath, ok := r.URL.Query()["path"]
	if !ok || len(qPath[0]) < 1 {
		http.Error(w,"Url Param 'path' is missing",400)
		return
	}

	paths := strings.Split(qPath[0],"/")

	repository.FileMutex.Lock() // START FROM FILE REPOSITORY

	fileRepository := repository.GetFileRepository()
	tempPath := ""
	if qPath[0] != "/" {
		for _, path := range paths {
			tempPath += path
			//  TODO: Extension check , isDirectory?
			if fileRepository[tempPath] == nil {
				//  check if directory exists or not
				http.Error(w, "Directory does not exist. Cannot upload file there", 400)
				repository.FileMutex.Unlock()
				return
			}
			tempPath += "/"
		}
	}
	// upload of 50 MB files.
	// TODO: Add File Size Check
	r.ParseMultipartForm(50 << 20)

	file, handler, err := r.FormFile("myFile")
	if err != nil {
		fmt.Println("Error Retrieving the File")
		fmt.Println(err)
		repository.FileMutex.Unlock()
		return
	}
	defer file.Close()
	fmt.Printf("Uploaded File: %+v\n", handler.Filename)
	fmt.Printf("File Size: %+v\n", handler.Size)
	fmt.Printf("MIME Header: %+v\n", handler.Header)

	id, err := uuid.NewUUID()
	if err != nil {
		fmt.Println(err)
		repository.FileMutex.Unlock()
		return
	}

	tempFile, err := os.Create("file-server/" + id.String())
	if err != nil {
		fmt.Println(err)
		repository.FileMutex.Unlock()
		return
	}
	defer tempFile.Close()

	// read all of the contents of our uploaded file into a
	// byte array
	fileBytes, err := ioutil.ReadAll(file)
	if err != nil {
		fmt.Println(err)
		repository.FileMutex.Unlock()
		return
	}
	// write this byte array to our temporary file
	tempFile.Write(fileBytes)
	// return that we have successfully uploaded our file!
	fmt.Fprintf(w, "Successfully Uploaded File\n")

	newFile := models.FileModel{ID: id.String(), IsDirectory: false, Name: handler.Filename}
	if qPath[0] == "/" {
		qPath[0] = ""
	} else {
		qPath[0] += "/"
	}

	fileRepository[qPath[0] + handler.Filename] = &newFile
	repository.FileMutex.Unlock() // END FROM FILE REPOSITORY

	fmt.Println(fileRepository)

}

