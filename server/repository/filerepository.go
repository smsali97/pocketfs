package repository

import (
	"../models"
	"github.com/sasha-s/go-deadlock"
	"sync"
)

var FileRepository map[string]*models.FileModel
var fileOnce sync.Once
var FileMutex deadlock.RWMutex

func GetFileRepository() map[string]*models.FileModel {

	fileOnce.Do(func() {
		FileRepository = make(map[string]*models.FileModel)
	})

	return FileRepository
}
