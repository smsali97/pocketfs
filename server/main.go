package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net"
	"net/http"
	"time"

	"../server/repository"
	"../server/services"
	"../server/util"
)

var PORT = "49401"

const SUBNET_MASK = "255.255.255.0"

func setupRoutes() {
	http.HandleFunc("/directory", services.MainDirectoryService)
	http.HandleFunc("/download", services.DownloadFile)
	http.HandleFunc("/upload", services.UploadFile)
	http.HandleFunc("/remove", services.RemoveFile)
	http.ListenAndServe(":8080", nil)
}

func GetOutboundIP() net.IP {
	conn, err := net.Dial("udp", "8.8.8.8:80")
	if err != nil {
		log.Fatal(err)
	}
	defer conn.Close()

	localAddr := conn.LocalAddr().(*net.UDPAddr)

	return localAddr.IP
}

func main() {
	IP_ADDRESS := GetOutboundIP()
	fmt.Println("My IP: " + IP_ADDRESS.String())

	mask := net.CIDRMask(24, 32)
	ip := net.IP(IP_ADDRESS)
	broadcast := makeBroadcast(ip, mask)

	services.AddServer(IP_ADDRESS.String(), PORT) // add yourself to the repository

	services.CleanFiles()
	go setupRoutes()
	go services.ListenForBroadcast(IP_ADDRESS.String(), PORT, broadcast.String())

	services.SendHello(broadcast.String(), PORT) // send hello to others so they know you exist and can contact you

	go services.HandleFileTransfers() // handle incoming and outgoing file transfers

	pingServers(broadcast) // periodically ping servers
}

func pingServers(broadcast net.IP) {
	CLIENT_PORT := "2222"
	// I do not exist yet, how can I ping?!
	if repository.CurrentServer == nil {
		return
	}

	addr, err := net.ResolveUDPAddr("udp4", broadcast.String()+":"+PORT)
	addr2, err := net.ResolveUDPAddr("udp4", broadcast.String()+":"+CLIENT_PORT)
	util.CheckError(err)

	conn, err := net.DialUDP("udp4", nil, addr)
	conn2, err := net.DialUDP("udp4", nil, addr2)
	util.CheckError(err)

	ctr := 0
	for {
		repository.ServerMutex.RLock()
		server, err := json.Marshal(repository.CurrentServer)
		repository.ServerMutex.RUnlock()
		util.CheckError(err)
		_, err = conn.Write([]byte(fmt.Sprintf("PING %s", server)))
		_, err = conn2.Write([]byte(fmt.Sprintf("PING %s", server)))
		util.CheckError(err)
		ctr++
		time.Sleep(5 * time.Second)
	}
}

func makeBroadcast(ip net.IP, mask net.IPMask) net.IP {
	broadcast := net.IP(make([]byte, 4))
	for i := range ip {
		broadcast[i] = ip[i] | ^mask[i]
	}
	return broadcast
}
