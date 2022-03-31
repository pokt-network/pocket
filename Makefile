# TODO(discuss): Determine if we want to use Makefile or mage.go and merge the two.

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
	mage build

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
	mockgen --source=${modules_dir}/persistence_module.go -destination=${modules_dir}/mocks/persistence_module_mock.go -aux_files=github.com/pokt-network/pocket/${modules_dir}=${modules_dir}/module.go
	mockgen --source=${modules_dir}/p2p_module.go -destination=${modules_dir}/mocks/p2p_module_mock.go -aux_files=github.com/pokt-network/pocket/${modules_dir}=${modules_dir}/module.go
	mockgen --source=${modules_dir}/utility_module.go -destination=${modules_dir}/mocks/utility_module_mock.go -aux_files=github.com/pokt-network/pocket/${modules_dir}=${modules_dir}/module.go
	mockgen --source=${modules_dir}/consensus_module.go -destination=${modules_dir}/mocks/consensus_module_mock.go -aux_files=github.com/pokt-network/pocket/${modules_dir}=${modules_dir}/module.go
	echo "Mocks generated in ${modules_dir}/mocks"

.PHONY: test_all
## Run all go unit tests
test_all: # generate_mocks
	go test ./...

.PHONY: test_pre2p
## Run all go unit tests in the pre2p module
test_pre2p: # generate_mocks
	go test ./pre2p/...

.PHONY: test_shared
## Run all go unit tests in the shared module
test_shared: # generate_mocks
	go test ./shared/...

.PHONY: test_pre_persistence
## Run all go per persistence unit tests
test_pre_persistence: # generate_mocks
	go test ./persistence/pre_persistence/...

# TODO(team): Tested locally with `protoc` version `libprotoc 3.19.4`. In the near future, only the Dockerfiles will be used to compile protos.

.PHONY: protogen_show
## A simple `find` command that shows you the generated protobufs.
protogen_show:
	find . -name "*.pb.go" | grep -v -e "prototype" -e "vendor"

.PHONY: protogen_clean
## Remove all the generated protobufs.
protogen_clean:
	find . -name "*.pb.go" | grep -v -e "prototype" -e "vendor" | xargs rm

.PHONY: protogen_local
## Generate go structures for all of the protobufs
protogen_local:
	$(eval proto_dir = "./shared/types/proto/")

	protoc -I=${proto_dir} -I=./shared/types/proto --go_out=./ ./shared/types/proto/*.proto
	protoc -I=${proto_dir} -I=./persistence/pre_persistence/proto --go_out=./ ./persistence/pre_persistence/proto/*.proto

	echo "View generated proto files by running: make protogen_show"

# TODO(team): Delete this once the `prototype` directory is removed.
.PHONY: protogen_local_prototype
## V1 Integration - Use `protoc` to generate consensus .go files from .proto files.
protogen_local_prototype:
	$(eval prefix = "./prototype")
	$(eval proto_dir = "${prefix}/shared/types/proto/")

	protoc -I=${proto_dir} --go_out=./ ${proto_dir}/*.proto
	protoc -I=${proto_dir} -I=${prefix}/persistence/pre_persistence/proto --go_out=${prefix} ${prefix}/persistence/pre_persistence/proto/*.proto
	protoc -I=${proto_dir} -I=${prefix}/p2p/pre_p2p/types/proto --go_out=${prefix} ${prefix}/p2p/pre_p2p/types/proto/*.proto
	protoc -I=${proto_dir} -I=${prefix}/utility/proto --go_out=${prefix} ${prefix}/utility/proto/*.proto
	protoc -I=${proto_dir} -I=${prefix}/consensus/types/proto --go_out=${prefix} ${prefix}/consensus/types/proto/*.proto

	echo "View generated proto files by running: make protogen_show"

.PHONY: protogen_docker_m1
## TODO(derrandz): Test, validate & update.
protogen_docker_m1:
	docker build  -t pocket/proto-generator -f ./build/Dockerfile.m1.proto . && docker run --platform=linux/amd64 -it -v $(CWD)/shared:/usr/src/app/shared pocket/proto-generator

.PHONY: protogen_docker
## TODO(derrandz): Test, validate & update.
protogen_docker:
	docker build -t pocket/proto-generator -f ./build/Dockerfile.proto . && docker run -it pocket/proto-generator

.PHONY: gofmt
## Format all the .go files in the project in place.
gofmt:
	gofmt -w -s .