help:
	@echo -e "Usage: make [command]\n\
Commands:\n\
\n\
    (Simnet related commands)\n\
    simnet-deploy    deploy to remote simnet droplet\n\
    simnet-init      init simnet docker containers\n\
    simnet-ps        watch remote simnet droplet currently working containers\n\
    simnet-logs      watch remote simnet droplet's logs\n\
    simnet-purge     stops and remove siment dockers, purge all data\n\
\n\
    (Testnet related commands)\n\
    testnet-deploy   deploy to remote testnet droplet\n\
    testnet-ps       watch remote testnet droplet currently working containers\n\
\n\
    (Mainnet related commands)\n\
    mainnet-deploy   deploy to remote mainnet droplet\n\
    mainnet-ps       watch remote mainnet droplet currently working containers\n\
    mainnet-logs     watch remote mainnet droplet's logs\n\
\n\
    (Commands bellow are for rare use and required only when configs are changed or during first deploy)\n\
    simnet-rsyslog-deploy      deploy rsyslog confg to simnet host\n\
    simnet-logrotate-deploy    deploy logrotate config to simnet host\n\
    testnet-rsyslog-deploy     deploy rsyslog confg to testnet host\n\
    testnet-logrotate-deploy   deploy logrotate config to testnet host\n\
    mainnet-rsyslog-deploy     deploy rsyslog confg to mainnet host\n\
    mainnet-logrotate-deploy   deploy logrotate config to mainnet host"



GREEN := "\\033[0;32m"
NC := "\\033[0m"
define print
	echo $(GREEN)$1$(NC)
endef

RED := "\e[31m"
define exit_error
	echo $(RED)$1$(NC)
	exit
endef


# # # # # # # #
# Bitlum name #
# # # # # # # #

# BITLUM_NAME used as `username` in slack notifications.

check-bitlum-user-set:
ifndef BITLUM_NAME
	@$(call exit_error,"BITLUM_NAME environment variable is undefined. This \
should be your bitlum name used in slack. Add \
'export BITLUM_NAME=\"your name\"' to your '~/.profile' or \
'~/.bashrc'.")
endif



# # # # #
# Slack #
# # # # #

SLACK_HOOK := $(SLACK_HOOK)

# We use USERNAME makefile variable to make variable interpolation work
# in commands below.
USERNAME := $(BITLUM_NAME)

simnet-start-notification:
	@$(call print,"Notify about start...")
	curl -X POST -H 'Content-type: application/json' -w "\n" \
		--data '{"text":"`simnet.connector` deploy started...","username":"$(USERNAME)"}' \
		$(SLACK_HOOK)

simnet-end-notification:
	@$(call print,"Notify about end...")
	curl -X POST -H 'Content-type: application/json' -w "\n" \
		--data '{"text":"`simnet.connector` deploy ended","username":"$(USERNAME)"}' \
		$(SLACK_HOOK)

testnet-start-notification:
	@$(call print,"Notify about start...")
		curl -X POST -H 'Content-type: application/json' -w "\n" \
		--data '{"text":"`testnet.connector` deploy started...","username":"$(USERNAME)"}' \
		$(SLACK_HOOK)

testnet-end-notification:
	@$(call print,"Notify about end...")
	curl -X POST -H 'Content-type: application/json' -w "\n" \
		--data '{"text":"`testnet.connector` deploy ended","username":"$(USERNAME)"}' \
		$(SLACK_HOOK)

mainnet-start-notification:
	@$(call print,"Notify about start...")
		curl -X POST -H 'Content-type: application/json' -w "\n" \
		--data '{"text":"`mainnet.connector` deploy started...","username":"$(USERNAME)"}' \
		$(SLACK_HOOK)

mainnet-end-notification:
	@$(call print,"Notify about end...")
	curl -X POST -H 'Content-type: application/json' -w "\n" \
		--data '{"text":"`mainnet.connector` deploy ended","username":"$(USERNAME)"}' \
		$(SLACK_HOOK)



# # # # # # # # # #
# Docker machine  #
# # # # # # # # # #

# NOTE: Eval function if working only with "&&" because every operation in
# the makefile is working in standalone shell.

simnet-build-compose:
	@$(call print,"Activating simnet.connector.bitlum.io machine && building...")

	eval `docker-machine env simnet.connector.bitlum.io` && \
		cd ./docker/simnet && \
		PRIVATE_IP=10.135.63.178 \
		docker-compose up --build -d

simnet-ps:
	@$(call print,"Activating simnet.connector.bitlum.io machine && getting ps")
	eval `docker-machine env simnet.connector.bitlum.io` && \
		docker ps

