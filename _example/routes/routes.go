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

// EquipRouter conveniently wraps the Backbone in a new struct named Deps.
func EquipRouter(b *router.Backbone) http.Handler {
	rtr := router.NewRouter(b)
	d := &Deps{b, first.New(b.DbHandle)} // NOT generated in this example.

	// Attach HTTP Handler directly to the http.ServeMux
	rtr.Router.HandleFunc("GET /test1", d.hndlr1)

	// Or attach HTTP errHandlers via a convenient func.
	rm := routeMenu(d)
	rtr.AddRoutes(rm, b)

	return rtr
}

// Create a list of the library's Endpoints. Easy to write, easy to read.
func routeMenu(d *Deps) []router.Endpoint {
	return []router.Endpoint{
		{"GET /test2", d.hndlr2},
		{"GET /readAuthorName/{surname}", d.readAuthorName},
	}
}

func (d *Deps) hndlr1(w http.ResponseWriter, req *http.Request) {
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
