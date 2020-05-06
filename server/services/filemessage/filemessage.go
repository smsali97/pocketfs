package filemessage

import (
	"../../models"
	"../../repository"
	"errors"
	"github.com/google/uuid"
	"io/ioutil"
	"os"
)

type FileMessage int

type FileMessageRequest struct {
	File *models.FileModel
	FileContents []byte
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
	tempFile.Write(request.FileContents)
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

	return nil
}


