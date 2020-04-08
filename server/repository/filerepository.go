package repository

import (
	"../models"
	"sync"
)

var FileRepository map[string]*models.FileModel
var once sync.Once

func GetRepository() map[string]*models.FileModel {

	once.Do(func()  {
		FileRepository = make(map[string]*models.FileModel)
	})

	return FileRepository
}


