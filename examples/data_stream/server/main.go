package main

import (
	"fmt"
	"github.com/evindunn/sbtp/pkg"
	"net"
	"strings"
	"unicode"
)

func streamHandler(req *sbtp.SBTPPacket, res *sbtp.SBTPPacket) error {
	payload := req.GetPayload()
	printableOnly := strings.Map(func(r rune) rune {
		if unicode.IsPrint(r) {
			return r
		}
		return -1
	}, string(payload))
	res.SetPayload([]byte("OK"))
	fmt.Print(printableOnly)
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
