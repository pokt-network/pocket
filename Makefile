include build.mk

CWD ?= CURRENT_WORKING_DIRECTIONRY_NOT_SUPPLIED

# IMPROVE: Add `-shuffle=on` to the `go test` command to randomize the order in which tests are run.

# An easy way to turn off verbose test output for some of the test targets. For example
#  `make test_persistence` by default enables verbose testing
#  `VERBOSE_TEST="" make test_persistence` is an easy way to run the same tests without verbose output
VERBOSE_TEST ?= -v

.SILENT:

.PHONY: list ## List all make targets
list:
	@${MAKE} -pRrn : -f $(MAKEFILE_LIST) 2>/dev/null | awk -v RS= -F: '/^# File/,/^# Finished Make data base/ {if ($$1 !~ "^[#.]") {print $$1}}' | egrep -v -e '^[^[:alnum:]]' -e '^$@$$' | sort

.PHONY: help ## Prints all the targets in all the Makefiles
.DEFAULT_GOAL := help
help:
	@grep -h -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | awk 'BEGIN {FS = ":.*?## "}; {printf "\033[36m%-30s\033[0m %s\n", $$1, $$2}'

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

.PHONY: warn_destructive
warn_destructive: ## Print WARNING to the user
	@echo "This is a destructive action that will affect docker resources outside the scope of this repo!"

.PHONY: go_vet
go_vet: ## Run `go vet` on all files in the current project
	go vet ./...

.PHONY: go_staticcheck
go_staticcheck: ## Run `go staticcheck` on all files in the current project
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

.PHONY: go_oapi-codegen
### Checks if oapi-codegen is installed
go_oapi-codegen:
	{ \
	if ! command -v oapi-codegen >/dev/null; then \
		echo "Install with 'go install github.com/deepmap/oapi-codegen/cmd/oapi-codegen@v1.11.0'"; \
	fi; \
	}

.PHONY: go_clean_deps
go_clean_deps: ## Runs `go mod tidy` && `go mod vendor`
	go mod tidy && go mod vendor

.PHONY: go_lint
go_lint: ## Run all linters that are triggered by the CI pipeline
	golangci-lint run ./...

.PHONY: gofmt
gofmt: ## Format all the .go files in the project in place.
	gofmt -w -s .

.PHONY: install_cli_deps
install_cli_deps: ## Installs `protoc-gen-go`, `mockgen`, 'protoc-go-inject-tag' and other tooling
	go install "google.golang.org/protobuf/cmd/protoc-gen-go@v1.28" && protoc-gen-go --version
	go install "github.com/golang/mock/mockgen@v1.6.0" && mockgen --version
	go install "github.com/favadi/protoc-go-inject-tag@latest"
	go install "github.com/deepmap/oapi-codegen/cmd/oapi-codegen@v1.11.0"
	curl -fsSL https://raw.githubusercontent.com/tilt-dev/tilt/master/scripts/install.sh | bash
	curl -fsSL https://raw.githubusercontent.com/helm/helm/main/scripts/get-helm-3 | bash

.PHONY: develop_start
develop_start: ## Run all of the make commands necessary to develop on the project
		make docker_loki_check && \
		make clean_mocks && \
		make protogen_clean && make protogen_local && \
		make go_clean_deps && \
		make mockgen && \
		make generate_rpc_openapi && \
		make build

.PHONY: develop_test
develop_test: docker_check ## Run all of the make commands necessary to develop on the project and verify the tests pass
		make develop_start && \
		make test_all

.PHONY: client_start
client_start: docker_check ## Run a client daemon which is only used for debugging purposes
	docker-compose -f build/deployments/docker-compose.yaml up -d client

.PHONY: rebuild_client_start
rebuild_client_start: docker_check ## Rebuild and run a client daemon which is only used for debugging purposes
	docker-compose -f build/deployments/docker-compose.yaml up -d --build client

.PHONY: client_connect
client_connect: docker_check ## Connect to the running client debugging daemon
	docker exec -it client /bin/bash -c "POCKET_P2P_IS_CLIENT_ONLY=true go run -tags=debug app/client/*.go debug"

