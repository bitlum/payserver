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

start-notification:
		@$(call print, "Notify about start...")
		curl -X POST -H 'Content-type: application/json' \
		--data '{"text":"`simnet.exchange` deploy started...", \
		"username": "$(USER)"}' \
		 $(HOOK)

end-notification:
		@$(call print, "Notify about end...")
		curl -X POST -H 'Content-type: application/json' \
		--data '{"text":"`simnet.exchange` deploy ended", \
		"username": "$(USER)"}' \
		 $(HOOK)

# # # # # # # # # #
# Docker machine  #
# # # # # # # # # #

# NOTE: Eval function if working only with "&&" because every operation in
# the makefile is working in standalone shell.
build-compose:
		@$(call print, "Activating simnet.connector.bitlum.io machine && building...")
		eval `docker-machine env simnet.connector.bitlum.io` && \
		docker-compose -f ./docker/simnet-docker-compose.yml up --build -d

# NOTE: Eval function if working only with "&&" because every operation in
# the makefile is working in standalone shell.
logs:
		@$(call print, "Activating simnet.connector.bitlum.io machine && fetching logs")
		eval `docker-machine env simnet.connector.bitlum.io` && \
		docker-compose -f ./docker/simnet-docker-compose.yml logs --tail=1000 -f

# # # # # # # # #
# Golang build  #
# # # # # # # # #

clean:
		@$(call print, "Removing build binaries...")
		rm ./docker/connector/connector

build:
		@$(call print, "Building exchange...")
		GOOS=linux GOARCH=amd64 go build -v -i -o ./docker/connector/connector

ifeq ($(HOOK),)
deploy:
		@$(call print, "You forgot specify hook!")
else
deploy: build \
		build-compose \
		clean
endif

.PHONY: deploy \
		logs \
		build \
		clean
