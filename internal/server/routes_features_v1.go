package server

import (
	"context"
	"database/sql"
	"errors"
	"net/http"
	"time"
)

const (
	StatusClientClosed = 499
	TIMEOUT_REQUEST    = time.Second * 1
)

func addFeaturesV1(router *http.ServeMux, b *Backbone) {
	readAuthor := http.HandlerFunc(b.readAuthor)
	router.HandleFunc("GET /author/{surname}", readAuthor)
}

func (b *Backbone) readAuthor(w http.ResponseWriter, req *http.Request) {
	surname := req.PathValue("surname")
	timer, cancel := context.WithTimeout(req.Context(), TIMEOUT_REQUEST)
	defer cancel()
	result, err := b.FirstDB.GetAuthor(timer, surname)

	if err != nil {
		switch {
		case errors.Is(err, context.Canceled):
			b.Logger.Error("readAuthor", "status", StatusClientClosed)
		case errors.Is(err, context.DeadlineExceeded):
			b.Logger.Error("readAuthor", "status", http.StatusRequestTimeout)
			http.Error(w, "timeout", http.StatusRequestTimeout)
		case errors.Is(err, sql.ErrNoRows):
			w.WriteHeader(http.StatusNoContent)
		default:
			b.Logger.Error("readAuthor", "err", err.Error())
			http.Error(w, err.Error(), http.StatusInternalServerError)
		}
	}

	w.Write([]byte(result.Name))
}
