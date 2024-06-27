package main

import (
	"bufio"
	"errors"
	"fmt"
	"github.com/evindunn/sbtp/pkg"
	"golang.org/x/term"
	"io"
	"net"
	"os"
	"os/signal"
	"time"
)

func main() {
	const serverAddr = "127.0.0.1:8000"

	// Put stdin in unbuffered mode
	oldState, err := term.MakeRaw(int(os.Stdin.Fd()))
	if err != nil {
		panic(err)
	}
	defer func() {
		err = term.Restore(int(os.Stdin.Fd()), oldState)
		if err != nil {
			panic(err)
		}
	}()

	client := sbtp.NewSBTPClient()
	userInput := bufio.NewReader(os.Stdin)

	err = client.Connect("tcp", serverAddr)
	if err != nil {
		panic(err)
	}
	defer func() {
		if err := client.Close(); err != nil && !errors.Is(err, io.EOF) {
			panic(err)
		}
	}()
	tcpConn := client.Conn.(*net.TCPConn)
	err = tcpConn.SetKeepAlive(true)
	if err != nil {
		panic(err)
	}

	fmt.Printf("Connecting to %s...\r\n\r\n", serverAddr)
	interrupt := make(chan os.Signal, 1)
	signal.Notify(interrupt, os.Interrupt)

	shouldStop := false
	for !shouldStop {
		select {
		case <-interrupt:
			shouldStop = true
			break
		default:
			userByte, err := userInput.ReadByte()
			if err != nil {
				fmt.Printf("Error reading user input: %v\r\n", err)
				shouldStop = true
				break
			}
			fmt.Printf("%c", userByte)
			_, err = client.Request([]byte{userByte})
			if err != nil {
				panic(err)
			}
			time.Sleep(1 / 60 * time.Second)
		}
	}
}
