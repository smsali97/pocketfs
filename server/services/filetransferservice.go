package services

import (
	"../repository"
	"../services/filemessage"
	"../util"
	"fmt"
	"net"
	"net/rpc"
	"sync"
)


var FileChannel chan *filemessage.FileMessageRequest
var RPC_PORT string

func init() {
	FileChannel = make(chan *filemessage.FileMessageRequest)
	print("Instantiating channels...")
	RPC_PORT = "1234"
}

func HandleFileTransfers() {
	go handleIncomingFiles()
	handleOutgoingFiles()
}

func HandleFileRequests(filePathRequest string) []*filemessage.AskFileReply {
	repository.ServerMutex.RLock()
	defer repository.ServerMutex.RUnlock()
	serverRepository := repository.GetServerRepository()
	var waitGroup sync.WaitGroup
	fileRequest := filemessage.AskFileRequest{}
	fileRequest.FilePath = filePathRequest
	var replies []*filemessage.AskFileReply

	for _, server := range serverRepository {
		if !server.IsAlive {
			continue
		}
		client, err := rpc.Dial("tcp", server.IP + ":" + RPC_PORT)
		if err != nil {
			fmt.Println("%v Server is down",server.IP)
		}
		reply := &filemessage.AskFileReply{}

		waitGroup.Add(1)
		go func() {
			defer waitGroup.Done()
			err := client.Call("FileMessage.AskForFile", fileRequest, reply)
			if err != nil {
				fmt.Println(" error when sending message to server ",err,server.IP)
			} else {
				if reply.IsSuccessful {
					replies = append(replies,reply)
				}
			}
		}()
	}
	waitGroup.Wait()
	return replies
}

func handleOutgoingFiles() {

	for {
		fileRequest := <-FileChannel
		fmt.Println("Received a file request")

		repository.ServerMutex.RLock()
		//SENDING_THRESHOLD := len(repository.GetServerRepository()) - 1
		SENT_CTR := 0
		QUORUM_CTR := 0
		var waitGroup sync.WaitGroup
		serverRepo := repository.GetServerRepository()
		for _, server := range serverRepo {
			if !server.IsAlive {
				continue
			}

			client, err := rpc.Dial("tcp", server.IP+":"+RPC_PORT)
			if err != nil {
				fmt.Println(server.IP + " Server is down")
				delete(serverRepo,server.IP)
				continue
			}
			reply := &filemessage.FileMessageReply{}

			waitGroup.Add(1)
			go func() {
				defer waitGroup.Done()
				err := client.Call("FileMessage.SendFile", fileRequest, reply)
				if err != nil {
					fmt.Println(" error when sending message to server ", err, server.IP)
				} else {
					if reply.IsSuccessful {
						QUORUM_CTR += 1
					}
				}
			}()
			SENT_CTR += 1
		}
		waitGroup.Wait()
		fmt.Printf("Got a quorum of %d for %d\n", QUORUM_CTR, SENT_CTR)
		repository.ServerMutex.RUnlock()

	}
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