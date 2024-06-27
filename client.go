package sbtp

import (
	"errors"
	"fmt"
	"github.com/evindunn/sbtp/internal"
	"net"
	"sync"
	"time"
)

// SBTPClient is a wrapper around a [net.Conn] for managing connections to SBTP servers
type SBTPClient struct {
	Conn    net.Conn
	timeout time.Duration
	lock    sync.Mutex
}

// NewSBTPClient creates a ready-to-use [SBTPClient]
func NewSBTPClient() *SBTPClient {
	return &SBTPClient{
		Conn:    nil,
		timeout: time.Second * 5,
		lock:    sync.Mutex{},
	}
}

// SetTimeout sets the timeout for read and write operations to the underlying [net.Conn]
func (c *SBTPClient) SetTimeout(timeout time.Duration) {
	c.lock.Lock()
	defer c.lock.Unlock()
	c.timeout = timeout
}

func (c *SBTPClient) getTimeout() time.Duration {
	c.lock.Lock()
	defer c.lock.Unlock()
	return c.timeout
}

// Connect disconnects from any existing SBTP server and begins a new connection over the given protocol and serverAddr
func (c *SBTPClient) Connect(protocol string, serverAddr string) error {
	err := c.Close()
	if err != nil {
		return err
	}

	conn, err := net.DialTimeout(protocol, serverAddr, c.getTimeout())
	if err != nil {
		return err
	}

	c.Conn = conn
	return nil
}

// Close closes the current SBTP server connection, if any
func (c *SBTPClient) Close() error {
	if c.Conn != nil {
		return c.Conn.Close()
	}
	return nil
}

// Request sends an [SBTPPacket] with the given payload to the currently connected server and returns an
// [SBTPPacket] response
func (c *SBTPClient) Request(requestPayload []byte) (*SBTPPacket, error) {
	if c.Conn == nil {
		return nil, errors.New("not connected")
	}

	request := NewSBTPPacket(c.Conn.LocalAddr())
	response := NewSBTPPacket(c.Conn.RemoteAddr())

	request.SetPayload(requestPayload)

	err := internal.UpdateDeadline(c.Conn, c.getTimeout())
	if err != nil {
		return nil, fmt.Errorf("error setting request deadline: %s", err)
	}

	_, err = request.WriteTo(c.Conn)
	if err != nil {
		return nil, err
	}

	err = internal.UpdateDeadline(c.Conn, c.getTimeout())
	if err != nil {
		return nil, fmt.Errorf("error setting response deadline: %s", err)
	}

	_, err = response.ReadFrom(c.Conn)
	if err != nil {
		return nil, err
	}

	return response, nil
}
