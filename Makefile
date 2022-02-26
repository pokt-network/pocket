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
	go build -v cmd/pocket/main.go

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
	$(eval modules_dir = "shared/modules")
	mockgen --source=${modules_dir}/persistence_module.go -destination=${modules_dir}/mocks/persistence_module_mock.go -aux_files=pocket/${modules_dir}=${modules_dir}/module.go
	mockgen --source=${modules_dir}/p2p_module.go -destination=${modules_dir}/mocks/p2p_module_mock.go -aux_files=pocket/${modules_dir}=${modules_dir}/module.go
	mockgen --source=${modules_dir}/utility_module.go -destination=${modules_dir}/mocks/utility_module_mock.go -aux_files=pocket/${modules_dir}=${modules_dir}/module.go
	mockgen --source=${modules_dir}/consensus_module.go -destination=${modules_dir}/mocks/consensus_module_mock.go -aux_files=pocket/${modules_dir}=${modules_dir}/module.go
	echo "Mocks generated in ${modules_dir}/mocks"

.PHONY: test_all
## Run all go unit tests
test_all: # mockgen
	go test ./...

.PHONY: test_pre2p
## Run all go unit tests in the pre2p module
test_pre2p: # mockgen
	go test ./pre2p/...

.PHONY: test_consensus
## Run all go unit tests in the consensus module
test_consensus: # mockgen
	go test ./consensus/...

.PHONY: test_vrf
## Run all go unit tests in the consensus module
test_vrf: # mockgen
	go test -v ./consensus/leader_election/vrf

# TODO(team): Add more protogen targets here.
.PHONY: protogen_local
## V1 Integration - Use `protoc` to generate consensus .go files from .proto files.
protogen_local:
	$(eval proto_dir = "./shared/types/proto/")

	protoc -I=${proto_dir} -I=./pre2p/types/proto --go_out=./ ./pre2p/types/proto/*.proto

	echo "View generated proto files by running: make protogen_show"

# TODO(team): Delete this once the `prototype` directory is removed.
.PHONY: protogen_local_prototype
## V1 Integration - Use `protoc` to generate consensus .go files from .proto files.
protogen_local_prototype:
	$(eval prefix = "./prototype")
	$(eval proto_dir = "${prefix}/shared/types/proto/")

	protoc -I=${proto_dir} --go_out=./ ${proto_dir}/*.proto
	protoc -I=${proto_dir} -I=${prefix}/persistence/pre_persistence/proto --go_out=./ ${prefix}/persistence/pre_persistence/proto/*.proto
	protoc -I=${proto_dir} -I=${prefix}/p2p/pre_p2p/types/proto --go_out=./ ${prefix}/p2p/pre_p2p/types/proto/*.proto
	protoc -I=${proto_dir} -I=${prefix}/utility/proto --go_out=./ ${prefix}/utility/proto/*.proto
	protoc -I=${proto_dir} -I=${prefix}/consensus/types/proto --go_out=./ ${prefix}/consensus/types/proto/*.proto

	echo "View generated proto files by running: make protogen_show"

.PHONY: protogen_show
## A simple `find` command that shows you the generated protobufs.
protogen_show:
	find . -name "*.pb.go" | grep -v "./prototype"

.PHONY: protogen_m1
## TODO(derrandz): Test, validate & update.
protogen_m1:
	docker build  -t pocket/proto-generator -f ./build/Dockerfile.m1.proto . && docker run --platform=linux/amd64 -it -v $(CWD)/shared:/usr/src/app/shared pocket/proto-generator

.PHONY: protogen
## TODO(derrandz): Test, validate & update.
protogen:
	docker build -t pocket/proto-generator -f ./build/Dockerfile.proto . && docker run -it pocket/proto-generator
