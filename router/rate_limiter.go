package router

import (
	"net/http"
	"strconv"
	"time"

	"github.com/Shoowa/vamos/config"

	"golang.org/x/time/rate"
)

const (
	HEADER_RATE_LIMIT     = "RateLimit-Limit"
	HEADER_RATE_REMAINING = "RateLimit-Remaining"
	HEADER_RATE_RETRY     = "Retry-After"
)

// CreateRateLimiter defines the amount of seconds and tokens permitted.
func CreateRateLimiter(cfg *config.RateLimiter) *rate.Limiter {
	window := time.Second * cfg.Window
	limit := rate.Every(window / cfg.Amount)
	return rate.NewLimiter(limit, cfg.Burst)
}

// Limit is a closure that configures a middleware func that responds with
// headers to help clients manage the rate of requests.
func Limit(limiter *rate.Limiter, next http.Handler) http.Handler {
	amount := strconv.Itoa(limiter.Burst())

	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		available := strconv.FormatFloat(limiter.Tokens(), 'f', 0, 64)
		w.Header().Set(HEADER_RATE_LIMIT, amount)
		w.Header().Set(HEADER_RATE_REMAINING, available)

		if limiter.Allow() == false {
			reservation := limiter.Reserve()
			delay := reservation.Delay()
			if delay > 0 {
				reservation.Cancel()
				retryTime := strconv.FormatFloat(delay.Seconds(), 'f', 0, 64)
				w.Header().Set(HEADER_RATE_RETRY, retryTime)
			}
			http.Error(w, http.StatusText(http.StatusTooManyRequests), http.StatusTooManyRequests)
			return
		}

		next.ServeHTTP(w, r)
	})
}

func optionalGlobalRateLimiter(cfg *config.RateLimiter, next http.Handler) http.Handler {
	if cfg.Active == false {
		return next
	}

	limiter := CreateRateLimiter(cfg)
	return Limit(limiter, next)
}
