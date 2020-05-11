package filemessage

import (
	"../../models"
	"../../repository"
	"errors"
	"fmt"
	"io/ioutil"
	"os"
	"strings"
	"time"
)

type FileMessage int

type FileMessageType int

const (
	CREATE FileMessageType = 1 + iota
	UPDATE
	DELETE
)

type FileMessageRequest struct {
	RequestId string
	File *models.FileModel
	FileContents []byte
	PreviousID string
	MessageType FileMessageType
	Confirmed chan bool
}

type FileMessageReply struct {
	OriginalRequestId string
	IP string
	IsSuccessful bool
}

type AskFileRequest struct {
	FilePath string
}
type AskFileReply struct {
	File models.FileModel
	FileContents []byte
	IsSuccessful bool
}

type StatusRequest struct {
	Id string
	Status bool
}


func init() {
	RequestStatus = make(map[string]bool)
}

var RequestStatus map[string]bool

func FetchStatus(id string) (bool, error) {
	timeout := time.After(10 * time.Second)
	tick := time.Tick(250 * time.Millisecond)
	// Keep trying until we're timed out or got a result or got an error
	for {
		select {
		// Got a timeout! fail with a timeout error
		case <-timeout:
			return false, errors.New("request timed out")
		// Got a tick, we should check on doSomething()
		case <-tick:
			data, ok :=  RequestStatus[id]
			// Error from doSomething(), we should bail
			if ok {
				return data, nil
			}
		}
	}
}

func (t *FileMessage) UpdateStatus(request StatusRequest, reply *bool) error {
	RequestStatus[request.Id] = request.Status
	*reply = true
	return nil
}

func RemoveStatusEntry(id string) {
	delete(RequestStatus,id) // clean up
}

func (t *FileMessage) AskForFile(request *AskFileRequest, reply *AskFileReply) error {
	reply.IsSuccessful = false

	repository.FileMutex.RLock()
	defer repository.FileMutex.RUnlock()
	if file, ok := repository.FileRepository[request.FilePath]; ok {
		path2, err := os.Getwd()
		if err != nil {
			fmt.Println(err)
		}
		fmt.Println(path2)
		fmt.Println(file.ID)
		//Check if file exists and open
		fileBytes, err := ioutil.ReadFile("file-server/" + file.ID)
		if err != nil {
			return err
		}
		reply.FileContents = fileBytes
		reply.File = *file
		reply.IsSuccessful = true
	} else {
		return errors.New("I dont have the file either!")
	}
	return nil
}


func (t *FileMessage) SendFile(request *FileMessageRequest, reply *FileMessageReply) error {
	// chesck if it should be routed to send directory
	// TODO: sending file commit for one only?
	if request.MessageType == UPDATE {
		err := os.Remove("file-server/"+ request.PreviousID)
		if err != nil {
			fmt.Println("Unable to delete my own previous file " + err.Error())
			//TODO: SHould i return, what if new file is valid
		}
	}

	reply.IsSuccessful = false
	reply.IP = repository.CurrentServer.IP

	fileSent := request.File

	repository.FileMutex.Lock()
	defer repository.FileMutex.Unlock()
	fileRepository := repository.GetFileRepository()
	if fileRepository[fileSent.Path] != nil && fileRepository[fileSent.Path].VersionNumber > fileSent.VersionNumber {
		return  errors.New("I have a later version!")
	}

	tempFile, err := os.Create("file-server/" + fileSent.ID)
	if err != nil {
		return err
	}
	// write this byte array to our temporary file
	var n int
	n, err = tempFile.Write(request.FileContents)
	if err != nil {
		return err
	} else {
		fmt.Println("Wrote " ,n," of bytes")
	}
	//TODO: Disabled file sync for now
	//err = tempFile.Sync()
	//if err != nil {
	//	return err
	//}
	if fileRepository[fileSent.Path] != nil {
		err := os.Remove("file-server/" + fileRepository[fileSent.Path].ID)
		if err != nil {
			fmt.Println("Couldnt remove my previous file")
			return err
		}
		//fileRepository[fileSent.Path].VersionNumber = fileSent.VersionNumber
		//fileRepository[fileSent.Path].LastModified = fileSent.LastModified
		//fileRepository[fileSent.Path].ID =  fileSent.ID
		//fileRepository[fileSent.Path].SizeInBytes = fileSent.SizeInBytes
	}
	err = tempFile.Close()
	if err != nil {
		return err
	}
	reply.IsSuccessful = true
	reply.IP = repository.CurrentServer.IP
	go func() {
		ok, error := FetchStatus(request.RequestId)
		if error != nil {
			fmt.Println(error.Error())
			return
		}
		if !ok {
			fmt.Println("Aborting.. Quorum failed")
			return
		}
		fmt.Println("Got Acknowledgement from server. Performing operation")
		delete(RequestStatus,request.RequestId)
		fileRepository[fileSent.Path] = fileSent
		print(fileRepository)
	}()
	return nil
}

