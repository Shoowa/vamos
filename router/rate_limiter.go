package router

import (
	"net/http"
	"time"

	"github.com/Shoowa/vamos/config"

	"golang.org/x/time/rate"
)

func CreateRateLimiter(cfg *config.RateLimiter) *rate.Limiter {
	window := time.Second / cfg.Window
	limit := rate.Every(window / cfg.Amount)
	return rate.NewLimiter(limit, cfg.Burst)
}

func Limit(limiter *rate.Limiter, next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if limiter.Allow() == false {
			http.Error(w, http.StatusText(429), http.StatusTooManyRequests)
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
