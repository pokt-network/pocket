CWD ?= CURRENT_WORKING_DIRECTIONRY_NOT_SUPPLIED

# This flag is useful when running the consensus unit tests. It causes the test to wait up to the
# maximum delay specified in the source code and errors if additional unexpected messages are received.
# For example, if the test expects to receive 5 messages within 2 seconds:
# 	When EXTRA_MSG_FAIL = false: continue if 5 messages are received in 0.5 seconds
# 	When EXTRA_MSG_FAIL = true: wait for another 1.5 seconds after 5 messages are received in 0.5
#		                        seconds, and fail if any additional messages are received.
EXTRA_MSG_FAIL ?= false

# An easy way to turn off verbose test output for some of the test targets. For example
#  `$ make test_persistence` by default enables verbose testing
#  `VERBOSE_TEST="" make test_persistence` is an easy way to run the same tests without verbose output
VERBOSE_TEST ?= -v

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

.PHONY: docker_check
# Internal helper target - check if docker is installed
docker_check:
	{ \
	if ( ! ( command -v docker >/dev/null && command -v docker-compose >/dev/null )); then \
		echo "Seems like you don't have Docker or docker-compose installed. Make sure you review docs/development/README.md before continuing"; \
		exit 1; \
	fi; \
	}

.PHONY: prompt_user
# Internal helper target - prompt the user before continuing
prompt_user:
	@echo "Are you sure? [y/N] " && read ans && [ $${ans:-N} = y ]

.PHONY: go_vet
## Run `go vet` on all files in the current project
go_vet:
	go vet ./...

.PHONY: go_staticcheck
## Run `go staticcheck` on all files in the current project
go_staticcheck:
	{ \
	if command -v staticcheck >/dev/null; then \
		staticcheck ./...; \
	else \
		echo "Install with 'go install honnef.co/go/tools/cmd/staticcheck@latest'"; \
	fi; \
	}

.PHONY: go_doc
# INCOMPLETE: Generate documentation for the current project using `godo`
go_doc:
	{ \
	if command -v godoc >/dev/null; then \
		echo "Visit http://localhost:6060/pocket"; \
		godoc -http=localhost:6060  -goroot=${PWD}/..; \
	else \
		echo "Install with 'go install golang.org/x/tools/cmd/godoc@latest'"; \
	fi; \
	}

.PHONY: go_protoc-go-inject-tag
### Checks if protoc-go-inject-tag is installed
go_protoc-go-inject-tag:
	{ \
	if ! command -v protoc-go-inject-tag >/dev/null; then \
		echo "Install with 'go install github.com/favadi/protoc-go-inject-tag@latest'"; \
	fi; \
	}

.PHONY: go_clean_deps
## Runs `go mod tidy` && `go mod vendor`
go_clean_deps:
	go mod tidy && go mod vendor

.PHONY: go_lint
## Run all linters that are triggered by the CI pipeline
go_lint:
	golangci-lint run ./...

.PHONY: gofmt
## Format all the .go files in the project in place.
gofmt:
	gofmt -w -s .

.PHONY: install_cli_deps
## Installs `protoc-gen-go` and `mockgen`
install_cli_deps:
	go install "google.golang.org/protobuf/cmd/protoc-gen-go@v1.28" && protoc-gen-go --version
	go install "github.com/golang/mock/mockgen@v1.6.0" && mockgen --version
	go install "github.com/favadi/protoc-go-inject-tag@latest"

.PHONY: develop_test
## Run all of the make commands necessary to develop on the project and verify the tests pass
develop_test: docker_check
		make go_clean_deps && \
		make mockgen && \
		make protogen_clean && make protogen_local && \
		make test_all


.PHONY: client_start
## Run a client daemon which is only used for debugging purposes
client_start: docker_check
	docker-compose -f build/deployments/docker-compose.yaml up -d client

.PHONY: client_connect
## Connect to the running client debugging daemon
client_connect: docker_check
	docker exec -it client /bin/bash -c "go run app/client/*.go"

.PHONY: build_and_watch
## Continous build Pocket's main entrypoint as files change
build_and_watch:
	/bin/sh ${PWD}/build/scripts/watch_build.sh

# TODO(olshansky): Need to think of a Pocket related name for `compose_and_watch`, maybe just `pocket_watch`?
.PHONY: compose_and_watch
## Run a localnet composed of 4 consensus validators w/ hot reload & debugging
compose_and_watch: docker_check db_start monitoring_start
	docker-compose -f build/deployments/docker-compose.yaml up --force-recreate node1.consensus node2.consensus node3.consensus node4.consensus

.PHONY: rebuild_and_compose_and_watch
## Rebuilds the container from scratch and launches compose_and_watch
rebuild_and_compose_and_watch: docker_check db_start monitoring_start
	docker-compose -f build/deployments/docker-compose.yaml up --build --force-recreate node1.consensus node2.consensus node3.consensus node4.consensus

