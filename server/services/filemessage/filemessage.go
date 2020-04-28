package filemessage

import (
	"../../models"
	"../../repository"
	"errors"
	"github.com/google/uuid"
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

func (t *FileMessage) SendFile(request *FileMessageRequest, reply *FileMessageReply) error {
	reply.IsSuccessful = false
	reply.IP = repository.CurrentServer.IP

	fileSent := request.File

	repository.FileMutex.Lock()
	fileRepository := repository.GetFileRepository()
	if fileRepository[fileSent.Path] != nil && fileRepository[fileSent.Path].VersionNumber > fileSent.VersionNumber {
		repository.FileMutex.Unlock()
		return  errors.New("I have a later version!")
	}

	id, err := uuid.NewUUID()
	tempFile, err := os.Create("file-server/" + id.String())
	if err != nil {
		repository.FileMutex.Unlock()
		return err
	}
	defer tempFile.Close()


	// write this byte array to our temporary file
	tempFile.Write(request.FileContents)
	if fileRepository[fileSent.Path] != nil {
		err := os.Remove("file-server/" + fileRepository[fileSent.Path].ID)
		if err != nil {
			repository.FileMutex.Unlock()
			return err
		}
		// TODO: Check Path formulation
		fileRepository[fileSent.Path].VersionNumber = fileSent.VersionNumber
		fileRepository[fileSent.Path].LastModified = fileSent.LastModified
		fileRepository[fileSent.Path].ID = id.String()
	} else {
		fileRepository[fileSent.Path] = fileSent
	}
	reply.IsSuccessful = true
	reply.IP = repository.CurrentServer.IP

	return nil
}


