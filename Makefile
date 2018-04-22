.ONESHELL:

GREEN := "\\033[0;32m"
NC := "\\033[0m"
define print
	echo $(GREEN)$1$(NC)
endef

USER := $(shell whoami)


# # # # #
# Slack #
# # # # #

start-simnet-notification:
		@$(call print, "Notify about start...")
		curl -X POST -H 'Content-type: application/json' \
		--data '{"text":"`simnet.connector` deploy started...", \
		"username": "$(USER)"}' \
		 $(HOOK)

end-simnet-notification:
		@$(call print, "Notify about end...")
		curl -X POST -H 'Content-type: application/json' \
		--data '{"text":"`simnet.connector` deploy ended", \
		"username": "$(USER)"}' \
		 $(HOOK)

start-testnet-notification:
		@$(call print, "Notify about start...")
		curl -X POST -H 'Content-type: application/json' \
		--data '{"text":"`testnet.connector` deploy started...", \
		"username": "$(USER)"}' \
		 $(HOOK)

end-testnet-notification:
		@$(call print, "Notify about end...")
		curl -X POST -H 'Content-type: application/json' \
		--data '{"text":"`testnet.connector` deploy ended", \
		"username": "$(USER)"}' \
		 $(HOOK)

# # # # # # # # # #
# Docker machine  #
# # # # # # # # # #

# NOTE: Eval function if working only with "&&" because every operation in
# the makefile is working in standalone shell.

simnet-build-compose:
		@$(call print, "Activating simnet.connector.bitlum.io machine && building...")

		cp ./docker/connector/configs/simnet/connector.conf ./docker/connector/connector.conf
		cp ./docker/connector/configs/simnet/lnd-tls.cert  ./docker/connector/lnd-tls.cert

		eval `docker-machine env simnet.connector.bitlum.io` && \
		docker-compose -f ./docker/docker-compose.yml up --build -d

		rm ./docker/connector/connector.conf
		rm ./docker/connector/lnd-tls.cert

testnet-build-compose:
		@$(call print, "Activating testnet.connector.bitlum.io machine && building...")

		cp ./docker/connector/configs/testnet/connector.conf ./docker/connector/connector.conf
		cp ./docker/connector/configs/testnet/lnd-tls.cert  ./docker/connector/lnd-tls.cert

		eval `docker-machine env testnet.connector.bitlum.io` && \
		docker-compose -f ./docker/docker-compose.yml up --build -d

		rm ./docker/connector/connector.conf
		rm ./docker/connector/lnd-tls.cert

simnet-logs:
		@$(call print, "Activating simnet.connector.bitlum.io machine && fetching logs")
		eval `docker-machine env simnet.connector.bitlum.io` && \
		docker-compose -f ./docker/docker-compose.yml logs --tail=1000 -f

testnet-logs:
		@$(call print, "Activating testnet.connector.bitlum.io machine && fetching logs")
		eval `docker-machine env testnet.connector.bitlum.io` && \
		docker-compose -f ./docker/docker-compose.yml logs --tail=1000 -f

simnet-ps:
		@$(call print, "Activating testnet.connector.bitlum.io machine && fetching logs")
		eval `docker-machine env testnet.connector.bitlum.io` && \
		docker ps

testnet-ps:
		@$(call print, "Activating testnet.connector.bitlum.io machine && fetching logs")
		eval `docker-machine env testnet.connector.bitlum.io` && \
		docker ps

# # # # # # # # #
# Golang build  #
# # # # # # # # #

clean:
		@$(call print, "Removing build connector binaries...")
		rm ./docker/connector/connector

build:
		@$(call print, "Building connector...")
		GOOS=linux GOARCH=amd64 go build -v -i -o ./docker/connector/connector

ifeq ($(HOOK),)
simnet-deploy:
		@$(call print, "You forgot specify hook!")
else
simnet-deploy: \
		start-simnet-notification \
		build \
		simnet-build-compose \
		clean \
		end-simnet-notification
endif

ifeq ($(HOOK),)
testnet-deploy:
		@$(call print, "You forgot specify hook!")
else
testnet-deploy: \
		start-testnet-notification \
		build \
		testnet-build-compose \
		clean \
		end-testnet-notification
endif

.PHONY: deploy \
		simnet-logs \
		testnet-logs \
		simnet-ps \
		testnet-ps \
		build \
		clean
