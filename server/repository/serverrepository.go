package repository

import (
	"../models"
	"sync"
)

var serverRepository map[string]*models.ServerModel
var CurrentServer *models.ServerModel
var once sync.Once
var ServerMutex sync.RWMutex

func GetServerRepository() map[string]*models.ServerModel {

	once.Do(func()  {
		serverRepository = make(map[string]*models.ServerModel)
	})

	return serverRepository
}


