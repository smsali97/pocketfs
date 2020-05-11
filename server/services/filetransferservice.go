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
		var requestIps []string
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
			requestIps = append(requestIps,server.IP)
			waitGroup.Add(1)
			go func() {
				defer waitGroup.Done()
				var err error
				if fileRequest.MessageType == filemessage.CREATE && fileRequest.File.IsDirectory {
					err = client.Call("FileMessage.SendDirectory", fileRequest, reply)
				} else if fileRequest.MessageType == filemessage.DELETE && fileRequest.File.IsDirectory {
					err = client.Call("FileMessage.DeleteDirectory", fileRequest, reply)
				} else if fileRequest.MessageType == filemessage.DELETE {
					err = client.Call("FileMessage.DeleteFile", fileRequest, reply)
				}  else {
					err = client.Call("FileMessage.SendFile", fileRequest, reply)
				}
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
		if SENT_CTR - QUORUM_CTR <= 1 && QUORUM_CTR > 0 {
			fmt.Println("Committing transaction..")
			statusRequest := filemessage.StatusRequest{
				Id:    fileRequest.RequestId,
				Status: true,
			}
			for _, requestIp := range requestIps {
				client, _ := rpc.Dial("tcp", requestIp+":"+RPC_PORT)
				_ = client.Call("FileMessage.UpdateStatus", statusRequest, nil)
			}
		} else {
			fmt.Println("Quorum failed. Rollback...")
			statusRequest := filemessage.StatusRequest{
				Id:    fileRequest.RequestId,
				Status: false,
			}
			for _, requestIp := range requestIps {
				client, _ := rpc.Dial("tcp", requestIp+":"+RPC_PORT)
				_ = client.Call("FileMessage.UpdateStatus", statusRequest, nil)
			}
		}
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