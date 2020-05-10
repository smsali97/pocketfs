package services

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net"
	"net/http"
	"strings"
	"time"

	"../models"
	"../repository"
	"../util"
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
			handlePing(commands, broadcast)
			break
		case "HI":
			handleHello(commands)
			break
		}
		fmt.Printf("%s sent this: %s, I'm all good bruh!\n", src, b[:n])
	}
}

func handlePing(commands []string, broadcast string) {
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
		serverRepository := repository.GetServerRepository()
		serverRepository[server.ID] = &server
	}
	repository.ServerMutex.Unlock()

	// send to proxy too
	CLIENT_PORT := "2222"
	addr2, err := net.ResolveUDPAddr("udp4", broadcast+":"+CLIENT_PORT)
	conn2, err := net.DialUDP("udp4", nil, addr2)
	server2, err := json.Marshal(serverRepository[server.ID])
	_, err = conn2.Write([]byte(fmt.Sprintf("PING %s", server2)))

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
		go giveDirectories(msgServer.ID)
		print(serverRepository)
	}

}

func giveDirectories(id string) {
	if id == repository.CurrentServer.ID {
		return // i dont need to give myself directories?
	}

	repository.ServerMutex.RLock()
	serverRepository := repository.GetServerRepository()
	defer repository.ServerMutex.RUnlock()
	url := serverRepository[id].IP + "/directory"

	restClient := http.Client{
		Timeout: time.Second * 20, // Maximum of 2 secs
	}
	var myFiles []models.FileModel
	for _, file := range repository.GetFileRepository() {
		myFiles = append(myFiles,*file)
	}
	byteArray, err := json.Marshal(myFiles)
	if err != nil {
		fmt.Println("Give Directories: Couldnt send files")
	}
	req, err := http.NewRequest(http.MethodPut, url,bytes.NewBuffer(byteArray))
	if err != nil {
		fmt.Println("Give Directories: Couldnt call to server ")
		fmt.Println(err)
		return
	}
	res, getErr := restClient.Do(req)
	if getErr != nil {
		fmt.Println("Couldnt get from server ")
		fmt.Println(err)
		return
	}
	body, readErr := ioutil.ReadAll(res.Body)
	if readErr != nil {
		fmt.Println("Couldnt read from server ")
		fmt.Println(readErr)
		return
	}
	fmt.Println("I got this from the server after sending files " + string(body))
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