.PHONY: db_start
## Start a detached local postgres and admin instance (this is auto-triggered by compose_and_watch)
db_start: docker_check
	docker-compose -f build/deployments/docker-compose.yaml up --no-recreate -d db pgadmin

.PHONY: db_cli
## Open a CLI to the local containerized postgres instance
db_cli:
	echo "View schema by running 'SELECT schema_name FROM information_schema.schemata;'"
	docker exec -it pocket-db bash -c "psql -U postgres"

.PHONY: db_drop
## Drop all schemas used for LocalNet development matching `node%`
db_drop: docker_check
	docker exec -it pocket-db bash -c "psql -U postgres -d postgres -a -f /tmp/scripts/drop_all_schemas.sql"

.PHONY: db_bench_init
## Initialize pgbench on local postgres - needs to be called once after container is created.
db_bench_init: docker_check
	docker exec -it pocket-db bash -c "pgbench -i -U postgres -d postgres"

.PHONY: db_bench
## Run a local benchmark against the local postgres instance - TODO(olshansky): visualize results
db_bench: docker_check
	docker exec -it pocket-db bash -c "pgbench -U postgres -d postgres"

.PHONY: db_admin
## Helper to access to postgres admin GUI interface
db_admin:
	echo "Open http://0.0.0.0:5050 and login with 'pgadmin4@pgadmin.org' and 'pgadmin4'.\n The password is 'postgres'"

.PHONY: docker_kill_all
## Kill all containers started by the docker-compose file
docker_kill_all: docker_check
	docker-compose -f build/deployments/docker-compose.yaml down

.PHONY: docker_wipe
## [WARNING] Remove all the docker containers, images and volumes.
docker_wipe: docker_check prompt_user
	docker ps -a -q | xargs -r -I {} docker stop {}
	docker ps -a -q | xargs -r -I {} docker rm {}
	docker images -q | xargs -r -I {} docker rmi {}
	docker volume ls -q | xargs -r -I {} docker volume rm {}

.PHONY: monitoring_start
## Start grafana, metrics and logging system (this is auto-triggered by compose_and_watch)
monitoring_start: docker_check
	docker-compose -f build/deployments/docker-compose.yaml up --no-recreate -d grafana loki vm

.PHONY: make docker_loki_install
## Installs the loki docker driver
docker_loki_install: docker_check
	docker plugin install grafana/loki-docker-driver:latest --alias loki --grant-all-permissions

.PHONY: mockgen
## Use `mockgen` to generate mocks used for testing purposes of all the modules.
mockgen:
	$(eval modules_dir = "shared/modules")
	rm -rf ${modules_dir}/mocks
	go generate ./${modules_dir}
	echo "Mocks generated in ${modules_dir}/mocks"

	$(eval p2p_types_dir = "p2p/types")
	$(eval p2p_type_mocks_dir = "p2p/types/mocks")
	rm -rf ${p2p_type_mocks_dir}
	go generate ./${p2p_types_dir}
	echo "P2P mocks generated in ${p2p_types_dir}/mocks"

# TODO(team): Tested locally with `protoc` version `libprotoc 3.19.4`. In the near future, only the Dockerfiles will be used to compile protos.

.PHONY: protogen_show
## A simple `find` command that shows you the generated protobufs.
protogen_show:
	find . -name "*.pb.go" | grep -v -e "prototype" -e "vendor"

.PHONY: protogen_clean
## Remove all the generated protobufs.
protogen_clean:
	find . -name "*.pb.go" | grep -v -e "prototype" -e "vendor" | xargs -r rm

