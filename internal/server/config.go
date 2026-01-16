package server

import (
	"errors"
	"fmt"
	"time"
)

type Config struct {
	Host            string
	Port            int
	ReadTimeout     time.Duration
	WriteTimeout    time.Duration
	IdleTimeout     time.Duration
	MaxConnections  int
	ReadBufferSize  int
	WriteBufferSize int
}

func NewDefaultConfig() *Config {
	return &Config{
		Host:            "0.0.0.0",
		Port:            6379,
		ReadTimeout:     0,
		WriteTimeout:    0,
		IdleTimeout:     0,
		MaxConnections:  0,
		ReadBufferSize:  1024 * 4, // 4 KB
		WriteBufferSize: 1024 * 4, // 4 KB
	}
}

func (c *Config) Address() string {
	return fmt.Sprintf("%s:%d", c.Host, c.Port)
}

func (c *Config) Validate() error {
	if c.Port < 0 || c.Port > 65535 {
		return errors.New("port must be between 0 and 65535")
	}
	if c.ReadBufferSize < 0 {
		return errors.New("read buffer size must be non-negative")
	}
	if c.WriteBufferSize < 0 {
		return errors.New("write buffer size must be non-negative")
	}
	if c.MaxConnections < 0 {
		return errors.New("max connection must be non negative")
	}
	return nil
}

func (c *Config) WithHost(host string) *Config {
	c.Host = host
	return c
}

func (c *Config) WithPort(port int) *Config {
	c.Port = port
	return c
}

func (c *Config) WithReadTimeout(readTimeout time.Duration) *Config {
	c.ReadTimeout = readTimeout
	return c
}

func (c *Config) WithWriteTimeout(writeTimeout time.Duration) *Config {
	c.WriteTimeout = writeTimeout
	return c
}

func (c *Config) WithIdleTimeout(idleTimeout time.Duration) *Config {
	c.IdleTimeout = idleTimeout
	return c
}

func (c *Config) WithMaxConnections(maxConnections int) *Config {
	c.MaxConnections = maxConnections
	return c
}

func (c *Config) WithReadBufferSize(readBufferSize int) *Config {
	c.ReadBufferSize = readBufferSize
	return c
}

func (c *Config) WithWriteBufferSize(writeBufferSize int) *Config {
	c.WriteBufferSize = writeBufferSize
	return c
}
