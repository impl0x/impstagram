package config

import (
	"time"
)

// contains http *Requests* and client config, not webserver or rest api

type HttpConfig struct {
	DialerTimeout         time.Duration `env:"HTTP_DIALER_TIMEOUT"`
	DialerKeepAlive       time.Duration `env:"HTTP_DIALER_KEEP_ALIVE"`
	TLSHandshakeTimeout   time.Duration `env:"HTTP_TLS_HANDSHAKE_TIMEOUT"`
	ResponseHeaderTimeout time.Duration `env:"HTTP_RESPONSE_HEADER_TIMEOUT"`
	ExpectContinueTimeout time.Duration `env:"HTTP_EXPECT_CONTINUE_TIMEOUT"`
	MaxIdleConns          int           `env:"HTTP_MAX_IDLE_CONNS" validate:"min=1"`
	MaxIdleConnsPerHost   int           `env:"HTTP_MAX_IDLE_CONNS_PER_HOST" validate:"min=1"`
	MaxConnsPerHost       int           `env:"HTTP_MAX_CONNS_PER_HOST" validate:"min=1"`
	IdleConnTimeout       time.Duration `env:"HTTP_IDLE_CONN_TIMEOUT"`
	ForceAttemptHTTP2     bool          `env:"HTTP_FORCE_ATTEMPT_HTTP2" validate:"min=1"`
	Timeout               time.Duration `env:"HTTP_TIMEOUT"`
}

var DefaultHttpConfig = HttpConfig{
	DialerTimeout:         5 * time.Second,  // Time to establish a TCP connection
	DialerKeepAlive:       30 * time.Second, // Keep TCP connection alive
	TLSHandshakeTimeout:   5 * time.Second,  // Max time waiting for TLS handshake
	ResponseHeaderTimeout: 5 * time.Second,  // Max time waiting for headers after sending request
	ExpectContinueTimeout: 1 * time.Second,
	MaxIdleConns:          100,              // Total idle connections across all hosts
	MaxIdleConnsPerHost:   100,              // Boost from default 2 to prevent frequent reconnects
	MaxConnsPerHost:       100,              // Absolute limit to prevent overwhelming a single service
	IdleConnTimeout:       90 * time.Second, // Time before closing unused idle connections
	ForceAttemptHTTP2:     true,
	Timeout:               10 * time.Second, // Hard maximum timeout for the entire request-response life cycle
}