.PHONY: build_and_watch
build_and_watch: ## Continous build Pocket's main entrypoint as files change
	/bin/sh ${PWD}/build/scripts/watch_build.sh

# TODO(olshansky): Need to think of a Pocket related name for `compose_and_watch`, maybe just `pocket_watch`?
.PHONY: compose_and_watch
compose_and_watch: docker_check db_start monitoring_start ## Run a localnet composed of 4 consensus validators w/ hot reload & debugging
	docker-compose -f build/deployments/docker-compose.yaml up --force-recreate node1.consensus node2.consensus node3.consensus node4.consensus

.PHONY: rebuild_and_compose_and_watch
rebuild_and_compose_and_watch: docker_check db_start monitoring_start ## Rebuilds the container from scratch and launches compose_and_watch
	docker-compose -f build/deployments/docker-compose.yaml up --build --force-recreate node1.consensus node2.consensus node3.consensus node4.consensus

.PHONY: db_start
db_start: docker_check ## Start a detached local postgres and admin instance; compose_and_watch is responsible for instantiating the actual schemas
	docker-compose -f build/deployments/docker-compose.yaml up --no-recreate -d db pgadmin

.PHONY: db_cli
db_cli: ## Open a CLI to the local containerized postgres instance
	echo "View schema by running 'SELECT schema_name FROM information_schema.schemata;'"
	docker exec -it pocket-db bash -c "psql -U postgres"

psqlSchema ?= node1

.PHONY: db_cli_node
db_cli_node: ## Open a CLI to the local containerized postgres instance for a specific node
	echo "View all avialable tables by running \dt"
	docker exec -it pocket-db bash -c "PGOPTIONS=--search_path=${psqlSchema} psql -U postgres"

.PHONY: db_drop
db_drop: docker_check ## Drop all schemas used for LocalNet development matching `node%`
	docker exec -it pocket-db bash -c "psql -U postgres -d postgres -a -f /tmp/scripts/drop_all_schemas.sql"

.PHONY: db_bench_init
db_bench_init: docker_check ## Initialize pgbench on local postgres - needs to be called once after container is created.
	docker exec -it pocket-db bash -c "pgbench -i -U postgres -d postgres"

.PHONY: db_bench
db_bench: docker_check ## Run a local benchmark against the local postgres instance - TODO(olshansky): visualize results
	docker exec -it pocket-db bash -c "pgbench -U postgres -d postgres"

.PHONY: db_show_schemas
db_show_schemas: docker_check ## Show all the node schemas in the local SQL DB
	docker exec -it pocket-db bash -c "psql -U postgres -d postgres -a -f /tmp/scripts/show_all_schemas.sql"

.PHONY: db_admin
db_admin: ## Helper to access to postgres admin GUI interface
	echo "Open http://0.0.0.0:5050 and login with 'pgadmin4@pgadmin.org' and 'pgadmin4'.\n The password is 'postgres'"

.PHONY: docker_kill_all
docker_kill_all: docker_check ## Kill all containers started by the docker-compose file
	docker-compose -f build/deployments/docker-compose.yaml down

.PHONY: docker_wipe
docker_wipe: docker_check warn_destructive prompt_user ## [WARNING] Remove all the docker containers, images and volumes.
	docker ps -a -q | xargs -r -I {} docker stop {}
	docker ps -a -q | xargs -r -I {} docker rm {}
	docker images -q | xargs -r -I {} docker rmi {}
	docker volume ls -q | xargs -r -I {} docker volume rm {}

.PHONY: docker_wipe_nodes
docker_wipe_nodes: docker_check prompt_user db_drop ## [WARNING] Remove all the node containers
	docker ps -a -q --filter="name=node*" | xargs -r -I {} docker stop {}
	docker ps -a -q --filter="name=node*" | xargs -r -I {} docker rm {}

.PHONY: monitoring_start
monitoring_start: docker_check ## Start grafana, metrics and logging system (this is auto-triggered by compose_and_watch)
	docker-compose -f build/deployments/docker-compose.yaml up --no-recreate -d grafana loki vm

