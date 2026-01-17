package server

import (
	"context"
	"errors"
	"net"
	"sync"
	"time"

	"github.com/ElshadHu/verdis/internal/command"
	"github.com/ElshadHu/verdis/internal/command/standard"
	"github.com/ElshadHu/verdis/internal/mvcc"
)

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
	// TODO: Handle error when another server is running on same port
	s.listener, err = net.Listen("tcp", s.cfg.Address())
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