simnet-logs:
	@$(call print,"Connecting to simnet.connector.bitlum.io machine && fetching logs")
	docker-machine ssh simnet.connector.bitlum.io tail -f /var/log/connector/*

simnet-purge:
	@$(call print,"Purgin simnet.connector.bitlum.io machine")

	eval `docker-machine env simnet.connector.bitlum.io` && \
		docker stop `docker ps -q` || true

	docker-machine ssh simnet.connector.bitlum.io \
		rm -rf /connector/*

testnet-build-compose:
	@$(call print,"Activating testnet.connector.bitlum.io machine && building...")

	eval `docker-machine env testnet.connector.bitlum.io` && \
		cd ./docker/testnet && \
		EXTERNAL_IP=207.154.224.115 \
		PRIVATE_IP=0.0.0.0 \
		docker-compose up --build -d

testnet-ps:
	@$(call print,"Activating testnet.connector.bitlum.io machine && getting ps")

	eval `docker-machine env testnet.connector.bitlum.io` && \
		docker ps

mainnet-build-compose:
	@$(call print,"Activating mainnet.connector.bitlum.io machine && building...")

	eval `docker-machine env mainnet.connector.bitlum.io` && \
		cd ./docker/mainnet && \
		docker-compose up --build -d

mainnet-ps:
	@$(call print,"Activating mainnet.connector.bitlum.io machine && getting ps")

	eval `docker-machine env mainnet.connector.bitlum.io` && \
		docker ps

mainnet-logs:
	@$(call print,"Connecting to mainnet.connector.bitlum.io machine && fetching logs")
	docker-machine ssh mainnet.connector.bitlum.io tail -f /var/log/connector/*


# # # # # # # # #
# Golang build  #
# # # # # # # # #

simnet-clean:
	@$(call print,"Removing simnet build connector binaries...")
	rm -rf ./docker/simnet/connector/bin

simnet-build:
	@$(call print,"Building simnet connector...")
	mkdir -p ./docker/simnet/connector/bin
	GOOS=linux CC=/usr/local/gcc-4.8.1-for-linux64/bin/x86_64-pc-linux-gcc \
	CGO_ENABLED=1 GOARCH=amd64 go build -v -i -o ./docker/simnet/connector/bin/connector
	GOOS=linux CC=/usr/local/gcc-4.8.1-for-linux64/bin/x86_64-pc-linux-gcc \
	CGO_ENABLED=1 GOARCH=amd64 go build -v -i -o ./docker/simnet/connector/bin/pscli ./cmd/pscli

simnet-deploy: \
	check-bitlum-user-set \
	simnet-start-notification \
	simnet-build \
	simnet-build-compose \
	simnet-clean \
	simnet-end-notification

testnet-deploy: \
	check-bitlum-user-set \
	testnet-start-notification \
	testnet-build-compose \
	testnet-end-notification

mainnet-deploy: \
	mainnet-start-notification \
	check-bitlum-user-set \
	mainnet-build-compose \
	mainnet-end-notification

# # # # # # # #
# Simnet init #
# # # # # # # #

simnet-init:
	@$(call print,"Initing simnet...")

	eval `docker-machine env simnet.connector.bitlum.io` && \
		cd ./docker/simnet && \
		perl init.pl



# # # # # # # # # # # # # # # # #
# Rsyslog and logrotate deploy  #
# # # # # # # # # # # # # # # # #

simnet-rsyslog-deploy:
	@$(call print,"Deploying simnet rsyslog...")

	docker-machine scp \
		./docker/simnet/rsyslog.conf \
		simnet.connector.bitlum.io:/etc/rsyslog.d/10-connector.conf

	docker-machine ssh simnet.connector.bitlum.io systemctl restart syslog.service

simnet-logrotate-deploy:
	@$(call print,"Deploying simnet logrotate...")

	docker-machine scp \
		./docker/simnet/logrotate.conf \
		simnet.connector.bitlum.io:/etc/logrotate.d/connector

testnet-rsyslog-deploy:
	@$(call print,"Deploying testnet rsyslog...")

	docker-machine scp ./docker/testnet/rsyslog.conf \
		testnet.connector.bitlum.io:/etc/rsyslog.d/10-connector.conf

	docker-machine ssh testnet.connector.bitlum.io \
		systemctl restart syslog.service

testnet-logrotate-deploy:
	@$(call print,"Deploying testnet logrotate...")

	docker-machine scp \
		./docker/testnet/logrotate.conf \
		testnet.connector.bitlum.io:/etc/logrotate.d/connector

mainnet-rsyslog-deploy:
	@$(call print,"Deploying mainnet rsyslog...")

	docker-machine scp ./docker/mainnet/rsyslog.conf \
		mainnet.connector.bitlum.io:/etc/rsyslog.d/10-connector.conf

	docker-machine ssh mainnet.connector.bitlum.io \
		systemctl restart syslog.service

mainnet-logrotate-deploy:
	@$(call print,"Deploying mainnet logrotate...")

	docker-machine scp \
		./docker/mainnet/logrotate.conf \
		mainnet.connector.bitlum.io:/etc/logrotate.d/connector



.PHONY: simnet-deploy \
	simnet-init \
	simnet-ps \
	simnet-logs \
	simnet-purge \
	simnet-rsyslog-deploy \
	simnet-logrotate-deploy \
	testnet-deploy \
	testnet-ps \
	testnet-rsyslog-deploy \
	testnet-logrotate-deploy \
	mainnet-deploy \
	mainnet-ps \
	mainnet-logs \
	mainnet-rsyslog-deploy \
	mainnet-logrotate-deploy