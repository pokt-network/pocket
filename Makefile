CWD ?= CURRENT_WORKING_DIRECTIONRY_NOT_SUPPLIED

.SILENT:

help:
	printf "Available targets\n\n"
	awk '/^[a-zA-Z\-\_0-9]+:/ { \
		helpMessage = match(lastLine, /^## (.*)/); \
		if (helpMessage) { \
			helpCommand = substr($$1, 0, index($$1, ":")-1); \
			helpMessage = substr(lastLine, RSTART + 3, RLENGTH); \
			printf "%-30s %s\n", helpCommand, helpMessage; \
		} \
	} \
	{ lastLine = $$0 }' $(MAKEFILE_LIST)


prompt_user:
	@echo "Are you sure? [y/N] " && read ans && [ $${ans:-N} = y ]

.PHONY: build
## Build Pocket's main entrypoint
build:
	go build -v cmd/v1/main.go

.PHONY: build_and_watch
## Continous build Pocket's main entrypoint as files change
build_and_watch:
	/bin/sh ${PWD}/scripts/watch_build.sh

.PHONY: client_start
## Run a client daemon which is only used for debugging purposes
client_start:
	docker-compose -f build/deployments/docker-compose.yaml up -d client

.PHONY: client_connect
## Connect to the running client debugging daemon
client_connect:
	docker exec -it client /bin/bash -c "go run cmd/client/*.go"

.PHONY: compose_and_watch
## Run a localnet composed of 4 consensus validators w/ hot reload & debugging
compose_and_watch:
	docker-compose -f build/deployments/docker-compose.yaml up --force-recreate node1.consensus node2.consensus node3.consensus node4.consensus

.PHONY: compose_and_watch
## Kill all containers started by the docker-compose file
docker_kill_all:
	docker-compose -f build/deployments/docker-compose.yaml down

.PHONY: docker_wipe
## [WARNING] Remove all the docker containers, images and volumes.
docker_wipe: prompt_user
	docker ps -a -q | xargs -r -I {} docker stop {}
	docker ps -a -q | xargs -r -I {} docker rm {}
	docker images -q | xargs -r -I {} docker rmi {}
	docker volume ls -q | xargs -r -I {} docker volume rm {}

.PHONY: mockgen
## Use `mockgen` to generate mocks used for testing purposes of all the modules.
mockgen:
	mockgen --source=pkg/shared/modules/pocket_module.go -destination=pkg/shared/modules/mocks/pocket_module_mock.go

	mockgen --source=pkg/shared/modules/persistence_module.go -destination=pkg/shared/modules/mocks/persistence_module_mock.go -aux_files=github.com/pocket-network/pkg/shared/modules=./pkg/shared/modules/pocket_module.go
	mockgen --source=pkg/shared/modules/utility_module.go -destination=pkg/shared/modules/mocks/utility_module_mock.go -aux_files=github.com/pocket-network/pkg/shared/modules=./pkg/shared/modules/pocket_module.go
	mockgen --source=pkg/shared/modules/p2p_module.go -destination=pkg/shared/modules/mocks/p2p_module_mock.go -aux_files=github.com/pocket-network/pkg/shared/modules=./pkg/shared/modules/pocket_module.go
	mockgen --source=pkg/p2p/p2p_types/network.go -destination=pkg/p2p/p2p_types/mocks/network_mock.go
	mockgen --source=pkg/shared/modules/consensus_module.go -destination=pkg/shared/modules/mocks/consensus_module_mock.go -aux_files=github.com/pocket-network/pkg/shared/modules=./pkg/shared/modules/pocket_module.go

.PHONY: test_all
## Run all go unit tests
test_all: # generate_mocks
	go test ./...

.PHONY: protogen_local
## V1 Integration - Use `protoc` to generate consensus .go files from .proto files.
protogen_local:
	protoc -I=./shared/types/proto/ -I=./consensus/types/proto --go_out=./ ./consensus/types/proto/*.proto
	protoc -I=./shared/types/proto/ --go_out=./ ./shared/types/proto/*.proto
	protoc -I=./shared/types/proto/ -I=./persistence/pre_persistence/proto --go_out=./ ./persistence/pre_persistence/proto/*.proto
	protoc -I=./shared/types/proto/ -I=./p2p/pre_p2p/types/proto --go_out=./ ./p2p/pre_p2p/types/proto/*.proto
	protoc -I=./shared/types/proto/ -I=./utility/proto --go_out=./ ./utility/proto/*.proto

.PHONY: protogen_m1
## TODO(derrandz): Test, validate & update.
protogen_m1:
	docker build  -t pocket/proto-generator -f ./build/Dockerfile.m1.proto . && docker run --platform=linux/amd64 -it -v $(CWD)/shared:/usr/src/app/shared pocket/proto-generator

.PHONY: protogen
## TODO(derrandz): Test, validate & update.
protogen:
	docker build -t pocket/proto-generator -f ./build/Dockerfile.proto . && docker run -it pocket/proto-generator
