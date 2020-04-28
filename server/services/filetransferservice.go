package services

import (
	"../models"
	"../repository"
	"../services/filemessage"
	"../util"
	"fmt"
	"io/ioutil"
	"net"
	"net/rpc"
	"os"
	"sync"
)


var FileChannel chan *models.FileModel


func HandleFileTransfers(fileChannel chan *models.FileModel) {
	FileChannel = fileChannel

	go handleIncomingFiles()
	handleOutgoingFiles()
}

func handleOutgoingFiles() {
	RPC_PORT := "1234"

	fileData := <- FileChannel
	tempFile, err := os.Open("file-server/" + fileData.ID)
	defer tempFile.Close()
	util.CheckError(err)
	data, err := ioutil.ReadAll(tempFile)
	util.CheckError(err)
	fileRequest := &filemessage.FileMessageRequest{
		File:         fileData,
		FileContents: data,
	}

	repository.ServerMutex.RLock()
	//SENDING_THRESHOLD := len(repository.GetServerRepository()) - 1
	SENT_CTR := 0
	QUORUM_CTR := 0
	var waitGroup sync.WaitGroup
	for _, server := range repository.GetServerRepository() {
		if !server.IsAlive {
			continue
		}

		client, err := rpc.Dial("tcp", server.IP + ":" + RPC_PORT)
		if err != nil {
			fmt.Println("%v Server is down",server.IP)
		}
		reply := &filemessage.FileMessageReply{}

		waitGroup.Add(1)
		go func() {
			defer waitGroup.Done()
			err := client.Call("FileMessage.SendFile", fileRequest, reply)
			if err != nil {
				fmt.Println(" error when sending message to server ",err,server.IP)
			} else {
				if reply.IsSuccessful {
					QUORUM_CTR += 1
				}
			}
		}()
		SENT_CTR += 1
	}
	repository.ServerMutex.RUnlock()
	waitGroup.Wait()
	fmt.Printf("Got a quorum of %d for %d\n",QUORUM_CTR,SENT_CTR)

}

func handleIncomingFiles() {
	RPC_PORT := "1234"

	fileMessage := new(filemessage.FileMessage)

	rpc.Register(fileMessage)

	tcpAddr, err := net.ResolveTCPAddr("tcp", ":" + RPC_PORT)
	util.CheckError(err)

	listener, err := net.ListenTCP("tcp", tcpAddr)
	util.CheckError(err)

	go rpc.Accept(listener)
}