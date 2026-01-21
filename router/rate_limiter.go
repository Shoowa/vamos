package router

import (
	"net/http"

	"github.com/Shoowa/vamos/config"

	"golang.org/x/time/rate"
)

func CreateRateLimiter(cfg *config.RateLimiter) *rate.Limiter {
	avg := rate.Limit(cfg.Average)
	return rate.NewLimiter(avg, cfg.Burst)
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
