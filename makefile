####################################################
####### Podman VM & Containers #####################
####### Directories are layered as follows: ########
####### host -> VM -> container ####################
####################################################
SYSD_FILES_ON_HOST = _linux/*.{container,service,target,conf}
VOLUME_VM_CONTAINER_FILES = /data/setup/*.container
HOST_SETUP_DIR = ~/podman_vm/setup
SYSD_DIR_ON_VM = .config/containers/systemd
SYSD_PODMAN_GEN = /usr/lib/systemd/system-generators/podman-system-generator
SYSD_RELOAD = systemctl --user daemon-reload
DEV_TARGETS = secrets.target databases.target queue.target
DEV_SERVICES = dev_openbao dev_postgres openbao_add_pw nats
WEB_CFSSL = https://github.com/cloudflare/cfssl/releases/download/
CFSSL = v1.6.5/cfssl_1.6.5_linux_arm64
CFSSLJSON = v1.6.5/cfssljson_1.6.5_linux_arm64
BIN = /usr/local/bin/

podman_create_vm:
	-rm -rf ~/podman_vm && mkdir -p ~/podman_vm/{postgres,setup,ca} #Create VM volume on MacOS Host.
	-cp ${SYSD_FILES_ON_HOST} ${HOST_SETUP_DIR} #Add SystemD scripts to VM.
	-cp _example/testdata/setup_db1.sql ${HOST_SETUP_DIR} #Add sql script to Postgres container volume
	podman machine init --cpus=4 -m=2048 --disk-size 8 dev_vamos -v ~/podman_vm:/data # Define hardware of VM
	podman system connection default dev_vamos # set connection for dev_vamos VM as the default connection
	podman machine start dev_vamos # Start the VM
	podman machine ssh dev_vamos \
		"sudo curl -o cfssl --output-dir ${BIN} -SL ${WEB_CFSSL}${CFSSL}; \
		sudo curl -o cfssljson --output-dir ${BIN} -SL ${WEB_CFSSL}${CFSSLJSON}; \
		sudo chmod +x ${BIN}cfssl ${BIN}cfssljson; \
		mkdir ${SYSD_DIR_ON_VM} && cp ${VOLUME_VM_CONTAINER_FILES} ${SYSD_DIR_ON_VM}; \
		cp /data/setup/*.{service,target} .config/systemd/user; \
		${SYSD_RELOAD}; sleep 2; \
		systemctl --user enable ${DEV_TARGETS} --now"

podman_delete_vm:
	podman machine stop dev_vamos
	sleep 3
	podman machine reset -f
	rm -rf ~/podman_vm/*

podman_copy_from_host_to_vm:
	-cp ${SYSD_FILES_ON_HOST} ${HOST_SETUP_DIR}
	-cp _example/testdata/setup_db1.sql ${HOST_SETUP_DIR}
	podman machine ssh dev_vamos \
		"cp ${VOLUME_VM_CONTAINER_FILES} ${SYSD_DIR_ON_VM}; \
		cp /data/setup/*.{service,target} .config/systemd/user; \
		${SYSD_RELOAD}; sleep 2; \
		systemctl --user enable ${DEV_TARGETS} --now"

# make systemd_verify name=dev_postgres
systemd_verify:
	podman machine ssh dev_vamos "systemd-analyze --user --generators=true verify ${name}.service"

quadlet_preview:
	podman machine ssh dev_vamos "${SYSD_PODMAN_GEN} --user -dryrun" 

podman_stop_dev:
	podman machine ssh dev_vamos "systemctl --user stop ${DEV_SERVICES}"

podman_start_dev:
	podman machine ssh dev_vamos "systemctl --user start ${DEV_SERVICES}"

podman_status_dev:
	podman machine ssh dev_vamos "systemctl --user status ${DEV_SERVICES}" | less

# make podman_show_logs name=dev_postgres
podman_show_logs:
	podman machine ssh dev_vamos "journalctl --user -u ${name}.service" | less


###############################
##### GO / POSTGRES TASKS #####
###############################
SQLC_FILE = sqlc/sqlc.yaml
SQLC_MIGR_1 = sqlc/migrations/first
SQLC_MIGR_2 = sqlc/migrations/second

# Download executables into GO BIN.
download_generators:
	@go install github.com/sqlc-dev/sqlc/cmd/sqlc@latest
	@go install -tags 'postgres' github.com/golang-migrate/migrate/v4/cmd/migrate@latest

# Generate the application code for Postgres queries.
generate_db_code:
	@sqlc generate -f ${SQLC_FILE}

# make create_db1_migration name=create_authors
create_db1_migration:
	@migrate create -ext sql -dir ${SQLC_MIGR_1} -seq ${name}

create_db2_migration:
	@migrate create -ext sql -dir ${SQLC_MIGR_2} -seq ${name}

# make migrate_test_up TEST_DB="postgres://tester@localhost:5432/test_data?sslmode=disable"
migrate_test_up:
	@migrate -database ${TEST_DB} -path ${SQLC_MIGR_1} up

migrate_test_down:
	@migrate -database ${TEST_DB} -path ${SQLC_MIGR_1} down -all


############################
###### SECRETS TEST ########
############################
secrets_test:
	@go test ./secrets -count=1 --tags=integration


#########################################
###### POSTGRES INTEGRATION TEST ########
#########################################
postgres_test:
	-@go test ./data/rdbms -count=1 --tags=integration

test_database: migrate_test_up generate_db_code postgres_test migrate_test_down


####################################
##### Download Metrics Tooling #####
####################################
download_prometheus:
	@go get github.com/prometheus/client_golang/prometheus
	@go get github.com/prometheus/client_golang/prometheus/promauto
	@go get github.com/prometheus/client_golang/prometheus/promhttp
