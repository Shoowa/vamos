package client

import (
	"crypto/tls"
	"log/slog"
	"net"
	"net/http"
	"time"

	"github.com/Shoowa/vamos/config"
	"golang.org/x/time/rate"
)

// RecorderTransport combines a custom http.Transport with a Logger and a Rate
// Limiter, and amends the RoundTrip method to log the duration of a request and
// any errors.
type RecorderTransport struct {
	Transport http.RoundTripper
	Control   *rate.Limiter
	Logger    *slog.Logger
}

// RoundTrip leverages the http.RoundTripper interface to add logging and rate
// limiting to any request. Potential improvement: Add atomic counters to record
// HTTP status codes. Also add a gauge to measure active TCP connections.
func (rt *RecorderTransport) RoundTrip(req *http.Request) (response *http.Response, err error) {
	// Abort request when tokens are depleted.
	err = rt.Control.Wait(req.Context())
	if err != nil {
		return nil, err
	}

	// Record basic request data.
	rt.Logger.Info("Request", "Method", req.Method, "URL", req.URL.String())

	// Schedule the logging of the response and any errors after the func exits
	// and before it returns.
	begin := time.Now()
	defer func() {
		elapsed := time.Since(begin)
		if err != nil {
			rt.Logger.Error("Failed request", "Method", req.Method, "URL", req.URL.String(), "Elapsed", elapsed, "Error", err.Error())
			return
		}
		if response.StatusCode != http.StatusOK {
			rt.Logger.Info("Response", "Status", response.StatusCode, "Method", req.Method, "URL", req.URL.String(), "Elapsed", elapsed)
		}
	}()

	response, err = rt.Transport.RoundTrip(req)
	return response, err
}

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

// CreateHTTPClient produces a client that can log requests, errors, and their
// duration.
func CreateHTTPClient(cfg *config.Config, logger *slog.Logger) *http.Client {
	cfgLimit := cfg.Client.RateLimiter
	window := rate.Every(time.Second / time.Duration(cfgLimit.Amount))
	limiter := rate.NewLimiter(window, cfgLimit.Burst)

	transporter := createHTTPTransport(cfg)
	specialTransporter := &RecorderTransport{
		Transport: transporter,
		Control:   limiter,
		Logger:    logger,
	}

	httpClient := &http.Client{
		Transport: specialTransporter,
		Timeout:   cfg.Client.TimeoutTotal * time.Second,
	}

	return httpClient
}
