package services

import (
	"../models"
	"../repository"
	"fmt"
	"github.com/google/uuid"
	"time"
)

func AddServer(ip string, port string) {
	serverID, err := uuid.NewUUID()
	if err != nil {
		fmt.Println(err)
		return
	}
	newServer := &models.ServerModel{ID: serverID.String(), IP: ip, Port: port, IsAlive: true, Latency: 0, TimeSinceLastAlive: 0, LastSeen: time.Now()}
	repository.CurrentServer = newServer

	repository.ServerMutex.Lock()
	serverRepository := repository.GetServerRepository()
	serverRepository[newServer.ID] = newServer

	repository.ServerMutex.Unlock()
}
