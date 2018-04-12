GREEN := "\\033[0;32m"
NC := "\\033[0m"
define print
	echo $(GREEN)$1$(NC)
endef

USER := $(shell whoami)

start-notification:
		@$(call print, "Notify about start...")
		curl -X POST -H 'Content-type: application/json' \
		--data '{"text":"`simnet.connector` deploy started...", \
		"username": "$(USER)"}' $(HOOK)

end-notification:
		@$(call print, "Notify about end...")
		curl -X POST -H 'Content-type: application/json' \
		--data '{"text":"`simnet.connector` deploy ended", \
		"username": "$(USER)"}' $(HOOK)

build:
		@$(call print, "Building connector...")
		GOOS=linux GOARCH=amd64 go build -v -i -o ./docker/connector/connector

		@$(call print, "Activating simnet.connector.bitlum.io machine...")
		eval `docker-machine env simnet.connector.bitlum.io`

		@$(call print, "Running docker-compose...")
		docker-compose -f ./docker/simnet-docker-compose.yml -p connector up --build -d

		@$(call print, "Deactivating simnet.connector.bitlum.io machine...")
		eval `docker-machine env -u`

		@$(call print, "Removing build binaries...")
		rm ./docker/connector/connector

ifeq ($(HOOK), "")
deploy: @$(call print, "You forgot specify hook!")
else
deploy: start-notification build end-notification
endif

.PHONY: deploy

