# TODO(pocket/issues/43): Delete this files after moving the necessary helpers to mage.go.

CWD ?= CURRENT_WORKING_DIRECTIONRY_NOT_SUPPLIED

# This flag is useful when running the consensus unit tests. It causes the test to wait up to the
# maximum delay specified in the source code and errors if additional unexpected messages are received.
# For example, if the test expects to receive 5 messages within 2 seconds:
# 	When EXTRA_MSG_FAIL = false: continue if 5 messages are received in 0.5 seconds
# 	When EXTRA_MSG_FAIL = true: wait for another 1.5 seconds after 5 messages are received in 0.5
#		                        seconds, and fail if any additional messages are received.
EXTRA_MSG_FAIL ?= false

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

.PHONY: go_vet
## Run `go vet` on all files in the current project
go_vet:
	go vet ./...

.PHONY: go_staticcheck
## Run `go staticcheck` on all files in the current project
go_staticcheck:
	@if builtin type -P "staticcheck"; then staticcheck ./... ; else echo "Install with 'go install honnef.co/go/tools/cmd/staticcheck@latest'"; fi

.PHONY: go_clean_dep
## Runs `go mod vendor` && `go mod tidy`
	go mod vendor && go mod tidy

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
	docker exec -it client /bin/bash -c "go run app/client/*.go"

.PHONY: compose_and_watch
## Run a localnet composed of 4 consensus validators w/ hot reload & debugging
compose_and_watch: db_start
	docker-compose -f build/deployments/docker-compose.yaml up --force-recreate node1.consensus node2.consensus node3.consensus node4.consensus

.PHONY: db_start
## Start a local postgres instance
db_start:
	docker-compose -f build/deployments/docker-compose.yaml up --no-recreate -d db

.PHONY: db_connect
## Connect to local db
db_connect:
	echo "View schema by running 'SELECT schema_name FROM information_schema.schemata;'"
	docker exec -it pocket-db bash -c "psql -U postgres"


# benchmark: pgbench
# geozone: postgis
# gui: some sort of postgres GUI


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
	go test ./... -p=1

.PHONY: test_utility_module
## Run all go utility module unit tests
test_utility_module: # generate_mocks
	go test -v ./shared/tests/utility_module/...

.PHONY: test_utility_types
## Run all go utility types module unit tests
test_utility_types: # generate_mocks
	go test -v ./utility/types/...

.PHONY: test_pre2p
## Run all go unit tests in the pre2p module
test_pre2p: # generate_mocks
	go test ./pre2p/...

.PHONY: test_shared
## Run all go unit tests in the shared module
test_shared: # generate_mocks
	go test ./shared/...

.PHONY: test_consensus
## Run all go unit tests in the Consensus module
test_consensus: # mockgen
	go test -v ./consensus/...

.PHONY: test_pre_persistence
## Run all go per persistence unit tests
test_pre_persistence: # generate_mocks
	go test ./persistence/pre_persistence/...

.PHONY: test_hotstuff
## Run all go unit tests related to hotstuff consensus
test_hotstuff: # mockgen
	go test -v ./consensus/consensus_tests -run Hotstuff -failOnExtraMessages=${EXTRA_MSG_FAIL}

.PHONY: test_pacemaker
## Run all go unit tests related to the hotstuff pacemaker
test_pacemaker: # mockgen
	go test -v ./consensus/consensus_tests -run Pacemaker -failOnExtraMessages=${EXTRA_MSG_FAIL}

.PHONY: test_vrf
## Run all go unit tests in the VRF library
test_vrf:
	go test -v ./consensus/leader_election/vrf

.PHONY: test_sortition
## Run all go unit tests in the Sortition library
test_sortition:
	go test -v ./consensus/leader_election/sortition

.PHONY: benchmark_sortition
## Benchmark the Sortition library
benchmark_sortition:
	go test -v ./consensus/leader_election/sortition -bench=.

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

	protoc --go_opt=paths=source_relative -I=${proto_dir} -I=./shared/types/proto         --go_out=./shared/types         ./shared/types/proto/*.proto
	protoc --go_opt=paths=source_relative -I=${proto_dir} -I=./utility/proto              --go_out=./utility/types        ./utility/proto/*.proto
	protoc --go_opt=paths=source_relative -I=${proto_dir} -I=./shared/types/genesis/proto --go_out=./shared/types/genesis ./shared/types/genesis/proto/*.proto
	protoc --go_opt=paths=source_relative -I=${proto_dir} -I=./consensus/types/proto      --go_out=./consensus/types      ./consensus/types/proto/*.proto

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

## Module commands

.PNONY: test_p2p_wire_codec
## Run the p2p wire codec behavior test
test_p2p_wire_codec:
	go test -run TestWireCodec -v -race ./p2p

.PHONY: test_p2p_socket
## Run the p2p net IO behaviors test
test_p2p_socket:
	go test -run TestSocket -v -race ./p2p

.PHONY: test_p2p_types
## Run p2p subcomponents' tests
test_p2p_types:
	go test -v -race ./p2p/types

.PHONY: test_p2p
## Run all p2p tests
test_p2p:
	go test -v -race ./p2p

.PHONY: todo_list
## List all the TODOs in the project (excludes vendor and prototype directories)
todo_list:
	grep --exclude-dir={.git,vendor,prototype} -r "TODO" .

.PHONY: todo_count
## Print a count of all the TODOs in the project
todo_count:
	grep --exclude-dir={.git,vendor,prototype} -r "TODO" . | wc -l