func (t *FileMessage) DeleteFile(request *FileMessageRequest, reply *FileMessageReply) error {
	reply.IsSuccessful = false

	repository.FileMutex.Lock()
	fileRepository := repository.FileRepository
	defer repository.FileMutex.Unlock()

	fileModel := fileRepository[request.File.Path]
	if fileModel == nil {
		return errors.New("Lol. I dont even have the file you wish to delete")
	}
	err := os.Remove("file-server/"+ fileModel.ID)
	if err != nil {
		return errors.New("Unable to physically remove the file from disk")
	}
	go func() {
		ok, error := FetchStatus(request.RequestId)
		if error != nil {
			fmt.Println(error.Error())
			return
		}
		if !ok {
			fmt.Println("Aborting.. Quorum failed")
			return
		}
		fmt.Println("Got Acknowledgement from server. Performing operation")
		delete(RequestStatus,request.RequestId)
		delete(fileRepository,fileModel.Path)
		print(fileRepository)
	}()
	return nil
}

func (t *FileMessage) DeleteDirectory(request *FileMessageRequest, reply *FileMessageReply) error {
	reply.IsSuccessful = false
	repository.FileMutex.Lock()
	//paths := strings.Split(qpath[0], "/")
	fileRepository := repository.GetFileRepository()
	// check all parent directories for correctly formulated path
	if request.File == nil {
		repository.FileMutex.Unlock()
		return errors.New("Directory does not exist")
	}
	path := request.File.Path
	// foo/bar/baz <--- foo/bar
	isDeleted := false
	var pathsToDelete []string
	for key := range fileRepository {
		if len(path) <= len(key) && key[:len(path)] == path {
			pathsToDelete = append(pathsToDelete, path)
			// TODO: What if its a file
			isDeleted = true
		}
	}
	defer repository.FileMutex.Unlock()
	if !isDeleted {
		return errors.New("Couldnt find any directory to delete with " + path)
	}
	reply.IsSuccessful = true
	go func() {
		ok, error := FetchStatus(request.RequestId)
		if error != nil {
			fmt.Println(error.Error())
			return
		}
		if !ok {
			fmt.Println("Aborting.. Quorum failed")
			return
		}
		fmt.Println("Got Acknowledgement from server. Performing operation")
		delete(RequestStatus,request.RequestId)
		for _, path := range pathsToDelete {
			delete(fileRepository, path)
		}
		print(fileRepository)
	}()
	fmt.Println(fileRepository)
	return nil
}

func (t *FileMessage) SendDirectory(request *FileMessageRequest, reply *FileMessageReply) error {
	reply.IsSuccessful = false
	path := request.File.Path
	paths := strings.Split(path, "/")

	repository.FileMutex.Lock()
	defer repository.FileMutex.Unlock()
	fileRepository := repository.GetFileRepository()
	// check all parent directories for correctly formulated path
	if path != "" {
		for i, path := range paths {
			if i != len(paths)-1 && fileRepository[path] == nil {
				return errors.New("Parent directory " + path + " doesnt exist")
			}
		}
	}


	if fileRepository[path] != nil {
		return errors.New(path + " directory already exists in parent")
	}
	reply.IsSuccessful = true
	go func() {
		ok, error := FetchStatus(request.RequestId)
		if error != nil {
			fmt.Println(error.Error())
			return
		}
		if !ok {
			fmt.Println("Aborting.. Quorum failed")
			return
		}
		fmt.Println("Got Acknowledgement from server. Performing operation")
		delete(RequestStatus,request.RequestId)
		fileRepository[path] = request.File
		print(fileRepository)
	}()
	print(fileRepository)
	return nil
}


