package main

import (
	"fmt"
	"github.com/evindunn/sbtp/v1"
	"net"
	"slices"
	"time"
)

func echoHandler(request *v1.SBTPPacket, response *v1.SBTPPacket) error {
	requestPayload := request.GetPayload()
	fmt.Printf("Echoing %d bytes back to %s\n", len(requestPayload), request.SourceAddr())
	response.SetPayload(requestPayload)
	return nil
}

func repeatHandler(_ *v1.SBTPPacket, response *v1.SBTPPacket) error {
	fmt.Println("Repeating response...")
	responsePayload := response.GetPayload()
	response.SetPayload(slices.Concat(responsePayload, responsePayload))
	return nil
}

func main() {
	const serverAddr = "127.0.0.1:8000"

	request := []byte("Hello!")

	server := v1.NewSBTPServer()
	server.AddHandler(echoHandler)
	server.AddHandler(repeatHandler)

	serverListener, err := net.Listen("tcp", serverAddr)
	if err != nil {
		panic(err)
	}
	serverListenerTCP := serverListener.(*net.TCPListener)

	go server.Start(serverListenerTCP)
	time.Sleep(2 * time.Second)

	client := v1.NewSBTPClient()
	err = client.Connect("tcp", serverAddr)
	if err != nil {
		err = fmt.Errorf("Error connecting to server: %s\n", err)
		panic(err)
	}

	for i := 0; i < 10; i++ {
		response, err := client.Request(request)
		if err != nil {
			err = fmt.Errorf("Request error: %s\n", err)
			panic(err)
		}
		fmt.Printf("Got %d bytes from server\n", len(response.GetPayload()))
	}

	time.Sleep(2 * time.Second)

	err = client.Close()
	if err != nil {
		panic(err)
	}

	fmt.Println("Stopping server...")
	server.Stop()
}
