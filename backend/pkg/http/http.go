// contains http client for requests not webserver or rest api
package http

import (
	"backend/internal/config"
	"net"
	"net/http"
)

// small wrapper around the native *http.Client
type Client struct {
	*http.Client
}

func NewConfiguredClient(cfg *config.HttpConfig) Client {
	dialer := &net.Dialer{
		Timeout:   cfg.DialerTimeout,
		KeepAlive: cfg.DialerKeepAlive,
	}
	transport := &http.Transport{
		DialContext:           dialer.DialContext,
		TLSHandshakeTimeout:   cfg.TLSHandshakeTimeout,
		ResponseHeaderTimeout: cfg.ResponseHeaderTimeout,
		ExpectContinueTimeout: cfg.ExpectContinueTimeout,

		MaxIdleConns:        cfg.MaxIdleConns,
		MaxIdleConnsPerHost: cfg.MaxIdleConnsPerHost,
		MaxConnsPerHost:     cfg.MaxConnsPerHost,
		IdleConnTimeout:     cfg.IdleConnTimeout,
		ForceAttemptHTTP2:   cfg.ForceAttemptHTTP2,
	}
	return Client{
		&http.Client{
			Transport: transport,
			Timeout:   cfg.Timeout,
		},
	}
}

func NewDefaultClient() Client {
	return Client{http.DefaultClient}
}
