package server

import (
	"context"
	"database/sql"
	"errors"
	"net/http"
	"time"

	"vamos/internal/metrics"
)

const (
	StatusClientClosed = 499
	TIMEOUT_REQUEST    = time.Second * 1
)

type errHandler func(http.ResponseWriter, *http.Request) error

func (b *Backbone) eHand(f errHandler) http.HandlerFunc {
	return func(w http.ResponseWriter, req *http.Request) {
		err := f(w, req)
		if err != nil {
			switch {
			case errors.Is(err, context.Canceled):
				b.Logger.Error("HTTP", "status", StatusClientClosed)
			case errors.Is(err, context.DeadlineExceeded):
				b.Logger.Error("HTTP", "status", http.StatusRequestTimeout)
				http.Error(w, "timeout", http.StatusRequestTimeout)
			case errors.Is(err, sql.ErrNoRows):
				w.WriteHeader(http.StatusNoContent)
			default:
				b.Logger.Error("HTTP", "err", err.Error())
				http.Error(w, err.Error(), http.StatusInternalServerError)
			}
		}
	}
}

func addFeaturesV1(router *http.ServeMux, b *Backbone) {
	router.HandleFunc("GET /author/{surname}", b.eHand(b.readAuthor))
}

func (b *Backbone) readAuthor(w http.ResponseWriter, req *http.Request) error {
	metrics.ReadAuthorCounter.Inc()
	surname := req.PathValue("surname")
	timer, cancel := context.WithTimeout(req.Context(), TIMEOUT_REQUEST)
	defer cancel()
	result, err := b.FirstDB.GetAuthor(timer, surname)

	if err != nil {
		return err
	}

	w.Write([]byte(result.Name))
	return nil
}