.PHONY: docker_loki_install
docker_loki_install: docker_check ## Installs the loki docker driver
	echo "Installing the loki docker driver...\n"
	docker plugin install grafana/loki-docker-driver:latest --alias loki --grant-all-permissions

.PHONY: docker_loki_check
## check if the loki docker driver is installed
docker_loki_check:
	if [ `docker plugin ls | grep loki: | wc -l` -eq 0 ]; then make docker_loki_install; fi

.PHONY: clean_mocks
clean_mocks: ## Use `clean_mocks` to delete mocks before recreating them. Also useful to cleanup code that was generated from a different branch
	$(eval modules_dir = "shared/modules")
	find ${modules_dir}/mocks -type f ! -name "mocks.go" -exec rm {} \;
	$(eval p2p_type_mocks_dir = "p2p/types/mocks")
	find ${p2p_type_mocks_dir} -type f ! -name "mocks.go" -exec rm {} \;

.PHONY: mockgen
mockgen: clean_mocks ## Use `mockgen` to generate mocks used for testing purposes of all the modules.
	$(eval modules_dir = "shared/modules")
	go generate ./${modules_dir}
	echo "Mocks generated in ${modules_dir}/mocks"
	
	$(eval DIRS = p2p persistence)
	for dir in $(DIRS); do \
		echo "Processing $$dir mocks..."; \
        find $$dir/types/mocks -type f ! -name "mocks.go" -exec rm {} \;; \
        go generate ./${dir_name}/...; \
        echo "$$dir mocks generated in $$dir/types/mocks"; \
    done
	
# TODO(team): Tested locally with `protoc` version `libprotoc 3.19.4`. In the near future, only the Dockerfiles will be used to compile protos.

.PHONY: protogen_show
protogen_show: ## A simple `find` command that shows you the generated protobufs.
	find . -name "*.pb.go" | grep -v -e "prototype" -e "vendor"

.PHONY: protogen_clean
protogen_clean: ## Remove all the generated protobufs.
	find . -name "*.pb.go" | grep -v -e "prototype" -e "vendor" | xargs -r rm

# IMPROVE: Look into using buf in the future; https://github.com/bufbuild/buf.
PROTOC = protoc --experimental_allow_proto3_optional --go_opt=paths=source_relative
PROTOC_SHARED = $(PROTOC) -I=./shared

