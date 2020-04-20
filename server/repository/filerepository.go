package repository

import (
	"../models"
	"sync"
)

var FileRepository map[string]*models.FileModel
var fileOnce sync.Once
var FileMutex sync.RWMutex


func GetFileRepository() map[string]*models.FileModel {

	fileOnce.Do(func()  {
		FileRepository = make(map[string]*models.FileModel)
	})

	return FileRepository
}


