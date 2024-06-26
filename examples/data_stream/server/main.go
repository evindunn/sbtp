package main

import (
	"fmt"
	"github.com/evindunn/sbtp"
	"net"
)

func streamHandler(req *sbtp.SBTPPacket, res *sbtp.SBTPPacket) error {
	payload := req.GetPayload()
	res.SetPayload([]byte("OK"))
	fmt.Print(string(payload))
	return nil
}

func main() {
	const serverAddr = "127.0.0.1:8000"

	server := sbtp.NewSBTPServer()
	server.AddHandler(streamHandler)

	serverListener, err := net.Listen("tcp", serverAddr)
	if err != nil {
		panic(err)
	}
	defer func() {
		if err := serverListener.Close(); err != nil {
			panic(err)
		}
	}()

	fmt.Printf("starting up on %s...\n\n", serverAddr)
	serverListenerTCP := serverListener.(*net.TCPListener)
	server.Start(serverListenerTCP)
}
