package router

import (
	"context"
	"net/http"
	"time"
)

const (
	TIMEOUT_REQUEST = time.Second * 1
)

func addFeaturesV1(router *http.ServeMux, b *Backbone) {
	router.HandleFunc("GET /author/{surname}", b.eHand(b.readAuthor))
}

func (b *Backbone) readAuthor(w http.ResponseWriter, req *http.Request) error {
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
