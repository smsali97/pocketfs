package repository

import (
	"../models"
	"sync"
)

var FileRepository map[string]*models.FileModel
var fileOnce sync.Once

func GetRepository() map[string]*models.FileModel {

	fileOnce.Do(func()  {
		FileRepository = make(map[string]*models.FileModel)
	})

	return FileRepository
}


