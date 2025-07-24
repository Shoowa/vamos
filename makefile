SYSD_FILES_ON_HOST = _linux/*.{container,service,target}
VOLUME_VM_CONTAINER_FILES = /data/setup/*.container
SYSD_DIR_ON_VM = .config/containers/systemd
SYSD_PODMAN_GEN = /usr/lib/systemd/system-generators/podman-system-generator
SYSD_RELOAD = systemctl --user daemon-reload
DEV_TARGETS = secrets.target databases.target
DEV_SERVICES = dev_openbao dev_postgres openbao_add_pw




####################################################
####### Podman VM & Containers #####################
####### Directories are layered as follows: ########
####### host -> VM -> container ####################
####################################################
podman_create_vm:
	-rm -rf ~/podman_vm && mkdir -p ~/podman_vm/{couchdb1,postgres,setup} #Create VM volume on MacOS Host.
	-cp ${SYSD_FILES_ON_HOST} ~/podman_vm/setup #Add SystemD scripts to VM.
	-cp _testdata/setup_db1.sql ~/podman_vm/setup #Add sql script to Postgres container volume
	podman machine init --cpus=4 -m=2048 --disk-size 8 dev_vamos -v ~/podman_vm:/data # Define hardware of VM
	podman system connection default dev_vamos # set connection for dev_vamos VM as the default connection
	podman machine start dev_vamos # Start the VM
	podman machine ssh dev_vamos \
		"mkdir ${SYSD_DIR_ON_VM} && cp ${VOLUME_VM_CONTAINER_FILES} ${SYSD_DIR_ON_VM}; \
		cp /data/setup/*.{service,target} .config/systemd/user; \
		${SYSD_RELOAD}; sleep 2; \
		systemctl --user enable ${DEV_TARGETS} --now"

podman_delete_vm:
	podman machine stop dev_vamos
	sleep 3
	podman machine reset -f
	rm -rf ~/podman_vm/*

podman_copy_from_host_to_vm:
	-cp ${SYSD_FILES_ON_HOST} ~/podman_vm/setup
	-cp _testdata/setup_db1.sql ~/podman_vm/setup
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
