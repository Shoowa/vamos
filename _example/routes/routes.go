package routes

import (
	"context"
	"errors"
	"net/http"
	"time"

	"_example/metric"
	"_example/sqlc/data/first" // NOT generated in this example

	"github.com/Shoowa/vamos/router"
)

const TIMEOUT_REQUEST = time.Second * 3

// Deps can hold the Bundle struct to access the Router, and can hold the
// Queries struct to easily query the database.
type Deps struct {
	*router.Backbone
	Query *first.Queries // NOT generated in this example.
}

// Create a menu of routes.
func (d *Deps) GetEndpoints() []router.Endpoint {
	return []router.Endpoint{
		{"GET /test2", d.hndlr2},
		{"GET /readAuthorName/{surname}", d.readAuthorName},
	}
}

// Create a native struct that embeds the Backbone struct and adds sqlC Queries.
func WrapBackbone(b *router.Backbone) *Deps {
	d := &Deps{b, first.New(b.DbHandle)}
	return d
}

// CreateEmptyDeps is useful for creating the backboneWrapper in tests.
func CreateEmptyDeps() *Deps {
	return &Deps{nil, nil}
}

func (d *Deps) Hndlr1(w http.ResponseWriter, req *http.Request) {
	d.Logger.Info("Test1, test1, test1...")
	w.Write([]byte("This is a public service announcement."))
}

// Contrived error.
func (d *Deps) hndlr2(w http.ResponseWriter, req *http.Request) error {
	d.Logger.Info("2, 2, 2...")
	err := errors.New("Break! Failure! Error!")
	if err != nil {
		return err
	}
	w.Write([]byte("This is a 2nd PSA!"))
	return nil
}

// Handler that leverages sqlC Queries.
func (d *Deps) readAuthorName(w http.ResponseWriter, req *http.Request) error {
	metric.ReadAuthCount.Inc()
	surname := req.PathValue("surname")
	timer, cancel := context.WithTimeout(req.Context(), TIMEOUT_REQUEST)
	defer cancel()
	result, err := d.Query.GetAuthor(timer, surname)

	if err != nil {
		return err
	}

	w.Write([]byte(result.Name))
	return nil
}
