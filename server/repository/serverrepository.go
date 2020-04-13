package repository

import (
	"../models"
	"sync"
)

var ServerRepository map[string]*models.ServerModel
var once sync.Once

func GetRepository() map[string]*models.ServerModel {

	once.Do(func()  {
		ServerRepository = make(map[string]*models.ServerModel)
	})

	return ServerRepository
}


