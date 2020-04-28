package services

import (
	"../models"
	"../repository"
	"../util"
	"encoding/json"
	"fmt"
	"net"
	"strings"
	"time"
)

var MAX_DATAGRAM_SIZE = 8192

func ListenForBroadcast(ip string, port string, broadcast string) {
	addr, err := net.ResolveUDPAddr("udp4", ip+":"+port)
	util.CheckError(err)
	l, err := net.ListenUDP("udp4", addr)
	util.CheckError(err)
	l.SetReadBuffer(MAX_DATAGRAM_SIZE)
	for {
		b := make([]byte, MAX_DATAGRAM_SIZE)
		n, src, err := l.ReadFromUDP(b)
		util.CheckError(err)

		commands := strings.Split(string(b[:n]), " ")
		switch commands[0] {
		case "PING":
			handlePing(commands)
			break
		case "HI":
			handleHello(commands)
			break
		}
		fmt.Printf("%s sent this: %s, I'm all good bruh!\n", src, b[:n])
	}
}

func handlePing(commands []string) {
	// TODO: handle case where you do not know about the server?
	if len(commands) < 2 {
		util.RaiseCustomError("Invalid Command Given For Ping")
	}
	repository.ServerMutex.Lock()
	serverRepository := repository.GetServerRepository()

	var server models.ServerModel
	err := json.Unmarshal([]byte(commands[1]), &server)
	util.CheckError(err)
	serverID := server.ID

	// youve seen the server before
	if serverRepository[serverID] != nil {
		serverRepository[serverID].Latency = time.Since(serverRepository[serverID].LastSeen).Seconds()
		serverRepository[serverID].LastSeen = time.Now()
		serverRepository[serverID].NoOfPings = serverRepository[serverID].NoOfPings + 1
	} else {

		//serverRepository[serverID].LastSeen = time.Now()
		//serverRepository[serverID].NoOfPings = 0
	}
	repository.ServerMutex.Unlock()
}

func handleHello(commands []string) {
	if len(commands) < 2 {
		util.RaiseCustomError("Invalid Command Given For HELLO")
	}
	var msgServer models.ServerModel
	err := json.Unmarshal([]byte(commands[1]), &msgServer)
	util.CheckError(err)

	serverRepository := repository.GetServerRepository()
	if serverRepository[msgServer.ID] == nil {
		repository.ServerMutex.Lock()

		serverRepository[msgServer.ID] = &msgServer

		repository.ServerMutex.Unlock()

		print(serverRepository)
	}

}

func SendHello(broadcast string, port string) {
	// I do not exist yet, how can I ping?!
	if repository.CurrentServer == nil {
		return
	}

	addr, err := net.ResolveUDPAddr("udp4", broadcast+":"+port)
	util.CheckError(err)

	conn, err := net.DialUDP("udp4", nil, addr)
	util.CheckError(err)
	msg, err := json.Marshal(repository.CurrentServer)
	util.CheckError(err)
	_, err = conn.Write([]byte(fmt.Sprintf("HI %s", string(msg))))
	util.CheckError(err)
}
