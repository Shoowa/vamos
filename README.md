# ¡Vamos!
This digital service is scaffolding for a Golang web server. It is configured
with logging, metrics, health checks, & profiling. It is also integrated with
secrets storage and a relational database.

A corporate development team can deploy a prototype into a production
environment. _¡Vamos!_ hastens development and eases operation.

Let's learn how to develop, build, test, and operate this application.

## Development Environment
This is for MacOS. You will need two things: _Podman_ and _Golang_.

A natively installed instance of Postgres is fine when it is the only
dependency, but I imagine anyone using this will have an existing installation
of Postgres configured for a different development context. We can use Postgres
inside a virtual machine to avoid disruptions. And we can add other databases
and dependencies.

A virtual machine managed by _podman_[^p1] will host databases needed by the
application. The virtual machine runs Linux, specifically Fedora
CoreOS.[^p2] And _systemD_ will manage containers hosting databases.

The included _makefile_ offers a command that copies a few _.container_ files
from a directory named *_linux/* to a new directory on the MacOS host. And
copies a _.sql_ initilization script for Postgres. Then uses _podman_ to create
a virtual machine named *dev_vamos* that can read the new directory. Then uses
_systemD_ to fetch container images and run them. And setup the Postgres
instance in a container.

All you need is an installation of _podman_. Postgres will need a minute to
start.
```bash
~/vamos $ make podman_create_vm
~/vamos $ podman ps -a
```

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

### Postgres Database
A container image of Postgres 18 Beta is preferred for the native _UUIDv7_
feature. How is a container obtained and managed by _podman_ in this development
environment?

A special *.container* file is read from a user directory named
*.config/containers/systemd/* in the VM by a _podman_ tool named _quadlet_. And
_quadlet_ parses the file to produce a *systemD service* file. The resulting
*.service* file can download a container image and run it. More details can be
studied in the _makefile_ under the command _podman_create_vm.

The _quadlet .container_ file includes a few commands commonly used to run
containers in both _Docker_ and _podman_.
```bash
# _linux/dev_postgres.container
[Unit]
Description=Launch Postgres 18 with native UUIDv7

[Container]
Image=docker.io/library/postgres:18beta1-alpine3.22
ContainerName=postgres
Environment=POSTGRES_PASSWORD=password
Environment=POSTGRES_USERNAME=postgres
Environment=POSTGRES_HOST_AUTH_METHOD=trust
PublishPort=5432:5432
Volume=/data/postgres:/var/lib/postgresql/18/docker
Volume=/data/setup/setup_db1.sql:/docker-entrypoint-initdb.d/setup_db1.sql
PidsLimit=100

[Service]
Restart=always

[Install]
WantedBy=databases.target
```

The *_testdata/setup_db1.sql* file will be copied from the project on the host to
the volume of the virtual machine, then mounted to the Postgres container.
Postgres only reads this file once during its initilization. It will skip
reading it whenever the container is started again.
```sql
-- _testdata/setup_db1.sql
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


### Logs
Information about the logger.

#### Configuration
Logging is configured as _debug_ in development or as _warn_ in production.
```yaml
# internal/config/dev.yml
---
logger:
  level: debug
```

The level is read in _logging.go_.
```go
// internal/logging/logging.go
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
// internal/logging/logging.go
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
~/vamos $ go build -v -ldflags="-s -X 'vamos/internal/config.AppVersion=v.0.0.0' "
~/vamos $ APP_ENV="DEV" ./vamos
{"time":"2025-07-24T13:05:01.477738-04:00","level":"INFO","msg":"Begin logging","version":{"lang":"go1.24.0","app":"v.0.0.0"},"level":"DEBUG"}
```


[^p1]: https://podman.io/docs/installation#macos
[^p2]: https://docs.fedoraproject.org/en-US/fedora-coreos/fcos-projects/
