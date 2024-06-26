package internal

import (
	"net"
	"time"
)

// UpdateDeadline sets the deadline on the given [net.Conn] to [time.Now] plus timeout
func UpdateDeadline(conn net.Conn, timeout time.Duration) error {
	err := conn.SetDeadline(time.Now().Add(timeout))
	if err != nil {
		return err
	}
	return nil
}
