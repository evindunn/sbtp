package sbtp

import (
	"errors"
	"fmt"
	"github.com/evindunn/sbtp/internal"
	"net"
	"time"
)

// SBTPClient is a wrapper around a [net.Conn] for managing connections to SBTP servers
type SBTPClient struct {
	conn    net.Conn
	timeout time.Duration
}

// NewSBTPClient creates a ready-to-use [SBTPClient]
func NewSBTPClient() *SBTPClient {
	return &SBTPClient{
		conn:    nil,
		timeout: time.Second * 5,
	}
}

// SetTimeout sets the timeout for read and write operations to the underlying [net.Conn]
func (c *SBTPClient) SetTimeout(timeout time.Duration) {
	c.timeout = timeout
}

// Connect disconnects from any existing SBTP server and begins a new connection over the given protocol and serverAddr
func (c *SBTPClient) Connect(protocol string, serverAddr string) error {
	err := c.Close()
	if err != nil {
		return err
	}

	conn, err := net.DialTimeout(protocol, serverAddr, c.timeout)
	if err != nil {
		return err
	}

	c.conn = conn
	return nil
}

// Close closes the current SBTP server connection, if any
func (c *SBTPClient) Close() error {
	if c.conn != nil {
		return c.conn.Close()
	}
	return nil
}

// Request sends an [SBTPPacket] with the given payload to the currently connected server and returns an
// [SBTPPacket] response
func (c *SBTPClient) Request(requestPayload []byte) (*SBTPPacket, error) {
	if c.conn == nil {
		return nil, errors.New("not connected")
	}

	request := NewSBTPPacket(c.conn.LocalAddr())
	response := NewSBTPPacket(c.conn.RemoteAddr())

	request.SetPayload(requestPayload)

	err := internal.UpdateDeadline(c.conn, c.timeout)
	if err != nil {
		return nil, fmt.Errorf("error setting request deadline: %s", err)
	}

	_, err = request.WriteTo(c.conn)
	if err != nil {
		return nil, err
	}

	err = internal.UpdateDeadline(c.conn, c.timeout)
	if err != nil {
		return nil, fmt.Errorf("error setting response deadline: %s", err)
	}

	_, err = response.ReadFrom(c.conn)
	if err != nil {
		return nil, err
	}

	return response, nil
}
