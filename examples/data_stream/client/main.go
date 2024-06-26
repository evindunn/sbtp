package main

import (
	"errors"
	"fmt"
	"github.com/evindunn/sbtp"
	"golang.org/x/term"
	"io"
	"os"
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
	userInput := make([]byte, 1)

	err = client.Connect("tcp", serverAddr)
	if err != nil {
		panic(err)
	}
	defer func() {
		if err := client.Close(); err != nil {
			panic(err)
		}
	}()

	fmt.Printf("Connecting to %s...\n\n", serverAddr)

	for {
		bytesRead, err := os.Stdin.Read(userInput)
		if errors.Is(err, io.EOF) {
			break
		}
		if err == nil && bytesRead == 1 {
			fmt.Print(string(userInput))
			_, err = client.Request(userInput)
			if err != nil {
				panic(err)
			}
		}
		time.Sleep(1 / 60 * time.Second)
	}
}
