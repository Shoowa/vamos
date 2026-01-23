# ¡Vamos!
[![Go Reference](https://pkg.go.dev/badge/github.com/Shoowa/vamos.svg)](https://pkg.go.dev/github.com/Shoowa/vamos)

A library for a Go HTTP server. It is configured with TLS 1.3, rate limiting,
logging, metrics, health checks, & profiling. It is integrated with Openbao,
Postgres, & Redis.

A virtual [development environment](#development-environment) is included in this repository.

A corporate development team can deploy a prototype into a production
environment as a micro-service and expect operational maturity. _Vamos_ hastens
development and eases operation.


## Quick Start
Provide the application a config file named _dev.json_ or _prod.json_ in the
_config_ directory. View the *_example/config/dev.json*. The file is concerned
with the following:
1. The location of the server guarding secrets.
    - Local file paths to read x509 cert & key, & intermediate CA.
2. The location of a Postgres instance and its sensitive credential.
3. Fake data to provide to Postgres for development & testing.
4. Server details, e.g., host, port, timeouts.
    - TLS config as server
    - TLS config as client
    - Optional rate limiter
5. Health evaluations.
6. Logging level.
7. Toggling of metrics.
8. Location of a Redis server.
    - Openbao HTTP endpoint & JSON key for password.
9. Fake data for a local Postgres server.

#### Sequence of Database List
Notice _data.relational_ in *_example/config/dev.json* is an array. The sequence
is preserved after the configuration is read. Accessing a database requires
acknowledging its position in the array. In the example, the command to connect
to a database includes a reference to its position in the array.
```go
// _example/main.go
package main
// abbreviated for clarity...

const DB_FIRST = 0

func main() {
	cfg := config.Read()

	db1, _ := rdbms.ConnectDB(cfg, DB_FIRST)
```

### TLS Configuration
Notice _httpserver.tls_server_ and _httpserver.tls_client_ represent different
sets of certificates and keys in a _TlsSecret_ struct. The former is for the Go
application to establish TLS connections with clients, and the latter is for
mutual TLS as a client inside a corporate network. The former will be used to
create a X509 Certificate that will be included in the TLS configuration of
_http.Server_. The latter can be used as a X509 Certificate in a _Redis_ client,
etc.

The field *httpserver.tls_server.cert_path* represents a HTTP endpoint offered
by _OpenBao_, and *httpserver.tls_server.cert_field* represents a JSON key in
the data read from _OpenBao_.

The _SkeletonKey_ from the _Secrets_ package can easily read sensitive data from
_OpenBao_ and transform it into a useful X509 certificate for any developer.

#### Overloaded TlsSecret struct
A notable problem with the current usage of the _TlsSecret_ struct is that I'm
forcing it to perform double-duty. When configuring the _Openbao_ client, the
_TlsSecret_ field named *cert_path* is simply a file path, so that the client
can read a local _.pem_ file to build a TLS connection to the secrets storage.
When configuring other clients to communicate with _Postgres_ and _Redis_, then
the fields represent HTTP endpoints hosted on the Openbao server. So an
executable reads locally hosted ._pem_ files to build a secure connection to the
Openbao server, then reads subsequent secrets from Openbao to build secure
connections to other servers. I should probably reform the _httpServer_ struct
and the _TlsSecret_ struct.

### Local dev Openbao NOT rotating certs
The local dev Openbao isn't rotating X509 certificates. I should probably employ
that feature, but currently I simply write certificates into secrets storage.


#### Build
After all that is defined, determine the version number of the application. This
is a good opportunity to include a tool that reads the Git Log and interprets
Conventional Commits to determine the version.

Provide two environmental variables: One to define whether this deployment
exists in development or production, and another to offer the Openbao access
token for secrets storage.
```bash
# Start the Dev Environment with the included makefile. See next section.
~/your_app $ go env -w GOEXPERIMENT=greenteagc
~/your_app $ go build -v -ldflags="-s -X 'github.com/Shoowa/vamos/config.AppVersion=v.0.0.0' " -o yourapp
~/your_app $ APP_ENV=DEV OPENBAO_TOKEN=token ./yourapp
```

## Development Environment
This is for MacOS. You will need two things: _Podman_ and _Golang_.
```bash
~/vamos $ make podman_create_vm
~/vamos $ podman ps -a
```
You will receive three things in this dev env.
1. Openbao to hold passwords and certificates. This can be improved to handle
   cert-rotation.
2. Postgres to permanently hold data.
3. Redis to temporarily hold data.

You will receive a new instance of Postgres with a user and database, and an
instance of Openbao with a loaded password kept at _dev-postgres-test_. That
path matches the config field _data.relational.[0].secret_ in the *_example/*.

Postgres & Openbao will need a few minutes to start.

A natively installed instance of Postgres is fine when it is the only
dependency, but I imagine anyone using this will have an existing installation
of Postgres configured for a different development context. We can use Postgres
inside a virtual machine to avoid disruptions. And we can add other databases
and dependencies.

A virtual machine managed by _podman_[^p1] will host databases needed by the
application. The virtual machine runs Linux, specifically Fedora CoreOS.[^p2]
And _systemD_ will manage containers hosting databases.

The included _makefile_ offers a command that copies a few _.container_ files
from a directory named *_linux/* to a new directory on the MacOS host. And
copies a _.sql_ initilization script for Postgres. Then uses _podman_ to create
a virtual machine named *dev_vamos* that can read the new directory. Then uses
_systemD_ to fetch container images and run them. And setup the Postgres
instance in a container.

Instead of using _podman_ commands to manipulate the containers directly, we can
use _systemD_ inside the Linux virtual machine to start and stop containers.
```bash
~/vamos $ podman machine ssh dev_vamos "systemctl --user status dev_postgres"
● dev_postgres.service - Launch Postgres 18 with native UUIDv7
     Loaded: loaded (/var/home/core/.config/containers/systemd/dev_postgres.container; generated)
    Drop-In: /usr/lib/systemd/user/service.d
             └─10-timeout-abort.conf
     Active: active (running) since Fri 2025-07-04 09:49:39 EDT; 6s ago
 Invocation: 3b0202c669c640a1a6a96bd8bab6f4d5
   Main PID: 9034 (conmon)
      Tasks: 24 (limit: 2155)
     Memory: 39.4M (peak: 55.5M)
        CPU: 287ms
     CGroup: /user.slice/user-501.slice/user@501.service/app.slice/dev_postgres.service
             ├─libpod-payload-4f575acdf6c9155ee2e079ba37c9220e9aef7bb47430af6c9ad969d26cf12d30
             │ ├─9036 postgres
             │ ├─9062 "postgres: io worker 1"
             │ ├─9063 "postgres: io worker 0"
             │ ├─9064 "postgres: io worker 2"
             │ ├─9065 "postgres: checkpointer "
             │ ├─9066 "postgres: background writer "
             │ ├─9068 "postgres: walwriter "
             │ ├─9069 "postgres: autovacuum launcher "
             │ └─9070 "postgres: logical replication launcher "
             └─runtime
               ├─9018 rootlessport
               ├─9025 rootlessport-child
               └─9034 /usr/bin/conmon --api-version 1 # removed for brevity
```

Connect to the database named *test_data* in the containerized Postgres instance
from the MacOS host.
```bash
~/vamos $ psql -h localhost -U tester -d test_data
```

Inspect the condition of Openbao and whether or not it received a password.
```bash
~/vamos podman machine ssh dev_vamos "systemctl --user status secrets.target dev_openbao openbao_add_pw"
```

Change the password archived in Openbao as much as you want.
```bash
# httpie command
~/vamos http POST :8200/v1/secret/data/dev-postgres-test X-Vault-Token:token Content-Type:application/json data:='{ "password": "openbao777" }'
```

### Postgres Database
A container image of Postgres 18 Beta is preferred for the native _UUIDv7_
feature. How is a container obtained and managed by _podman_ in this development
environment?

A special *.container* file is read from a user directory named
*.config/containers/systemd/* in the VM by a _podman_ tool named _quadlet_. And
_quadlet_ parses the file to produce a *systemD service* file. The resulting
*.service* file can download a container image and run it. More details can be
studied in the _makefile_ under the command _podman_create_vm_.

The _quadlet .container_ file includes a few commands commonly used to run
containers in both _Docker_ and _podman_.
```bash
# _linux/dev_postgres.container
[Unit]
Description=Launch Postgres 18 with native UUIDv7

[Container]
Image=docker.io/library/postgres:18beta2-alpine3.22
ContainerName=postgres
Environment=POSTGRES_PASSWORD=password
Environment=POSTGRES_USERNAME=postgres
Environment=POSTGRES_HOST_AUTH_METHOD=trust
PublishPort=5432:5432
Volume=/data/postgres:/var/lib/postgresql/18/docker
Volume=/data/setup/setup_db1.sql:/docker-entrypoint-initdb.d/setup_db1.sql
PidsLimit=100

[Service]
Restart=on-failure
RestartSec=10

[Install]
RequiredBy=databases.target
```

The *_example/testdata/setup_db1.sql* file will be copied from the project on the host to
the volume of the virtual machine, then mounted to the Postgres container.
Postgres only reads this file once during its initilization. It will skip
reading it whenever the container is started again.
```sql
-- _example/testdata/setup_db1.sql
DROP DATABASE IF EXISTS test_data;
CREATE DATABASE test_data;
CREATE USER tester WITH PASSWORD 'password';

\c test_data
GRANT ALL ON SCHEMA public TO tester;
```
Notice the command to switch from the default database to the newly created
*test_data* database. The default user must be in the latter database to
effectively grant privileges to another account.

To launch the Postgres development instance, simply ssh into the _podman_
virtual machine and order _systemD_ to start the service. Logs can be viewed via
_journalD_.
```bash
~/ $ podman machine ssh dev_vamos "systemctl --user start dev_postgres"
~/ $ podman machine ssh dev_vamos "journalctl --user -u dev_postgres"
```
The extension _.service_ is excluded from the commands for brevity.

## Database Tooling
A couple of CLI tools that won't be imported into the application.
```bash
~/vamos $ go install -tags 'postgres' github.com/golang-migrate/migrate/v4/cmd/migrate@latest
~/vamos $ go install github.com/sqlc-dev/sqlc/cmd/sqlc@latest
```

### Database Migration
The CLI tool _migrate_ creates numbered _.sql_ files that we can fill in
with SQL commands. Then it applies them in numbered order to a Postgres
database.[^d1]

Create a _.sql_ file that will hold the commands to create a table named
_authors_.
```bash
~/vamos/_example $ migrate create -ext sql -dir sqlc/migrations/first -seq create_authors
~/vamos/_example $ tree sqlc/migrations/first
sqlc/migrations/first
├── 000001_create_authors.down.sql
├── 000001_create_authors.up.sql
```
In _000001_create_authors.up.sql_, write the following SQL commands:
```sql
CREATE TABLE IF NOT EXISTS authors (
    id UUID DEFAULT uuidv7() PRIMARY KEY,
    name text NOT NULL,
    bio text
);
```

After writing a SQL command to create a table, apply the command. Notice the
subdirectory associated with a particular database, in this case _first_.
Notice the keyword _up_ as the final token in the command.
```bash
~/vamos/_example $ export TEST_DB=postgres://tester@localhost:5432/test_data?sslmode=disable
~/vamos/_example $ migrate -database $TEST_DB -path sqlc/migrations/first up
```
The creation of any tables and any adjustments offered by _*.up.sql_ can be
reversed by following the SQL commands written in _*.down.sql_ files.

### Database Code Generation
The command line tool _sqlC_ reads _.sql_ files and writes Go code we can
import into the application.[^d2]

```yaml
# sqlc/sqlc.yaml
version: "2"
sql:
  - engine: "postgresql"
    queries: "queries/first"
    schema: "migrations/first"
    gen:
      go:
        package: "first"
        out: "data/first"
        sql_package: "pgx/v5"
        emit_json_tags: true
```

In _sqlc/sqlc.yaml_, one or more database engines can be listed to help the Go
application connect to two different Postgres databases. Each entry relies on a
directory of _.sql_ files written for queries, and a directory of _.sql_ files
named _migrations_ written for creating tables. _sqlC_ reads these files as
inputs.

The produced code will reside in the _first_ package in a newly created
subdirectory named _data/first_ and another package can reside in a separate
subdirectory, i.e., _data/second_. The code will use the _pgx/v5_ driver, and
include JSON tags in the fields of the generated structs that represent data
entities.

After we draft a _.sql_ file for a hypothetical table of _authors_, like so:
```sql
-- sqlc/migrations/first/000001_create_authors.up.sql
CREATE TABLE IF NOT EXISTS authors (
    id UUID DEFAULT uuidv7() PRIMARY KEY,
    name text NOT NULL,
    bio text
);
```

We can execute the command to create Go code that will interact with the
Postgres database.
```bash
~/vamos/_example $ sqlc generate -f sqlc/sqlc.yaml
```

The tool _sqlC_ produces the following code in a _models.go_ file:
```go
// sqlc/data/first/models.go
// Code generated by sqlc. DO NOT EDIT.
// versions:
//   sqlc v1.28.0
package first
// abbreviated for clarity...

type Author struct {
	ID   pgtype.UUID `json:"id"`
	Name string      `json:"name"`
	Bio  pgtype.Text `json:"bio"`
}
```

_Author_ will be accessble in a method of a struct named _Queries_.
```go
// sqlc/data/first/authors.sql.go
// Code generated by sqlc. DO NOT EDIT.
// versions:
//   sqlc v1.28.0
// source: authors.sql
package first
// abbreviated for clarity...

const getAuthor = `-- name: GetAuthor :one
SELECT id, name, bio FROM authors WHERE name = $1 LIMIT 1

func (q *Queries) GetAuthor(ctx context.Context, name string) (Author, error) {
	row := q.db.QueryRow(ctx, getAuthor, name)
	var i Author
	err := row.Scan(&i.ID, &i.Name, &i.Bio)
	return i, err
}
```

And _Queries_ is generated in _sqlc/data/first/db.go_. It holds the database
handle, i.e., the connection pool.
```go
// sqlc/data/first/db.go
// Code generated by sqlc. DO NOT EDIT.
// versions:
//   sqlc v1.28.0
package first
// abbreviated for clarity...

func New(db DBTX) *Queries {
	return &Queries{db: db}
}
```

The Postgres connection pool created in _main()_ is transferred to _Backbone_
when configuring the *Backbone* with the _Options pattern_.[^o1]
```go
// _example/main.go
package main
// abbreviated for clarity...

func main() {
	db1, _ := rdbms.ConnectDB(cfg, DB_FIRST)

	backbone := router.NewBackbone(
		router.WithLogger(srvLogger),
		router.WithDbHandle(db1),
		router.WithCache(redis),
	)
}
```

The _Backbone struct_ holds the dependencies needed by the HTTP Handlers. It
resides in the _Router_ package.
```go
// router/backbone.go
package router
// abbreviated for clarity...

func WithDbHandle(dbHandle *pgxpool.Pool) Option {
	return func(b *Backbone) {
		b.DbHandle = dbHandle
	}
}

func WithCache(client *redis.Client) Option {
	return func(b *Backbone) {
		b.Cache = client
	}
}
```

In a downstream executable that imports this library and leverages _sqlC_, the
database handle will need to be transferred to the _*Queries_ struct and held
inside a wrapper.
```go
// _example/routes/routes.go
package routes
// abbreviated for clarity...

type Deps struct {
	*router.Backbone
	Query *first.Queries
}

func WrapBackbone(b *router.Backbone) *Deps {
	d := &Deps{b, first.New(b.DbHandle)}
	return d
}
```




## Develop
Create a feature with an existing SQL Table by following this process:
1. Draft a SQL query.
2. Generate Go code in _sqlc/data/_ based on the new SQL.
3. Draft a new HTTP Handler.
4. Register the new HTTP Handler with the Router.
5. Add a log line.
6. Add a metric line


### Draft A SQL Query
In the directory *_example/sqlc/queries/first*, add a file named _authors.sql_, then
write this inside it.
```sql
-- name: GetAuthor :one
SELECT * FROM authors WHERE name = $1 LIMIT 1;
```
Then use _sqlC_ to transform that SQL query into Go code.
```bash
~/vamos/_example $ sqlc generate -f sqlc/sqlc.yaml
```

_sqlC_ will read the comment, then create a _const_ with that name, and assign a
query to it. Then it will create a method with the same name that uses the
_const_.
```go
// sqlc/data/first/authors.sql.go
// Code generated by sqlc. DO NOT EDIT.
// versions:
//   sqlc v1.28.0
// source: authors.sql

package first

const getAuthor = `-- name: GetAuthor :one
SELECT id, name, bio FROM authors WHERE name = $1 LIMIT 1
`

func (q *Queries) GetAuthor(ctx context.Context, name string) (Author, error) {
	row := q.db.QueryRow(ctx, getAuthor, name)
	var i Author
	err := row.Scan(&i.ID, &i.Name, &i.Bio)
	return i, err
}
```
The method _GetAuthor()_ can be invoked inside an HTTP handler.

### HTTP Handlers, Databases, & Errors
Developers can focus on the package _router_ to create RESTful features.

Dependency injection is the technique used to provide database handles to the
HTTP handlers on the web server. Handlers are simply methods of the struct
_Backbone_, or methods of the struct wrapping _Backbone_ in a downstream
executable. Access a Postgres database in the field _DbHandle_ or through a
_Queries_ struct residing in the wrapper built in a downstream executable.

A Backbone method named _ServerError_ has been created to easily respond to
errant HTTP requests.
```go
// router/backbone.go
package router
// abbreviated for clarity...

func (b *Backbone) ServerError(w http.ResponseWriter, r *http.Request, err error) {
	method := r.Method
	path := r.URL.Path

	switch {
	case errors.Is(err, context.Canceled):
		b.Logger.Warn("HTTP", "status", StatusClientClosed, "method", method, "path", path)
	case errors.Is(err, context.DeadlineExceeded):
		b.Logger.Error("HTTP", "status", http.StatusGatewayTimeout, "method", method, "path", path)
		http.Error(w, "timeout", http.StatusGatewayTimeout)
	case errors.Is(err, sql.ErrNoRows):
		w.WriteHeader(http.StatusNoContent)
	default:
		b.Logger.Error("HTTP", "err", err.Error(), "method", method, "path", path)
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}
```

Error handling can be invoked in an executable's http.Handler like this:
```go
// _example/routes/routes.go
package routes
// abbreviated for clarity...

func (d *Deps) readAuthorName(w http.ResponseWriter, req *http.Request) {
	surname := req.PathValue("surname")

	timer, cancel := context.WithTimeout(req.Context(), TIMEOUT_REQUEST)
	defer cancel()

	result, err := d.Query.GetAuthor(timer, surname)

    // Pass err to the ServerError method, and return early.
	if err != nil {
        d.ServerError(w, req, err)
        return
	}

	w.Write([]byte(result.Name))
}
```


### Add New http.Handler to Router
In a downstream executable, add a method named _GetEndpoints()_ to the custom
dependency struct that wraps the _Backbone_ to conform to the library interface
_Gatherer_. This is required for the router to adopt routes written in the
executable.

Select the HTTP method that is most appropriate for the writing and reading of
data. The ability to select _GET_ or _POST_ as an argument in parameter
_pattern_ is a new feature of the language in version 1.22.[^r1]
```go
// _example/routes/routes.go
package routes
// abbreviated for clarity...

type Deps struct {
	*router.Backbone
	Query *first.Queries // NOT generated in this example.
}

func (d *Deps) GetEndpoints() []router.Endpoint {
	return []router.Endpoint{
		{"GET /test2", d.hndlr2},
		{"GET /readAuthorName/{surname}", d.readAuthorName},
	}
}
```

### Developer Logs
Inside a http.Handler, record errors and extra data by simply invoking the
_Logger_ residing in the _Backbone_ struct.

This is how a hypothetical http.Handler drafted in the library looks. Notice it
can directly access a _Backbone_ field.
```go
package router
// abbreviated for clarity...

func (b *Backbone) doSomething(w http.ResponseWriter, req *http.Request) {
	timer, cancel := context.WithTimeout(req.Context(), TIMEOUT_REQUEST)
	defer cancel()

	result, err := b.DbHandle.Ping(timer)

	if err != nil {
        d.Logger.Error("big_message", "err", err.Error())
        d.ServerError(w, req, err)
        return
	}

    b.Logger.Info("Did something important... but we should silently succeed.")
	w.Write([]byte("ok"))
}
```

This is a hypothetical http.Handler drafted in a downstream executable that
imports the library. It is a method on a struct named _Deps_ that wraps around
the _Backbone_. And _Deps_ holds a sqlC generated _Queries_ struct in a custom
field conveniently named _Query_.
```go
//routes/features_v1.go
package routes
// abbreviated for clarity...
import (
	"_example/sqlc/data/first" // sqlC generated code
	"github.com/Shoowa/vamos/router"
)

type Deps struct {
	*router.Backbone
	Query *first.Queries
}

d := &Deps{backbone, first.New(backbone.DbHandle)} // add sqlC *Queries struct

func (d *Deps) readAuthorName(w http.ResponseWriter, req *http.Request) {
	surname := req.PathValue("surname")

	timer, cancel := context.WithTimeout(req.Context(), TIMEOUT_REQUEST)
	defer cancel()

	result, err := d.Query.GetAuthor(timer, surname)

	if err != nil {
        d.ServerError(w, req, err)
        return
	}

	w.Write([]byte(result.Name))
}
```

### Metrics
Metrics are created by _Prometheus_ in the package _metrics_ and scraped on the
endpoint _/metrics_. The package captures go runtime metrics, e.g.,
*go_threads*, *go_goroutines*, etc.[^m2]

A convenient function for creating a Counter and registering it is available to
the downstream consumer of this library. Simply provide a name and description
for the Counter.
```go
// metrics/metrics.go
package metrics

import "github.com/prometheus/client_golang/prometheus"

func CreateCounter(name string, help string) prometheus.Counter {
	opts := prometheus.CounterOpts{
		Name: name,
		Help: help,
	}

	counter := prometheus.NewCounter(opts)
	prometheus.MustRegister(counter)
	return counter
}
```

New metrics need to be created in the executable, so they can be imported by a
HTTP Handler. This example shows an executable package named _metric_ that
imports the library's _metrics_ package. No creative naming here.
```go
package metric

import "github.com/Shoowa/vamos/metrics"

var ReadAuthCount = metrics.CreateCounter("read_authorSurname_count", "no_help")
```

Then the local _Counter_ is imported into the executable _routes_ package.
```go
// _example/routes/routes.go
package routes
// abbreviated for clarity...

import "metric/metric"

func (d *Deps) readAuthorName(w http.ResponseWriter, req *http.Request) {
    metric.ReadAuthCount.Inc()
    surname := req.PathValue("surname")
    // skipping body...
}
```

Observe the new data on the _/metrics_ route.
```bash
~/vamos $ curl localhost:8080/metrics
# abbreviated for clarity...
# TYPE promhttp_metric_handler_requests_total counter
promhttp_metric_handler_requests_total{code="200"} 0
promhttp_metric_handler_requests_total{code="500"} 0
promhttp_metric_handler_requests_total{code="503"} 0
# HELP read_authorSurname_count no_help
# TYPE read_authorSurname_count counter
read_authorSurname_count 0
```


### Health Record
Applications usually receive a request for a health status, then perform some
logic to evaluate the health of the application and the health of any
dependencies, then answer. That flow of events doesn't happen in this web app.

Instead, the web server responds to any request for health by simply reading
from a custom struct named _Health_ that resides in _Backbone_.
```go
// router/operations.go
package router
// abbreviated for clarity...

type Health struct {
	Rdbms    bool
	Heap     bool
}
```
_Health_ has several _boolean_ fields. Any request for the status of health is
answered by a method that reads from these fields and evaluates the totality of
the _boolean_ conditions.
```go
// router/operations.go
package router
// abbreviated for clarity...

func (h *Health) PassFail() bool {
	return h.Rdbms && h.Heap
}
```

The answer is then provided as a HTTP Header -- either 204 or 503.
```go
// router/routes_operations.go
package router
// abbreviated for clarity...

func (b *Backbone) Healthcheck(w http.ResponseWriter, r *http.Request) {
	status := b.Health.PassFail()

	if status {
		w.WriteHeader(http.StatusNoContent)
	} else {
		b.Logger.Error("Failed health check")
		w.WriteHeader(http.StatusServiceUnavailable)
	}
}
```

How is the health of those records evaluated? An individual function that
determines the condition of a resource is inserted into a timed loop inside a
_go routine_. Notice the function named _checkHeapSize_ is an argument to the
_beep_ function.
```go
// router/operations.go
package router
// abbreviated for clarity...

func (b *Backbone) SetupHealthChecks(cfg *config.Config) {
	pingDbTimer := time.Duration(cfg.Health.PingDbTimer)
	heapTimer := time.Duration(cfg.Health.HeapTimer)

	go beep(pingDbTimer, b.PingDB)
	go beep(heapTimer, checkHeapSize)
}
```

And _beep_ creates a _Ticker_[^t1] that will emit a signal periodically. Then
enters a loop that awaits the signal. Upon receiving the signal, a function
represented by the parameter _task_ is invoked. _checkHeapSize_ will be invoked
as the _task_.
```go
// router/operations.go
package router
// abbreviated for clarity...

func beep(seconds time.Duration, task func()) {
	ticker := time.NewTicker(seconds * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			task()
		}
	}
}
```

What is the benefit of this convoluted setup? No matter how often an external
service hammers the _/health_ endpoint, it will be less taxing because it simply
reads a _boolean_. The real work of evaluating any resource is held in a
discrete function, and there can be a few or many. They all run in the
background. They each update a particular health status on their own time. And
the configuration of time is determined by the operator of this application.


## Cache
Access the Redis client in the Backbone struct when constructing HTTP Handlers.
```go
package router
// abbreviated for clarity...

func (b *Deps) writeCache(w http.ResponseWriter, req *http.Request) {
    stuff := req.PathValue("item") // You'll probably use JSON instead.

	timer, cancel := context.WithTimeout(req.Context(), TIMEOUT_REQUEST)
	defer cancel()

	cacheErr := d.Cache.Set(timer, "KEY", stuff, 120*time.Second).Err()
    if cacheErr != nil {
        d.ServerError(w, req, cacheErr)
        return
    }

	w.Write([]byte("All good")) // Don't really do this in production.
}
```


## Build
Generate a SemVer based on the Git Commit record, then provide that value as
input to the build step. An informative record of Git Commits can aid any
operator during an incident.

```bash
~/vamos/_example $ go env -w GOEXPERIMENT=greenteagc
~/vamos/_example $ go build -v -ldflags="-s -X '/github.com/Shoowa/vamos/config.AppVersion=v.0.0.0' "
```
The linker flag _-s_ removes symbol table info and DWARF info to produce a
smaller executable. And _-X_[^b1] sets the value of a _string_ variable named
_AppVersion_ that resides in the _config_ package. This allows us to dynamically
write the version of the application after each new commit & build.

```go
package config

import (
    "os"
)

var AppVersion string
```


## Testing

### Native Functions & Discrete Packages
Three natively written functions determine equality, the absence of errors, and
truth. One less dependency in the application. Below is an example of a testing
function residing in _testhelper.go_.
```go
// testhelper/testhelper.go
func Equals(tb testing.TB, exp, act any) {
	if !reflect.DeepEqual(exp, act) {
		_, file, line, _ := runtime.Caller(1)
		fmt.Printf("\033[31m%s:%d:\n\n\texp: %#v\n\n\tgot: %#v\033[39m\n\n", filepath.Base(file), line, exp, act)
		tb.FailNow()
	}
}
```
These functions can be invoked by a test package. Use the dot at the beginning
of the import path to avoid prefacing every invocation with the name of the
_testhelper_ package.
```go
// secrets/secrets.go
package secrets_test

import (
	"testing"

	. "vamos/secrets"
	. "vamos/testhelper"
)

func Test_Configuration(t *testing.T) {
    // abbreviated for clarity...

    Equals(t, "token", openbao.Token)
}
```
Notice _secrets_test_ is a separate package from the package _secrets_. All the
tests reside in the former and the functionality resides in the latter. The
package _secrets_test_ needs to import the package _secrets_, and only public
functions & fields can be tested. This encourages _black box_ testing and clean
code.


### Integration Tests
A few steps are required to test interaction with a database.
1. Apply SQL commands to change the local development database.
2. Generate Go code in _sqlc/data/_ to interact with updated database.
3. Run Go tests marked _integration_.
4. Reverse SQL commands.

It is possible to write code into a *_test_ package that can create tables,
insert sample data, and then drop tables whenever a test is launched. Errors can
force the test to halt and leave the database with the new state without
reversing it. For this reason, it is easier to rely on a tool outside of the
application test suite to create and delete Postgres tables. I rely on
_migrate_. However, I prefer using code in the test suite to insert sample data.

Use a _make_ command to easily perform the aforementioned tasks.
```bash
~/vamos $ make test_database
```

In a downstream executable, integration tests can be invoked like this:
```bash
~/vamos/_example $ PROJECT_NAME=_example go test ./... -count=1 -tags=integration
```

#### Test Suite Setup & Teardown
The application will amend the test suite by first repositioning the root of a
test executable in order to read files that provide sample data and the
configuration file. Then the test suite will write data to the database, then
run the test functions. Lastly, the report is offered.
```go
// _example/tests/data_test.go
package data_test

import (
    // abbreviated for clarity...
    "testing"
	"github.com/Shoowa/vamos/testhelper"
)

func TestMain(m *testing.M) {
    // Direct app to read dev.json
	os.Setenv("APP_ENV", "DEV")

    // Reposition root of test executable.
	testhelper.Change_to_project_root()

	timer, _ := context.WithTimeout(context.Background(), time.Second*5)
	// Setup common resource for all integration tests in only this package.
	dbErr := testhelper.CreateTestTable(timer)
	if dbErr != nil {
		panic(dbErr)
	}
	os.Unsetenv("APP_ENV")

	code := m.Run()
	os.Exit(code)
}
```

The first function tested is the one that creates a connection pool. No other
test runs concurrently in this moment. The environment inside the test is
adjusted to induce reading configuration data for the _development_ environment.
```go
func Test_ConnectDB(t *testing.T) {
	t.Setenv("APP_ENV", "DEV")
	t.Setenv("OPENBAO_TOKEN", "token")
	cfg := config.Read()
	db, dbErr := rdbms.ConnectDB(cfg, cfg.Test.DbPosition)
	Ok(t, dbErr)
	t.Cleanup(func() { db.Close() })
}
```

In the *_example* executable, concurrent reading operations are tested in
_tests/data_test.go_. And they rely on a common connection pool created in the
same test group. The final action of the test group is to close the connection
pool.
```go
func Test_ReadingData(t *testing.T) {
	t.Setenv("APP_ENV", "DEV")
	t.Setenv("OPENBAO_TOKEN", "token")

	cfg := config.Read()
	db, _ := rdbms.ConnectDB(cfg, cfg.Test.DbPosition)
	q := first.New(db) // return sqlC generated *Queries

	timer, _ := context.WithTimeout(context.Background(), TIMEOUT_READ)

	t.Run("Read one author", func(t *testing.T) {
		readOneAuthor(t, q, timer)
	})

	t.Run("Read many authors", func(t *testing.T) {
		readManyAuthors(t, q, timer)
	})

	t.Run("Read most productive author", func(t *testing.T) {
		readMostProductiveAuthor(t, q, timer)
	})

	t.Run("Read most productive author & book", func(t *testing.T) {
		readMostProductiveAuthorAndBook(t, q, timer)
	})

	t.Cleanup(func() { db.Close() })
}
```


## Reliable Qualities

#### Postgres Connection
The Postgres connection pool retains access to the _Openbao_ secrets storage in
a  method named _BeforeConnect_. This method ensures that the connection pool
can read fresh credentials, so it enables the security practice of revoking &
rotating credentials.
```go
// data/rdbms/rdbms.go
package rdbms
// abbreviated for clarity...

func configure(cfg *config.Config, dbPosition int) (*pgxpool.Config, error) {
	db := WhichDB(cfg, dbPosition)
    // abbreviated function body for clarity...

	pgxConfig.BeforeConnect = func(ctx context.Context, cc *pgx.ConnConfig) error {
        secretsReader := new(secrets.SkeletonKey)
        secretsReader.Create(cfg)

		pw, pwErr := secretsReader.ReadPathAndKey(db.Secret, db.SecretKey)
		if pwErr != nil {
			return pwErr
		}

		cc.Password = pw
		return nil
	}
}
```


#### Graceful Shutdown
Requests need to be terminated during a rolling deployment in a manner that
preserves the data of the customer, enhances the user experience, and avoids
alarms that can mistakenly summon staff.

The webserver is launched in a separate _go routine_, then a _channel_ is opened
to receive termination signals. This blocks the main func until either signal 2
or signal 15 is received. Then the server is gracefully stopped. When that
fails, then the errors are logged and the server is forcefully stopped.
```go
// server/server.go
package server
// abbreviated for clarity...

func Start(l *slog.Logger, s *http.Server) {
	go gracefulIgnition(s)
	l.Info("HTTP Server activated")

	catchSigTerm()

	l.Info("Begin decommissioning HTTP server.")
	shutErr := gracefulShutdown(s)

	if shutErr != nil {
		l.Error("HTTP Server shutdown error", "ERR:", shutErr.Error())
		killErr := s.Close()
		if killErr != nil {
			l.Error("HTTP Server kill error", "ERR:", killErr.Error())
		}
	}

	l.Info("HTTP Server halted")
}
```

This is convenient to invoke as one line in an executable.
```go
// _example/main.go
package main
// abbreviated for clarity...

func main() {
    // skipping the setup...
    server.Start(logger, webserver)
}
```

Signal 15 allows the program to close listening connections and idle connections
while awaiting active connections. This is essential in a dynamic environment
like a _Kubernetes_ cluster. A kubelet transmits Signal 15 to a container and
pods wait 30 seconds for application cleanup.[^k1]
```go
// server/server.go
package server
// abbreviated for clarity...

func catchSigTerm() {
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)
	<-sigChan
}
```

After Signal 15 is received, _server.GracefulShutdown(webserver)_ is invoked.
It wraps _http.Server.Shutdown(shutdownCtx)_ with a 15 second timer. And the
cancellation function _stop()_ will also be invoked.
```go
// server/server.go
package server
// abbreviated for clarity...

const GRACE_PERIOD = time.Second * 15

func GracefulShutdown(s *http.Server) {
	quitCtx, quit := context.WithTimeout(context.Background(), GRACE_PERIOD)
	defer quit()

	err := s.Shutdown(quitCtx)
	if err != nil {
		return err
	}
	return nil
}
```

_stop()_ was assigned to the server during configuration. It signals to all the
child contexts derived from _base_ and used by the HTTP Handlers to terminate
any active connections.
```go
// server/server.go
package server
// abbreviated for clarity...

func NewServer(cfg *config.Config, router http.Handler) *http.Server {
	base, stop := context.WithCancel(context.Background())

	s := &http.Server{
		BaseContext:  func(lstnr net.Listener) context.Context { return base },
	}

	s.RegisterOnShutdown(stop) // Cancellation Func assigned to shutdown.
	return s
}
```

Inserting the webserver in a _go routine_ is required to avoid a hasty shutdown.
When _http.Server.Shutdown()_ is invoked, _http.Server.ListenAndServe()_ returns
immediately.[^s1] _ListenAndServe()_ was blocking in a _go routine_, and becomes
un-blocked. If _ListenAndServe()_ had been implemented in _main()_, then it
would immediately un-block and _main()_ would immediately return.


## Operate
Two environmental variables are needed by the application to read a
configuration file and access storage of sensitive credentials.
```bash
~/vamos $ APP_ENV=DEV OPENBAO_TOKEN=token ./vamos
```

### Rate Limiting
A simple Token Bucket rate limiter from the official external library can be
activated in the _config_ file. Toggle the field *global_rate_limiter_ to *true*
and define the amount of tokens refilled per second in the _average_ field, and
define the amount spent per second in the _burst_ field.
```json
{
    "httpserver": {
        "global_rate_limiter": {
            "active" : true,
            "average" : 100,
            "burst" : 200
        },
        "port": "8443",
        "timeout_read": 5,
        "timeout_write": 10,
        "timeout_idle": 60
    }
}
```
Consider the amount of goroutines monitored in _health.routines_per_core_ when
defining the amount of tolerated requests.

### Router Creation Requires An Interface
The _NewRouter_ function accepts a custom interface named _Gatherer_, so that it
can actually accept two different types of structs. The first struct,
_Backbone_,  will be directly used often in the library. The second will be used
in a downstream executable as a wrapper around the _Backbone_. Both can conform
to the _Gatherer_ interface by adopting certain methods enumerated in
_router/backbone.go_.

Though an interface isn't required to use a wrapper in a downstream executable,
it does ease testing. So I haphazardly drafted one.


### Metrics
Metrics are created by _Prometheus_ in the package _metrics_ and scraped on the
endpoint _/metrics_.

#### Configuration
Several Prometheus Collectors[^m3] and their sub-metrics can be toggled on or
off in the _config_ file. A set of runtime metrics measures garbage collection,
memory, and the scheduler[^m4], and even the CPU and Mutexes.[^m5] A Process
Collector measures the state of the CPU, MEM, file descriptors, and the start
time of the process.[^m6]
```json
"metrics": {
    "garbage_collection": true,
    "memory": true,
    "scheduler": false,
    "cpu": false,
    "lock": false,
    "process": false
}
```

The _NewDBStatsCollector_ expects a _DB_ struct from the STLDIB[^m7], so I can't
implement it with the _PGX_ connection pool struct.

#### HTTP Requests
New metrics needs to be registered to be activated.

The routing middleware in _router/middleware.go_ counts the number of HTTP
responses by HTTP Verb & Path.
```go
// router/middleware.go
package router
// abbreviated for clarity...

func logRequests(logger *slog.Logger, next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		logger.Info(
			"Inbound",
			"method", r.Method,
			"path", r.URL.Path,
			"uagent", r.Header.Get("User-Agent"),
		)

		next.ServeHTTP(w, r)
	})
}
```


### Logging Configuration
Logging is configured as _debug_ in development or as _warn_ in production.

The level is read in _logging.go_.
```go
// logging/logging.go
package logging

func configure(cfg *config.Config) *slog.HandlerOptions {
	logLevel := &slog.LevelVar{}
	if cfg.Logger.Level == "debug" {
		logLevel.Set(slog.LevelDebug)
	} else {
		logLevel.Set(slog.LevelWarn)
	}

	opts := &slog.HandlerOptions{Level: logLevel}
	return opts
}
```

The primary logger is configured to include two details that can aid anyone
debugging an incident in production. The version of the language, and the
version of the application. Every child logger inherits these details.
```go
// logging/logging.go
package logging

func CreateLogger(cfg *config.Config) *slog.Logger {
	goVersion := slog.String("lang", runtime.Version())
	appVersion := slog.String("app", config.AppVersion)
	group := slog.Group("version", goVersion, appVersion)

	opts := configure(cfg)
	handler := slog.NewJSONHandler(os.Stdout, opts)
	logger := slog.New(handler).With(group)

	slog.SetDefault(logger)
	return logger
}
```

This can be observed during startup.
```bash
~/vamos $ APP_ENV=DEV OPENBAO_TOKEN=token ./vamos
{"time":"2025-07-24T13:05:01.477738-04:00","level":"INFO","msg":"Begin logging","version":{"lang":"go1.24.0","app":"v.0.0.0"},"level":"DEBUG"}
```


### Logging Middleware
The middleware is configured in _router/middleware.go_ as a closure that passes
a logger into a _http.Handler_.
```go
// router/middleware.go
package router
// abbreviated for clarity...

func logRequests(logger *slog.Logger, next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		logger.Info(
			"Inbound",
			"method", r.Method,
			"path", r.URL.Path,
			"uagent", r.Header.Get("User-Agent"),
		)

		next.ServeHTTP(w, r)
	})
}
```
Details of every _request_ are recorded. The HTTP method, path, and _User-Agent_
header are highlighted. After those details are logged, the function continues
to the next _http.Handler_.

By satisfying this _interface_, the _http.Server_ can treat middleware as a
router.
```go
// server/server.go
package server
// abbreviated for clarity...

func NewServer(cfg *config.Config, router http.Handler) *http.Server {
	s := &http.Server{
		Addr:         ":" + cfg.HttpServer.Port,
		Handler:      router,
	}
	return s
}
```

The _Backbone_ struct conforms to the custom interface _Gatherer_, so it can be
accepted by the function _NewRouter_. _Backbone_ holds the logger that can be
used by HTTP Handlers and middleware.

```go
// router/router.go
package router
// abbreviated for clarity...

func NewRouter(dependencies Gatherer) http.Handler {
	mux := http.NewServeMux()

    // Read list of HTTP methods & http.Handlers.
	endpoints := dependencies.GetEndpoints()

    // Add each HTP path and handler to the router.
	for _, endpoint := range endpoints {
		mux.HandleFunc(endpoint.VerbAndPath, endpoint.Handler)
	}

    // Apply middleware to the router.
	responseRecordingMW := recordResponses(mux)
	loggingMW := logRequests(dependencies.GetLogger(), responseRecordingMW)
	gaugingMW := gaugeRequests(loggingMW)

	return gaugingMW
}
```

Then every incoming request is logged in a standard manner.
```bash
~/vamos $ APP_ENV=DEV OPENBAO_TOKEN=token ./vamos
# skipping other logs...
{"time":"2025-08-05T16:45:17.23609-04:00","level":"INFO","msg":"Inbound","version":{"lang":"go1.24.0","app":"v.0.0.0"},"server":{"method":"GET","path":"/health","uagent":"HTTPie/3.2.4"}}
```


### Continuous Profiling
We can obtain useful data from the production environment during a memory
problem.

A _Backbone_ field named _HeapSnapshot_ holds a pointer to a _buffer_ that
collects information generated by the
_runtime/pprof/WriteHeapProfile(io.Writer)_ function.
```go
// router/operations.go
package router
// abbreviated for clarity...

type Backbone struct {
	Logger       *slog.Logger
	Health       *Health
	DbHandle     *pgxpool.Pool
	HeapSnapshot *bytes.Buffer
}
```

The _Backbone_ struct implements the method _Write([]byte) (n int, err error)_
to comply with the _Writer interface_ expected by _WriteHeapProfile_.[^i1] And a
custom implemention resets the buffer before each write to avoid a memory leak.
```go
// router/operations.go
package router
// abbreviated for clarity...

func (b *Backbone) Write(p []byte) (n int, err error) {
	b.HeapSnapshot.Reset()
	return b.HeapSnapshot.Write(p)
}
```

After a configured threshold for memory is surpassed, heap data will be
gathered.
```go
// router/operations.go
package router
// abbreviated for clarity...

func (b *Backbone) CheckHeapSize(threshold uint64) {
	var stats runtime.MemStats
	runtime.ReadMemStats(&stats)

	if stats.HeapAlloc < threshold {
		b.Health.Heap = true
		return
	}

	b.Health.Heap = false
	b.Logger.Warn("Heap surpassed threshold!", "threshold", threshold, "allocated", stats.HeapAlloc)
	err := pprof.WriteHeapProfile(b)
	if err != nil {
		b.Logger.Error("Error writing heap profile", "ERR:", err.Error())
	}
}
```
Another method can be drafted that will read from the buffer and exfiltrate the
data for review by developers & operations staff.

#### Links
- [Guide](https://github.com/Shoowa/vamos?tab=readme-ov-file#vamos)


[^p1]: https://podman.io/docs/installation#macos
[^p2]: https://docs.fedoraproject.org/en-US/fedora-coreos/fcos-projects/
[^d1]: https://github.com/golang-migrate/migrate?tab=readme-ov-file#migrate
[^d2]: https://docs.sqlc.dev/en/stable/tutorials/getting-started-postgresql.html
[^b1]: https://pkg.go.dev/cmd/link
[^o1]: https://rednafi.com/go/dysfunctional_options_pattern/#functional-options-pattern
[^k1]: https://kubernetes.io/docs/concepts/workloads/pods/pod-lifecycle/#pod-termination-flow
[^s1]: https://pkg.go.dev/net/http#Server.Shutdown
[^t1]: https://pkg.go.dev/time#NewTicker
[^i1]: https://pkg.go.dev/runtime/pprof#WriteHeapProfile
[^m1]: https://prometheus.io/docs/tutorials/understanding_metric_types/
[^m2]: https://pkg.go.dev/github.com/prometheus/client_golang/prometheus#hdr-Advanced_Uses_of_the_Registry
[^m3]: https://pkg.go.dev/github.com/prometheus/client_golang/prometheus#Collector
[^m4]: https://pkg.go.dev/github.com/prometheus/client_golang/prometheus/collectors#pkg-variables
[^m5]: https://golang.bg/src/runtime/metrics/description.go
[^m6]: https://pkg.go.dev/github.com/prometheus/client_golang/prometheus/collectors#NewProcessCollector
[^m7]: https://pkg.go.dev/github.com/prometheus/client_golang/prometheus/collectors#NewDBStatsCollector
[^r1]: https://tip.golang.org/doc/go1.22#enhanced_routing_patterns
