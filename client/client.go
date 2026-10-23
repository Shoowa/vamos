package client

import (
	"crypto/tls"
	"net"
	"net/http"
	"time"

	"github.com/Shoowa/vamos/config"
)

func createHTTPTransport(cfg *config.Config) *http.Transport {
	c := cfg.Client

	tlsCfg := &tls.Config{
		InsecureSkipVerify: false,
		MinVersion:         tls.VersionTLS12,
		MaxVersion:         tls.VersionTLS13,
	}

	dialer := &net.Dialer{
		// Max time to establish a TCP connection.
		Timeout: cfg.Client.TimeoutTcpSetup * time.Second,

		// Interval between Keep-Alive probes.
		KeepAlive: cfg.Client.KeepAliveInterval * time.Second,
	}

	transporter := &http.Transport{
		DialContext: (dialer).DialContext,

		// Await secure handshake.
		TLSHandshakeTimeout: cfg.Client.TimeoutTls * time.Second,

		// Await response headers after sending a request.
		ResponseHeaderTimeout: c.TimeoutRespHeader * time.Second,

		// Connection pool rules.
		MaxConnsPerHost:     c.MaxConnsPerHost,
		MaxIdleConnsPerHost: c.MaxIdleConnsPerHost,
		MaxIdleConns:        c.MaxIdleConns,

		// Max time an un-used connection sits in the pool.
		IdleConnTimeout: time.Second * c.IdleConnTimeout,

		// Max MegaBytes accepted from a response header.
		MaxResponseHeaderBytes: 1024 * 1024 * c.MaxResponseHeaderMegaBytes,

		TLSClientConfig:       tlsCfg,
		ExpectContinueTimeout: 1 * time.Second,
		ForceAttemptHTTP2:     true,
	}

	return transporter
}