.PHONY: protogen_local
## Generate go structures for all of the protobufs
protogen_local: go_protoc-go-inject-tag
	$(eval proto_dir = ".")
	protoc --go_opt=paths=source_relative  -I=./shared/debug/proto        --go_out=./shared/debug       ./shared/debug/proto/*.proto        --experimental_allow_proto3_optional
	protoc --go_opt=paths=source_relative  -I=./shared/codec/proto        --go_out=./shared/codec       ./shared/codec/proto/*.proto        --experimental_allow_proto3_optional
	protoc --go_opt=paths=source_relative  -I=./utility/indexer/proto     --go_out=./utility/indexer/   ./utility/indexer/proto/*.proto     --experimental_allow_proto3_optional
	protoc --go_opt=paths=source_relative  -I=./persistence/proto         --go_out=./persistence/types  ./persistence/proto/*.proto         --experimental_allow_proto3_optional
	protoc-go-inject-tag -input="./persistence/types/*.pb.go"
	protoc --go_opt=paths=source_relative  -I=./utility/types/proto       --go_out=./utility/types      ./utility/types/proto/*.proto       --experimental_allow_proto3_optional
	protoc --go_opt=paths=source_relative  -I=./consensus/types/proto     --go_out=./consensus/types    ./consensus/types/proto/*.proto     --experimental_allow_proto3_optional
	protoc --go_opt=paths=source_relative  -I=./p2p/raintree/types/proto  --go_out=./p2p/types          ./p2p/raintree/types/proto/*.proto  --experimental_allow_proto3_optional
	protoc --go_opt=paths=source_relative  -I=./p2p/types/proto           --go_out=./p2p/types          ./p2p/types/proto/*.proto           --experimental_allow_proto3_optional
	protoc --go_opt=paths=source_relative  -I=./telemetry/proto           --go_out=./telemetry          ./telemetry/proto/*.proto           --experimental_allow_proto3_optional
	echo "View generated proto files by running: make protogen_show"

.PHONY: protogen_docker_m1
## TECHDEBT: Test, validate & update.
protogen_docker_m1: docker_check
	docker build  -t pocket/proto-generator -f ./build/Dockerfile.m1.proto . && docker run --platform=linux/amd64 -it -v $(CWD)/shared:/usr/src/app/shared pocket/proto-generator

.PHONY: protogen_docker
## TECHDEBT: Test, validate & update.
protogen_docker: docker_check
	docker build -t pocket/proto-generator -f ./build/Dockerfile.proto . && docker run -it -v $(CWD)/:/usr/src/app/ pocket/proto-generator

.PHONY: test_all
## Run all go unit tests
test_all: # generate_mocks
	go test -p 1 -count=1 ./...

.PHONY: test_all_with_json
## Run all go unit tests, output results in json file
test_all_with_json: # generate_mocks
	go test -p 1 -json ./... > test_results.json

.PHONY: test_all_with_coverage
## Run all go unit tests, output results & coverage into files
test_all_with_coverage: # generate_mocks
	go test -p 1 -v ./... -covermode=count -coverprofile=coverage.out
	go tool cover -func=coverage.out -o=coverage.out

.PHONY: test_race
## Identify all unit tests that may result in race conditions
test_race: # generate_mocks
	go test ${VERBOSE_TEST} -race ./...

.PHONY: test_utility
## Run all go utility module unit tests
test_utility: # generate_mocks
	go test ${VERBOSE_TEST} -p=1 -count=1  ./utility/...

.PHONY: test_shared
## Run all go unit tests in the shared module
test_shared: # generate_mocks
	go test ${VERBOSE_TEST} -p 1 ./shared/...

.PHONY: test_consensus
## Run all go unit tests in the Consensus module
test_consensus: # mockgen
	go test ${VERBOSE_TEST} ./consensus/...

.PHONY: test_consensus_concurrent_tests
## Run unit tests in the consensus module that could be prone to race conditions (#192)
test_consensus_concurrent_tests:
	for i in $$(seq 1 100); do go test -timeout 2s -count=1 -run ^TestHotstuff4Nodes1BlockHappyPath$  ./consensus/consensus_tests; done;
	for i in $$(seq 1 100); do go test -timeout 2s -count=1 -run ^TestHotstuff4Nodes1BlockHappyPath$  ./consensus/consensus_tests; done;
	for i in $$(seq 1 100); do go test -timeout 2s -count=1 -race -run ^TestTinyPacemakerTimeouts$  ./consensus/consensus_tests; done;
	for i in $$(seq 1 100); do go test -timeout 2s -count=1 -race -run ^TestHotstuff4Nodes1BlockHappyPath$  ./consensus/consensus_tests; done;

.PHONY: test_hotstuff
## Run all go unit tests related to hotstuff consensus
test_hotstuff: # mockgen
	go test ${VERBOSE_TEST} ./consensus/consensus_tests -run Hotstuff -failOnExtraMessages=${EXTRA_MSG_FAIL}

.PHONY: test_pacemaker
## Run all go unit tests related to the hotstuff pacemaker
test_pacemaker: # mockgen
	go test ${VERBOSE_TEST} ./consensus/consensus_tests -run Pacemaker -failOnExtraMessages=${EXTRA_MSG_FAIL}

.PHONY: test_vrf
## Run all go unit tests in the VRF library
test_vrf:
	go test ${VERBOSE_TEST} ./consensus/leader_election/vrf

.PHONY: test_sortition
## Run all go unit tests in the Sortition library
test_sortition:
	go test ${VERBOSE_TEST} ./consensus/leader_election/sortition

.PHONY: test_persistence
## Run all go unit tests in the Persistence module
test_persistence:
	go test ${VERBOSE_TEST} -p 1 -count=1 ./persistence/...

.PHONY: test_p2p_types
## Run p2p subcomponents' tests
test_p2p_types:
	go test ${VERBOSE_TEST} -race ./p2p/types

.PHONY: test_p2p
## Run all p2p
test_p2p:
	go test ${VERBOSE_TEST} -count=1 ./p2p/...

.PHONY: test_p2p_addrbook
## Run all P2P addr book related tests
test_p2p_addrbook:
	go test -run AddrBook -v -count=1 ./p2p/...

.PHONY: benchmark_sortition
## Benchmark the Sortition library
benchmark_sortition:
	go test ${VERBOSE_TEST} ./consensus/leader_election/sortition -bench=.

.PHONY: benchmark_p2p_addrbook
## Benchmark all P2P addr book related tests
benchmark_p2p_addrbook:
	go test -bench=. -run BenchmarkAddrBook -v -count=1 ./p2p/...

### Inspired by @goldinguy_ in this post: https://goldin.io/blog/stop-using-todo ###
# TODO          - General Purpose catch-all.
# TECHDEBT      - Not a great implementation, but we need to fix it later.
# IMPROVE       - A nice to have, but not a priority. It's okay if we never get to this.
# DISCUSS       - Probably requires a lengthy offline discussion to understand next steps.
# INCOMPLETE    - A change which was out of scope of a specific PR but needed to be documented.
# INVESTIGATE   - TBD what was going on, but needed to continue moving and not get distracted.
# CLEANUP       - Like TECHDEBT, but not as bad.  It's okay if we never get to this.
# HACK          - Like TECHDEBT, but much worse. This needs to be prioritized
# REFACTOR      - Similar to TECHDEBT, but will require a substantial rewrite and change across the codebase
# CONSIDERATION - A comment that involves extra work but was thoughts / considered as part of some implementation
# CONSOLIDATE   - We likely have similar implementations/types of the same thing, and we should consolidate them.
# DEPRECATE     - Code that should be removed in the future
# DISCUSS_IN_THIS_COMMIT - SHOULD NEVER BE COMMITTED TO MASTER. It is a way for the reviewer of a PR to start / reply to a discussion.
# TODO_IN_THIS_COMMIT    - SHOULD NEVER BE COMMITTED TO MASTER. It is a way to start the review process while non-critical changes are still in progress
TODO_KEYWORDS = -e "TODO" -e "TECHDEBT" -e "IMPROVE" -e "DISCUSS" -e "INCOMPLETE" -e "INVESTIGATE" -e "CLEANUP" -e "HACK" -e "REFACTOR" -e "CONSIDERATION" -e "TODO_IN_THIS_COMMIT" -e "DISCUSS_IN_THIS_COMMIT" -e "CONSOLIDATE" -e "DEPRECATE"

# How do I use TODOs?
# 1. <KEYWORD>: <Description of follow up work>;
# 	e.g. HACK: This is a hack, we need to fix it later
# 2. If there's a specific issue, or specific person, add that in paranthesiss
#   e.g. TODO(@Olshansk): Automatically link to the Github user https://github.com/olshansk
#   e.g. INVESTIGATE(#420): Automatically link this to github issue https://github.com/pokt-network/pocket/issues/420
#   e.g. DISCUSS(@Olshansk, #420): Specific individual should tend to the action item in the specific ticket
#   e.g. CLEANUP(core): This is not tied to an issue, or a person, but should only be done by the core team.
#   e.g. CLEANUP: This is not tied to an issue, or a person, and can be done by the core team or external contributors.
# 3. Feel free to add additional keywords to the list above

.PHONY: todo_list
## List all the TODOs in the project (excludes vendor and prototype directories)
todo_list:
	grep --exclude-dir={.git,vendor,prototype} -r ${TODO_KEYWORDS}  .

.PHONY: todo_count
## Print a count of all the TODOs in the project
todo_count:
	grep --exclude-dir={.git,vendor,prototype} -r ${TODO_KEYWORDS} . | wc -l

.PHONY: todo_this_commit
## List all the TODOs needed to be done in this commit
todo_this_commit:
	grep --exclude-dir={.git,vendor,prototype,.vscode} --exclude=Makefile -r -e "TODO_IN_THIS_COMMIT" -e "DISCUSS_IN_THIS_COMMIT"

# Default values for gen_genesis_and_config
numValidators ?= 4
numServiceNodes ?= 1
numApplications ?= 1
numFishermen ?= 1

.PHONY: gen_genesis_and_config
## Generate the genesis and config files for LocalNet
gen_genesis_and_config:
	go run ./build/config/main.go --genPrefix="gen." --numValidators=${numValidators} --numServiceNodes=${numServiceNodes} --numApplications=${numApplications} --numFishermen=${numFishermen}

.PHONY: gen_genesis_and_config
## Clear the genesis and config files for LocalNet
clear_genesis_and_config:
	rm build/config/gen.*.json