.PHONY: protogen_local
protogen_local: go_protoc-go-inject-tag ## Generate go structures for all of the protobufs
	# Shared
	$(PROTOC) -I=./shared/core/types/proto    --go_out=./shared/core/types          ./shared/core/types/proto/*.proto
	$(PROTOC) -I=./shared/modules/types/proto --go_out=./shared/modules/types ./shared/modules/types/proto/*.proto
	$(PROTOC) -I=./shared/messaging/proto     --go_out=./shared/messaging           ./shared/messaging/proto/*.proto
	$(PROTOC) -I=./shared/codec/proto         --go_out=./shared/codec               ./shared/codec/proto/*.proto

	# Runtime
	$(PROTOC) -I=./runtime/configs/types/proto				--go_out=./runtime/configs/types	./runtime/configs/types/proto/*.proto
	$(PROTOC) -I=./runtime/configs/proto	-I=./runtime/configs/types/proto     				--go_out=./runtime/configs      ./runtime/configs/proto/*.proto
	$(PROTOC_SHARED) -I=./runtime/genesis/proto  --go_out=./runtime/genesis ./runtime/genesis/proto/*.proto
	protoc-go-inject-tag -input="./runtime/genesis/*.pb.go"

	# Persistence
	$(PROTOC_SHARED) -I=./persistence/indexer/proto 	--go_out=./persistence/indexer ./persistence/indexer/proto/*.proto

	# Utility
	$(PROTOC_SHARED) -I=./utility/types/proto --go_out=./utility/types ./utility/types/proto/*.proto

	# Consensus
	$(PROTOC_SHARED) -I=./consensus/types/proto --go_out=./consensus/types ./consensus/types/proto/*.proto

	# P2P
	$(PROTOC_SHARED) -I=./p2p/raintree/types/proto --go_out=./p2p/types ./p2p/raintree/types/proto/*.proto

	# echo "View generated proto files by running: make protogen_show"

# CONSIDERATION: Some proto files contain unused gRPC services so we may need to add the following
#                if/when we decide to include it: `grpc--go-grpc_opt=paths=source_relative --go-grpc_out=./output/path`

.PHONY: protogen_docker_m1
## TECHDEBT: Test, validate & update.
protogen_docker_m1: docker_check
	docker build  -t pocket/proto-generator -f ./build/Dockerfile.m1.proto . && docker run --platform=linux/amd64 -it -v $(CWD)/shared:/usr/src/app/shared pocket/proto-generator

.PHONY: protogen_docker
## TECHDEBT: Test, validate & update.
protogen_docker: docker_check
	docker build -t pocket/proto-generator -f ./build/Dockerfile.proto . && docker run -it -v $(CWD)/:/usr/src/app/ pocket/proto-generator

.PHONY: generate_rpc_openapi
generate_rpc_openapi: go_oapi-codegen ## (Re)generates the RPC server and client infra code from the openapi spec file (./rpc/v1/openapi.yaml)
	oapi-codegen  --config ./rpc/server.gen.config.yml ./rpc/v1/openapi.yaml > ./rpc/server.gen.go
	oapi-codegen  --config ./rpc/client.gen.config.yml ./rpc/v1/openapi.yaml > ./rpc/client.gen.go
	echo "OpenAPI client and server generated"

.PHONY: swagger-ui
swagger-ui: ## Starts a local Swagger UI instance for the RPC API
	echo "Attempting to start Swagger UI at http://localhost:8080\n\n"
	docker run -p 8080:8080 -e SWAGGER_JSON=/v1/openapi.yaml -v $(shell pwd)/rpc/v1:/v1 swaggerapi/swagger-ui

.PHONY: generate_cli_commands_docs
generate_cli_commands_docs: ## (Re)generates the CLI commands docs (this is meant to be called by CI)
	$(eval cli_docs_dir = "app/client/cli/doc/commands")
	rm ${cli_docs_dir}/*.md >/dev/null 2>&1 || true
	cd app/client/cli/docgen && go run .
	echo "CLI commands docs generated in ${cli_docs_dir}"

.PHONY: test_all
test_all: ## Run all go unit tests
	go test -p 1 -count=1 ./...

.PHONY: test_all_with_json_coverage
test_all_with_json_coverage: generate_rpc_openapi ## Run all go unit tests, output results & coverage into json & coverage files
	go test -p 1 -json ./... -covermode=count -coverprofile=coverage.out | tee test_results.json | jq

.PHONY: test_race
test_race: ## Identify all unit tests that may result in race conditions
	go test ${VERBOSE_TEST} -race ./...

.PHONY: test_app
test_app: ## Run all go app module unit tests
	go test ${VERBOSE_TEST} -p=1 -count=1  ./app/...

.PHONY: test_utility
test_utility: ## Run all go utility module unit tests
	go test ${VERBOSE_TEST} -p=1 -count=1  ./utility/...

.PHONY: test_shared
test_shared: ## Run all go unit tests in the shared module
	go test ${VERBOSE_TEST} -p 1 ./shared/...

.PHONY: test_consensus
test_consensus: ## Run all go unit tests in the consensus module
	go test ${VERBOSE_TEST} -count=1 ./consensus/...

# These tests are isolated to a single package which enables logs to be streamed in realtime. More details here: https://stackoverflow.com/a/74903989/768439
.PHONY: test_consensus_e2e
test_consensus_e2e: ## Run all go t2e unit tests in the consensus module w/ log streaming
	go test ${VERBOSE_TEST} -count=1 ./consensus/e2e_tests/...

.PHONY: test_consensus_concurrent_tests
test_consensus_concurrent_tests: ## Run unit tests in the consensus module that could be prone to race conditions (#192)
	for i in $$(seq 1 100); do go test -timeout 2s -count=1 -run ^TestPacemakerTimeoutIncreasesRound$  ./consensus/e2e_tests; done;
	for i in $$(seq 1 100); do go test -timeout 2s -count=1 -run ^TestHotstuff4Nodes1BlockHappyPath$  ./consensus/e2e_tests; done;
	for i in $$(seq 1 100); do go test -timeout 2s -count=1 -race -run ^TestPacemakerTimeoutIncreasesRound$  ./consensus/e2e_tests; done;
	for i in $$(seq 1 100); do go test -timeout 2s -count=1 -race -run ^TestHotstuff4Nodes1BlockHappyPath$  ./consensus/e2e_tests; done;

.PHONY: test_hotstuff
test_hotstuff: ## Run all go unit tests related to hotstuff consensus
	go test ${VERBOSE_TEST} ./consensus/e2e_tests -run Hotstuff

.PHONY: test_pacemaker
test_pacemaker: ## Run all go unit tests related to the hotstuff pacemaker
	go test ${VERBOSE_TEST} ./consensus/e2e_tests -run Pacemaker

.PHONY: test_vrf
test_vrf: ## Run all go unit tests in the VRF library
	go test ${VERBOSE_TEST} ./consensus/leader_election/vrf

.PHONY: test_sortition
test_sortition: ## Run all go unit tests in the Sortition library
	go test ${VERBOSE_TEST} ./consensus/leader_election/sortition

.PHONY: test_persistence
test_persistence: ## Run all go unit tests in the Persistence module
	go test ${VERBOSE_TEST} -p 1 -count=1 ./persistence/...

.PHONY: test_persistence_state_hash
test_persistence_state_hash: ## Run all go unit tests in the Persistence module related to the state hash
	go test ${VERBOSE_TEST} -run TestStateHash -count=1 ./persistence/...

.PHONY: test_p2p
test_p2p: ## Run all p2p related tests
	go test ${VERBOSE_TEST} -count=1 ./p2p/...

.PHONY: test_p2p_raintree
test_p2p_raintree: ## Run all p2p raintree related tests
	go test ${VERBOSE_TEST} -run RainTreeNetwork -count=1 ./p2p/...

.PHONY: test_p2p_raintree_addrbook
test_p2p_raintree_addrbook: ## Run all p2p raintree addr book related tests
	go test ${VERBOSE_TEST} -run RainTreeAddrBook -count=1 ./p2p/...

# TIP: For benchmarks, consider appending `-run=^#` to avoid running unit tests in the same package

.PHONY: benchmark_persistence_state_hash
benchmark_persistence_state_hash: ## Benchmark the state hash computation
	go test ${VERBOSE_TEST} -cpu 1,2 -benchtime=1s -benchmem -bench=. -run BenchmarkStateHash -count=1 ./persistence/...

.PHONY: benchmark_sortition
benchmark_sortition: ## Benchmark the Sortition library
	go test ${VERBOSE_TEST} -bench=. -run ^# ./consensus/leader_election/sortition

.PHONY: benchmark_p2p_addrbook
benchmark_p2p_addrbook: ## Benchmark all P2P addr book related tests
	go test ${VERBOSE_TEST} -bench=. -run BenchmarkAddrBook -count=1 ./p2p/...

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
# ADDTEST       - Add more tests for a specific code section
# DEPRECATE     - Code that should be removed in the future
# RESEARCH      - A non-trivial action item that requires deep research and investigation being next steps can be taken
# DOCUMENT		- A comment that involves the creation of a README or other documentation
# DISCUSS_IN_THIS_COMMIT - SHOULD NEVER BE COMMITTED TO MASTER. It is a way for the reviewer of a PR to start / reply to a discussion.
# TODO_IN_THIS_COMMIT    - SHOULD NEVER BE COMMITTED TO MASTER. It is a way to start the review process while non-critical changes are still in progress
TODO_KEYWORDS = -e "TODO" -e "TECHDEBT" -e "IMPROVE" -e "DISCUSS" -e "INCOMPLETE" -e "INVESTIGATE" -e "CLEANUP" -e "HACK" -e "REFACTOR" -e "CONSIDERATION" -e "TODO_IN_THIS_COMMIT" -e "DISCUSS_IN_THIS_COMMIT" -e "CONSOLIDATE" -e "DEPRECATE" -e "ADDTEST" -e "RESEARCH"

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
todo_list: ## List all the TODOs in the project (excludes vendor and prototype directories)
	grep --exclude-dir={.git,vendor,prototype} -r ${TODO_KEYWORDS}  .

.PHONY: todo_count
todo_count: ## Print a count of all the TODOs in the project
	grep --exclude-dir={.git,vendor,prototype} -r ${TODO_KEYWORDS} . | wc -l

.PHONY: todo_this_commit
todo_this_commit: ## List all the TODOs needed to be done in this commit
	grep --exclude-dir={.git,vendor,prototype,.vscode} --exclude=Makefile -r -e "TODO_IN_THIS_COMMIT" -e "DISCUSS_IN_THIS_COMMIT"

# Default values for gen_genesis_and_config
numValidators ?= 4
numServiceNodes ?= 1
numApplications ?= 1
numFishermen ?= 1

.PHONY: gen_genesis_and_config
gen_genesis_and_config: ## Generate the genesis and config files for LocalNet
	go run ./build/config/main.go --genPrefix="gen." --numValidators=${numValidators} --numServiceNodes=${numServiceNodes} --numApplications=${numApplications} --numFishermen=${numFishermen}

.PHONY: gen_genesis_and_config
clear_genesis_and_config: ## Clear the genesis and config files for LocalNet
	rm build/config/gen.*.json

.PHONY: localnet_up
localnet_up: ## Starts up a k8s LocalNet with all necessary dependencies (tl;dr `tilt up`)
	tilt up --file=build/localnet/Tiltfile

.PHONY: localnet_client_debug
localnet_client_debug: ## Opens a `client debug` cli to interact with blockchain (e.g. change pacemaker mode, reset to genesis, etc). Though the node binary updates automatiacally on every code change (i.e. hot reloads), if client is already open you need to re-run this command to execute freshly compiled binary.
	kubectl exec -it deploy/pocket-v1-cli-client -- client debug

.PHONY: localnet_shell
localnet_shell: ## Opens a shell in the pod that has the `client` cli available. The binary updates automatically whenever the code changes (i.e. hot reloads).
	kubectl exec -it deploy/pocket-v1-cli-client -- /bin/bash

.PHONY: localnet_logs_validators
localnet_logs_validators: ## Outputs logs from all validators
	kubectl logs -l v1-purpose=validator --all-containers=true --tail=-1

.PHONY: localnet_logs_validators_follow
localnet_logs_validators_follow: ## Outputs logs from all validators and follows them (i.e. tail)
	kubectl logs -l v1-purpose=validator --all-containers=true --max-log-requests=1000 --tail=-1 -f

.PHONY: localnet_down
localnet_down: ## Stops LocalNet and cleans up dependencies (tl;dr `tilt down` + postgres database)
	tilt down --file=build/localnet/Tiltfile
	kubectl delete pvc --ignore-not-found=true data-dependencies-postgresql-0

.PHONY: check_cross_module_imports
check_cross_module_imports: ## Lists cross-module imports
	$(eval exclude_common=--exclude=Makefile --exclude-dir=shared --exclude-dir=app --exclude-dir=runtime)
	echo "persistence:\n"
	grep ${exclude_common} --exclude-dir=persistence -r "github.com/pokt-network/pocket/persistence" || echo "✅ OK!"
	echo "-----------------------"
	echo "utility:\n"
	grep ${exclude_common} --exclude-dir=utility -r "github.com/pokt-network/pocket/utility" || echo "✅ OK!"
	echo "-----------------------"
	echo "consensus:\n"
	grep ${exclude_common} --exclude-dir=consensus -r "github.com/pokt-network/pocket/consensus" || echo "✅ OK!"
	echo "-----------------------"
	echo "telemetry:\n"
	grep ${exclude_common} --exclude-dir=telemetry -r "github.com/pokt-network/pocket/telemetry" || echo "✅ OK!"
	echo "-----------------------"
	echo "p2p:\n"
	grep ${exclude_common} --exclude-dir=p2p -r "github.com/pokt-network/pocket/p2p" || echo "✅ OK!"
	echo "-----------------------"
	echo "runtime:\n"
	grep ${exclude_common} --exclude-dir=runtime -r "github.com/pokt-network/pocket/runtime" || echo "✅ OK!"
