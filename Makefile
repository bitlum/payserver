help:
	@echo -e "Usage: make [command]\n\
Commands:\n\
    simnet-deploy	deploy to remote simnet droplet\n\
    simnet-init     init simnet docker containers\n\
    simnet-logs		watch remote simnet droplet's logs\n\
    simnet-ps		watch remote simnet droplet currently working containers\n\
    testnet-deploy	deploy to remote testnet droplet\n\
    testnet-logs	watch remote testnet droplet's logs\n\
    tesnet-ps		watch remote testnet droplet currently working containers"



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


HOOK := "https://hooks.slack.com/services/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"


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

# We use USERNAME makefile variable to make variable interpolation work
# in commands below.
USERNAME := $(BITLUM_NAME)

simnet-start-notification:
	@$(call print,"Notify about start...")
	curl -X POST -H 'Content-type: application/json' -w "\n" \
		--data '{"text":"`simnet.connector` deploy started...","username":"$(USERNAME)"}' \
		$(HOOK)

simnet-end-notification:
	@$(call print,"Notify about end...")
	curl -X POST -H 'Content-type: application/json' -w "\n" \
		--data '{"text":"`simnet.connector` deploy ended","username":"$(USERNAME)"}' \
		$(HOOK)

testnet-start-notification:
	@$(call print,"Notify about start...")
		curl -X POST -H 'Content-type: application/json' -w "\n" \
		--data '{"text":"`testnet.connector` deploy started...","username":"$(USERNAME)"}' \
		$(HOOK)

testnet-end-notification:
	@$(call print,"Notify about end...")
	curl -X POST -H 'Content-type: application/json' -w "\n" \
		--data '{"text":"`testnet.connector` deploy ended","username":"$(USERNAME)"}' \
		$(HOOK)



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
	@$(call print,"Activating testnet.connector.bitlum.io machine && fetching logs")
	eval `docker-machine env testnet.connector.bitlum.io` && \
	docker ps

simnet-logs:
	@$(call print,"Activating simnet.connector.bitlum.io machine && fetching logs")
	eval `docker-machine env simnet.connector.bitlum.io` && \
	docker-compose -f ./docker/simnet/docker-compose.yml logs --tail=1000 -f

testnet-build-compose:
	@$(call print,"Activating testnet.connector.bitlum.io machine && building...")

	eval `docker-machine env testnet.connector.bitlum.io-for-zigzag` && \
		cd ./docker/testnet && \
		EXTERNAL_IP=207.154.224.115 \
		PRIVATE_IP=10.135.11.56 \
		EXCHANGE_DISABLED=1 \
		docker-compose up --build -d

	eval `docker-machine env testnet.connector.bitlum.io-for-exchange` && \
		cd ./docker/testnet && \
		EXTERNAL_IP=159.89.29.186 \
		PRIVATE_IP=10.135.98.234 \
		EXCHANGE_DISABLED=0 \
		docker-compose up --build -d

testnet-logs:
	@$(call print,"Activating testnet.connector.bitlum.io machine && fetching logs")
	eval `docker-machine env testnet.connector.bitlum.io` && \
	docker-compose -f ./docker/testnet/docker-compose.yml logs --tail=1000 -f

testnet-ps:
	@$(call print,"Activating testnet.connector.bitlum.io machine && fetching logs")
	eval `docker-machine env testnet.connector.bitlum.io` && \
	docker ps



# # # # # # # # #
# Golang build  #
# # # # # # # # #

simnet-clean:
	@$(call print,"Removing simnet build connector binaries...")
	rm -rf ./docker/simnet/connector/bin

simnet-build:
	@$(call print,"Building simnet connector...")
	mkdir -p ./docker/simnet/connector/bin
	GOOS=linux GOARCH=amd64 go build -v -i -o ./docker/simnet/connector/bin/connector

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



# # # # # # # #
# Simnet init #
# # # # # # # #

simnet-init:
    eval `docker-machine env simnet.connector.bitlum.io` && \
		cd ./docker/simnet && \
		perl init.pl



.PHONY: simnet-deploy \
	testnet-deploy \
	simnet-logs \
	testnet-logs \
	simnet-ps \
	testnet-ps \
	simnet-init
