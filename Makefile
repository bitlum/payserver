#!/usr/bin/env bash


GREEN := "\\033[0;32m"
NC := "\\033[0m"
define print
	echo $(GREEN)$1$(NC)
endef

deploy:
    @$(call print, "Running unit coverage tests.")
    GOOS=linux GOARCH=amd64 go build -v -i -o ./docker/connector/connector

eval `docker-machine env simnet.connector.bitlum.io`
docker-compose -f ./docker/simnet-docker-compose.yml -p connector up --build -d
eval `docker-machine env -u`

rm ./docker/connector/connector

.PHONY: deploy