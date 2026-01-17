package server

import (
	"bufio"
	"io"
	"log/slog"
	"net"

	"github.com/ElshadHu/verdis/internal/command"
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

// Serve handles the command loop for the connection
func (c *Connection) Serve(router *command.Router) {
	for {
		cmd, err := c.respConn.ReadCommand()
		if err != nil {
			if err == io.EOF {
				return
			}
			c.respConn.WriteResponse(protocol.NewError("ERR " + err.Error()))
			continue
		}
		result := router.Execute(cmd)
		if err := c.respConn.WriteResponse(result); err != nil {
			slog.Error("Unexpected result occured while attempting to write a response")
			return
		}
	}
}

func (c *Connection) Close() {
	c.respConn.Close()
}
