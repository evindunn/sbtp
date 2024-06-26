package sbtp

import (
	"errors"
	"fmt"
	"github.com/evindunn/sbtp/internal"
	"io"
	"net"
	"time"
)

const ServerDefaultTimeout = 30 * time.Second

// SBTPRequestHandler is the type for all SBTPServer requestHandlers
type SBTPRequestHandler func(request *SBTPPacket, response *SBTPPacket) error

// NetListenerWithDeadline represents a [net.Listener] with the SetDeadline method
type NetListenerWithDeadline interface {
	net.Listener
	SetDeadline(t time.Time) error
}

// SBTPServer is a convenience wrapper around a [net.Listener] that implements the SeadDeadline method
type SBTPServer struct {
	shouldStop      chan bool
	isStopped       chan bool
	timeout         time.Duration
	requestHandlers []SBTPRequestHandler
}

func NewSBTPServer() *SBTPServer {
	return &SBTPServer{
		shouldStop:      make(chan bool, 1),
		isStopped:       make(chan bool, 1),
		timeout:         ServerDefaultTimeout,
		requestHandlers: make([]SBTPRequestHandler, 0, 5),
	}
}

// SetTimeout sets the timeout for read and write operations given a [net.Conn] to a client
func (s *SBTPServer) SetTimeout(timeout time.Duration) {
	s.timeout = timeout
}

// AddHandler adds a new request/response handler to the SBTPServer. Handlers are run in the order in which they were
// passed to AddHandler.
func (s *SBTPServer) AddHandler(handler SBTPRequestHandler) {
	s.requestHandlers = append(s.requestHandlers, handler)
}

// handleConnection runs for each client connected to the SBTPServer, serving requests until the client disconnects
// or SBTPServer.timeout is hit on a read or write operation.
func (s *SBTPServer) handleConnection(conn net.Conn) {
	fmt.Printf("Accepted connection from %s\n", conn.RemoteAddr())

	for {
		request := NewSBTPPacket(conn.RemoteAddr())
		response := NewSBTPPacket(conn.LocalAddr())

		err := internal.UpdateDeadline(conn, s.timeout)
		if err != nil {
			fmt.Printf("Error setting request deadline: %s\n", err)
			continue
		}

		_, err = request.ReadFrom(conn)
		if errors.Is(err, io.EOF) {
			break
		}
		if err != nil {
			fmt.Printf("Error reading request: %s\n", err)
			break
		}

		for _, h := range s.requestHandlers {
			err = h(request, response)
			if err != nil {
				fmt.Printf("Error handling request: %s\n", err)
				continue
			}
		}

		err = internal.UpdateDeadline(conn, s.timeout)
		if err != nil {
			fmt.Printf("Error setting response deadline: %s\n", err)
			continue
		}

		_, err = response.WriteTo(conn)
		if errors.Is(err, io.EOF) {
			break
		}
		if err != nil {
			fmt.Printf("Error sending response: %s\n", err)
			break
		}
	}

	fmt.Printf("Connection to %s closed\n", conn.RemoteAddr())
}

// Start starts an SBTPServer on the given listener
func (s *SBTPServer) Start(listener NetListenerWithDeadline) {
	defer func() {
		s.isStopped <- true
	}()

	shouldStop := false
	for !shouldStop {
		err := listener.SetDeadline(time.Now().Add(s.timeout))
		if err != nil {
			fmt.Printf("Error setting listener deadline: %s\n", err)
		}

		select {
		case <-s.shouldStop:
			shouldStop = true
			break
		default:
			conn, err := listener.Accept()
			if err != nil {
				continue
			}
			go s.handleConnection(conn)
		}
	}
}

// Stop stops the currently running SBTPServer
func (s *SBTPServer) Stop() {
	s.shouldStop <- true
	waitFor := true
	for waitFor {
		select {
		case <-s.isStopped:
			waitFor = false
			break
		default:
			time.Sleep(1 * time.Second)
		}
	}
}
