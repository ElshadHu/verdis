package server

import (
	"bufio"
	"net"

	"github.com/ElshadHu/verdis/internal/protocol"
)

// Connection wraps a client connection with RESP protocol handling
type Connection struct {
	conn     net.Conn
	respConn *protocol.RESPConnection
}

func newConnection(s *Server, conn net.Conn) *Connection {
	// Set TCP buffer sizes
	if tcpConn, ok := conn.(*net.TCPConn); ok {
		tcpConn.SetReadBuffer(s.cfg.ReadBufferSize)
		tcpConn.SetWriteBuffer(s.cfg.WriteBufferSize)
	}

	reader := bufio.NewReader(conn)
	writer := bufio.NewWriter(conn)
	return &Connection{
		conn:     conn,
		respConn: protocol.NewRESPConnection(reader, writer),
	}
}

func (c *Connection) Close() {
	c.respConn.Close()
}
