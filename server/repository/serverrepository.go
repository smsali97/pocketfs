package repository

import (
	"../models"
	"github.com/sasha-s/go-deadlock"
	"sync"
)

var serverRepository map[string]*models.ServerModel
var CurrentServer *models.ServerModel
var once sync.Once
var ServerMutex deadlock.RWMutex

func GetServerRepository() map[string]*models.ServerModel {

	once.Do(func() {
		serverRepository = make(map[string]*models.ServerModel)
	})

	return serverRepository
}
