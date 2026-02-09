package server

import (
	"errors"
	"fmt"
	"net"
	"strconv"
	"time"
)

var (
	ErrInvalidPort             = errors.New("port must be between 0 and 65535")
	ErrInvalidAddress          = errors.New("invalid address")
	ErrNegativeReadTimeout     = errors.New("read timeout must be non-negative")
	ErrNegativeWriteTimeout    = errors.New("write timeout must be non-negative")
	ErrNegativeIdleTimeout     = errors.New("idle timeout must be non-negative")
	ErrNonPositiveMaxConns     = errors.New("max connections must be positive")
	ErrNonPositiveReadBufSize  = errors.New("read buffer size must be positive")
	ErrNonPositiveWriteBufSize = errors.New("write buffer size must be positive")
)

// ConfigOption applies a configuration setting to a Config.
type ConfigOption func(*Config) error

// Config holds the server network and resource settings.
type Config struct {
	Host string
	Port int

	// ReadTimeout is the max duration for reading a full request (where 0 = no timeout).
	ReadTimeout time.Duration

	// WriteTimeout is the max duration for writing a full response (where 0 = no timeout).
	WriteTimeout time.Duration

	// IdleTimeout is how long an idle connection is kept alive (where 0 = no timeout).
	IdleTimeout time.Duration

	// MaxConnections is the max number of concurrent connections (where 0 = unlimited).
	MaxConnections int

	// ReadBufferSize is the per-connection read buffer in bytes.
	ReadBufferSize int

	// WriteBufferSize is the per-connection write buffer in bytes.
	WriteBufferSize int
}

// NewDefaultConfig creates a Config with sensible defaults with variadic options.
func NewDefaultConfig(opts ...ConfigOption) (*Config, error) {
	conf := &Config{
		Host:            "0.0.0.0",
		Port:            6379,
		ReadBufferSize:  4096, // 4 KB
		WriteBufferSize: 4096, // 4 KB
	}

	for _, opt := range opts {
		if err := opt(conf); err != nil {
			return nil, fmt.Errorf("applying config option: %w", err)
		}
	}

	if err := conf.Validate(); err != nil {
		return nil, fmt.Errorf("invalid config: %w", err)
	}

	return conf, nil
}

// Address returns the host:port string for net.Listen.
func (c *Config) Address() string {
	return net.JoinHostPort(c.Host, strconv.Itoa(c.Port))
}

// Validate checks all config fields for invalid values.
func (c *Config) Validate() error {
	if c.Port < 0 || c.Port > 65535 {
		return ErrInvalidPort
	}
	if c.ReadTimeout < 0 {
		return ErrNegativeReadTimeout
	}
	if c.WriteTimeout < 0 {
		return ErrNegativeWriteTimeout
	}
	if c.IdleTimeout < 0 {
		return ErrNegativeIdleTimeout
	}
	if c.MaxConnections <= 0 {
		return ErrNonPositiveMaxConns
	}
	if c.ReadBufferSize <= 0 {
		return ErrNonPositiveReadBufSize
	}
	if c.WriteBufferSize <= 0 {
		return ErrNonPositiveWriteBufSize
	}
	return nil
}

// WithHost sets the listen host address.
func WithHost(host string) ConfigOption {
	return func(c *Config) error {
		c.Host = host
		return nil
	}
}

// WithPort sets the listen port.
func WithPort(port int) ConfigOption {
	return func(c *Config) error {
		c.Port = port
		return nil
	}
}

// WithAddress sets host and port from a "host:port" string.
func WithAddress(address string) ConfigOption {
	return func(c *Config) error {
		host, portStr, err := net.SplitHostPort(address)
		if err != nil {
			return fmt.Errorf("%w %q: %w", ErrInvalidAddress, address, err)
		}
		port, err := strconv.Atoi(portStr)
		if err != nil {
			return fmt.Errorf("%w %q: invalid port: %w", ErrInvalidAddress, address, err)
		}
		c.Host = host
		c.Port = port
		return nil
	}
}

// WithReadTimeout sets the per-connection read timeout.
func WithReadTimeout(d time.Duration) ConfigOption {
	return func(c *Config) error {
		c.ReadTimeout = d
		return nil
	}
}

// WithWriteTimeout sets the per-connection write timeout.
func WithWriteTimeout(d time.Duration) ConfigOption {
	return func(c *Config) error {
		c.WriteTimeout = d
		return nil
	}
}

// WithIdleTimeout sets the per-connection idle timeout.
func WithIdleTimeout(d time.Duration) ConfigOption {
	return func(c *Config) error {
		c.IdleTimeout = d
		return nil
	}
}

// WithMaxConnections sets the maximum concurrent connection limit.
func WithMaxConnections(n int) ConfigOption {
	return func(c *Config) error {
		c.MaxConnections = n
		return nil
	}
}

// WithReadBufferSize sets the per-connection read buffer size in bytes.
func WithReadBufferSize(size int) ConfigOption {
	return func(c *Config) error {
		c.ReadBufferSize = size
		return nil
	}
}

// WithWriteBufferSize sets the per-connection write buffer size in bytes.
func WithWriteBufferSize(size int) ConfigOption {
	return func(c *Config) error {
		c.WriteBufferSize = size
		return nil
	}
}
