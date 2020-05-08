package filemessage

import (
	"../../models"
	"../../repository"
	"errors"
	"fmt"
	"github.com/google/uuid"
	"io/ioutil"
	"os"
	"strings"
)

type FileMessage int

type FileMessageType int

const (
	CREATE FileMessageType = 1 + iota
	UPDATE
	DELETE
)

type FileMessageRequest struct {
	File *models.FileModel
	FileContents []byte
	PreviousID string
	MessageType FileMessageType
}

type FileMessageReply struct {
	IP string
	IsSuccessful bool
}

type DirectoryMessageRequest struct {
	File *models.FileModel
	FileContents []byte
}

type DirectoryMessageReply struct {
	IP string
	IsSuccessful bool
}
// TODO: Send Directory updates too

type AskFileRequest struct {
	FilePath string
}
type AskFileReply struct {
	File models.FileModel
	FileContents []byte
	IsSuccessful bool
}


func (t *FileMessage) AskForFile(request *AskFileRequest, reply *AskFileReply) error {
	reply.IsSuccessful = false

	repository.FileMutex.RLock()
	defer repository.FileMutex.RUnlock()
	if file, ok := repository.FileRepository[request.FilePath]; ok {

		//Check if file exists and open
		Openfile, err := os.Open("file-server/" + file.ID)
		if err != nil {
			return err
		}
		defer Openfile.Close() //Close after function return
		fileBytes, err := ioutil.ReadAll(Openfile)
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
	// check if it should be routed to send directory

	if request.File.IsDirectory {
		if request.MessageType == DELETE {
			return DeleteDirectory(request,reply)
		}
		return SendDirectory(request,reply)
	}
	if request.MessageType == DELETE {
		return DeleteFile(request, reply)
 	} else if request.MessageType == UPDATE {
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

	id, err := uuid.NewUUID()
	tempFile, err := os.Create("file-server/" + id.String())
	if err != nil {
		return err
	}
	defer tempFile.Close()
	// write this byte array to our temporary file
	var n int
	n, err = tempFile.Write(request.FileContents)
	if err != nil {
		return err
	} else {
		fmt.Println("Wrote " ,n," of bytes")
	}
	err = tempFile.Sync()
	if err != nil {
		return err
	}
	if fileRepository[fileSent.Path] != nil {
		err := os.Remove("file-server/" + fileRepository[fileSent.Path].ID)
		if err != nil {
			return err
		}
		// TODO: Check Path formulation
		fileRepository[fileSent.Path].VersionNumber = fileSent.VersionNumber
		fileRepository[fileSent.Path].LastModified = fileSent.LastModified
		fileRepository[fileSent.Path].ID = id.String()
		fileRepository[fileSent.Path].SizeInBytes = fileSent.SizeInBytes
	} else {
		fileRepository[fileSent.Path] = fileSent
	}
	reply.IsSuccessful = true
	reply.IP = repository.CurrentServer.IP
	print(fileRepository)
	return nil
}

func DeleteFile(request *FileMessageRequest, reply *FileMessageReply) error {
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
	delete(fileRepository,fileModel.Path)
	return nil
}

func DeleteDirectory(request *FileMessageRequest, reply *FileMessageReply) error {
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
	for key := range fileRepository {
		if len(path) <= len(key) && key[:len(path)] == path {
			delete(fileRepository, path)
			// TODO: What if its a file
			isDeleted = true
		}
	}
	repository.FileMutex.Unlock()
	if !isDeleted {
		return errors.New("Couldnt find any directory to delete with " + path)
	}
	reply.IsSuccessful = true
	fmt.Println(fileRepository)
	return nil
}

func SendDirectory(request *FileMessageRequest, reply *FileMessageReply) error {
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
	fileRepository[path] = request.File
	reply.IsSuccessful = true
	print(fileRepository)
	return nil
}


