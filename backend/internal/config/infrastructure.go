package config

import "time"

// Contains infrastructure related configs such as database, webserver, etc.
type Infrastructure struct {
	HttpClient HttpClient // http requests client
	HttpServer HttpServer // http server api
	Postgres   Postgres   // database
	Email      Email      // email sender
}

// http client config for sending requests
type HttpClient struct {
	DialerTimeout         time.Duration `env:"HTTP_CLIENT_DIALER_TIMEOUT"`
	DialerKeepAlive       time.Duration `env:"HTTP_CLIENT_DIALER_KEEP_ALIVE"`
	TLSHandshakeTimeout   time.Duration `env:"HTTP_CLIENT_TLS_HANDSHAKE_TIMEOUT"`
	ResponseHeaderTimeout time.Duration `env:"HTTP_CLIENT_RESPONSE_HEADER_TIMEOUT"`
	ExpectContinueTimeout time.Duration `env:"HTTP_CLIENT_EXPECT_CONTINUE_TIMEOUT"`
	MaxIdleConns          int           `env:"HTTP_CLIENT_MAX_IDLE_CONNS" validate:"min=1"`
	MaxIdleConnsPerHost   int           `env:"HTTP_CLIENT_MAX_IDLE_CONNS_PER_HOST" validate:"min=1"`
	MaxConnsPerHost       int           `env:"HTTP_CLIENT_MAX_CONNS_PER_HOST" validate:"min=1"`
	IdleConnTimeout       time.Duration `env:"HTTP_CLIENT_IDLE_CONN_TIMEOUT"`
	ForceAttemptHTTP2     bool          `env:"HTTP_CLIENT_FORCE_ATTEMPT_HTTP2"`
	Timeout               time.Duration `env:"HTTP_CLIENT_TIMEOUT"`
}

var DefaultHttpClient = HttpClient{
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

type HttpServer struct {
	Port              string        `env:"HTTP_SERVER_PORT,required"`
	ReadTimeout       time.Duration `env:"HTTP_SERVER_READ_TIMEOUT"`
	ReadHeaderTimeout time.Duration `env:"HTTP_SERVER_READ_HEADER_TIMEOUT"`
	WriteTimeout      time.Duration `env:"HTTP_SERVER_WRITE_TIMEOUT"`
	IdleTimeout       time.Duration `env:"HTTP_SERVER_IDLE_TIMEOUT"`
}

var DefaultHttpServer = HttpServer{
	ReadTimeout:       5 * time.Second,
	ReadHeaderTimeout: 2 * time.Second,
	WriteTimeout:      10 * time.Second,
	IdleTimeout:       120 * time.Second,
}

type Postgres struct {
	PoolMax uint32 `env:"PG_POOL_MAX"`
	URL     string `env:"PG_URL,required" validate:"url"`
}

var DefaultPostgres = Postgres{
	PoolMax: 25,
}

type Email struct {
	ResendApiKey string `env:"EMAIL_RESEND_API_KEY,required" validate:"startswith=re_"`
	EmailID      string `env:"EMAIL_ID,required" validate:"email"`
}
