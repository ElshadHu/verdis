package server

import (
	"context"
	"errors"
	"fmt"
	"net"
	"sync"
	"syscall"
	"time"

	"github.com/ElshadHu/verdis/internal/command"
	"github.com/ElshadHu/verdis/internal/command/standard"
	"github.com/ElshadHu/verdis/internal/command/version"
	"github.com/ElshadHu/verdis/internal/mvcc"
)

type ErrAddressInUse struct {
	Address string
	Err     error
}

func (e *ErrAddressInUse) Error() string {
	return fmt.Sprintf(
		"cannot start server: address %s is already in use\n"+
			"  Possible solutions:\n"+
			"  - Stop the existing process using this port (lsof -i :%s | grep LISTEN)\n"+
			"  - Use a different port with WithPort() configuration\n"+
			"  - Wait for the port to be released (TIME_WAIT state)",
		e.Address, extractPort(e.Address),
	)
}

func (e *ErrAddressInUse) Unwrap() error {
	return e.Err
}

func extractPort(addr string) string {
	_, port, err := net.SplitHostPort(addr)
	if err != nil {
		return addr
	}
	return port
}

type Server struct {
	cfg      *Config
	listener net.Listener
	router   *command.Router
	engine   *mvcc.Engine

	// conns is a hashset of connections protected by sync.Mutex
	conns map[*Connection]struct{}

	// mu protects conns
	mu sync.Mutex

	done bool
	wg   sync.WaitGroup

	// connLimit is a semaphore for limiting the number of connections
	connLimit chan struct{}
}

func NewServer(cfg *Config) (*Server, error) {
	if err := cfg.Validate(); err != nil {
		return nil, err
	}

	engine := mvcc.NewEngine()

	router := command.NewRouter()
	ctx := &command.Context{Engine: engine}
	router.SetContext(ctx)
	standard.RegisterAll(router)
	version.RegisterAll(router)

	return &Server{
		cfg:       cfg,
		router:    router,
		engine:    engine,
		conns:     make(map[*Connection]struct{}),
		connLimit: make(chan struct{}, cfg.MaxConnections),
	}, nil
}

func (s *Server) Start(ctx context.Context) error {
	var err error
	s.listener, err = s.listenWithRetry(ctx)
	if err != nil {
		return err
	}

	// Close listener when context is cancelled
	go func() {
		<-ctx.Done()
		s.Shutdown()
	}()

	for {
		if s.done {
			break
		}

		conn, err := s.listener.Accept()
		if err != nil {
			if s.done || errors.Is(err, net.ErrClosed) {
				break
			}

			time.Sleep(50 * time.Millisecond)
			continue
		}

		// Enforce connection limit
		select {
		case s.connLimit <- struct{}{}:
			// Allowed
		default:
			conn.Close()
			continue
		}

		// Wrap connection in RESPConnection
		respConn := newConnection(s, conn)

		s.mu.Lock()
		s.conns[respConn] = struct{}{}
		s.mu.Unlock()

		s.wg.Add(1)
		go func() {
			defer func() {
				s.mu.Lock()
				delete(s.conns, respConn)
				s.mu.Unlock()
				<-s.connLimit
				respConn.Close()
				s.wg.Done()
			}()
			respConn.Serve(s.router)
		}()
	}

	return nil
}

func (s *Server) listenWithRetry(ctx context.Context) (net.Listener, error) {
	addr := s.cfg.Address()

	const (
		maxRetries     = 3
		initialBackoff = 100 * time.Millisecond
		maxBackoff     = 2 * time.Second
	)

	var lastErr error
	backoff := initialBackoff

	for attempt := 0; attempt <= maxRetries; attempt++ {
		listener, err := net.Listen("tcp", addr)
		if err == nil {
			return listener, nil
		}

		lastErr = err

		if !isAddressInUseError(err) {
			return nil, fmt.Errorf("failed to listen on %s: %w", addr, err)
		}

		if attempt == maxRetries {
			break
		}

		select {
		case <-ctx.Done():
			return nil, ctx.Err()
		case <-time.After(backoff):
			backoff = min(backoff*2, maxBackoff)
		}
	}

	return nil, &ErrAddressInUse{Address: addr, Err: lastErr}
}

func isAddressInUseError(err error) bool {
	var opErr *net.OpError
	if errors.As(err, &opErr) {
		var syscallErr *syscall.Errno
		if errors.As(opErr.Err, &syscallErr) {
			return *syscallErr == syscall.EADDRINUSE
		}
		var sysErr syscall.Errno
		if errors.As(opErr.Err, &sysErr) {
			return sysErr == syscall.EADDRINUSE
		}
	}
	return false
}

// Shutdown gracefully stops the server
func (s *Server) Shutdown() {
	s.mu.Lock()
	if s.done {
		s.mu.Unlock()
		return
	}
	s.done = true

	if s.listener != nil {
		s.listener.Close()
	}
	// Close all active connections
	for c := range s.conns {
		c.conn.Close()
	}
	s.mu.Unlock()
	// wait for all connection handlers to finish
	s.wg.Wait()
}

// Address returns the current listening address
func (s *Server) Address() net.Addr {
	if s.listener != nil {
		return s.listener.Addr()
	}
	return nil
}
