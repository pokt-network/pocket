include scripts/build.mk
include scripts/test.mk
include scripts/todo.mk

CWD ?= CURRENT_WORKING_DIRECTIONRY_NOT_SUPPLIED

# This flag is useful when running the consensus unit tests. It causes the test to wait up to the
# maximum delay specified in the source code and errors if additional unexpected messages are received.
# For example, if the test expects to receive 5 messages within 2 seconds:
# 	When EXTRA_MSG_FAIL = false: continue if 5 messages are received in 0.5 seconds
# 	When EXTRA_MSG_FAIL = true: wait for another 1.5 seconds after 5 messages are received in 0.5
#		                        seconds, and fail if any additional messages are received.
EXTRA_MSG_FAIL ?= false

# IMPROVE: Add `-shuffle=on` to the `go test` command to randomize the order in which tests are run.

# An easy way to turn off verbose test output for some of the test targets. For example
#  `$ make test_persistence` by default enables verbose testing
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
## Internal helper target - check if docker is installed
docker_check:
	{ \
	if ( ! ( command -v docker >/dev/null && command -v docker-compose >/dev/null )); then \
		echo "Seems like you don't have Docker or docker-compose installed. Make sure you review docs/development/README.md before continuing"; \
		exit 1; \
	fi; \
	}

.PHONY: prompt_user
## Internal helper target - prompt the user before continuing
prompt_user:
	@echo "Are you sure? [y/N] " && read ans && [ $${ans:-N} = y ]

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
go_protoc-go-inject-tag: ## Checks if protoc-go-inject-tag is installed
	{ \
	if ! command -v protoc-go-inject-tag >/dev/null; then \
		echo "Install with 'go install github.com/favadi/protoc-go-inject-tag@latest'"; \
	fi; \
	}

.PHONY: go_oapi-codegen
go_oapi-codegen: ## Checks if oapi-codegen is installed
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

.PHONY: develop_start
develop_start: ## Run all of the make commands necessary to develop on the project
		make docker_loki_check && \
		make clean_mocks && \
		make protogen_clean && make protogen_local && \
		make go_clean_deps && \
		make mockgen && \
		make generate_rpc_openapi

.PHONY: develop_test
develop_test: docker_check ## Run all of the make commands necessary to develop on the project and verify the tests pass
		make develop_start && \
		make test_all

.PHONY: client_start
client_start: docker_check ## Run a client daemon which is only used for debugging purposes
	docker-compose -f build/deployments/docker-compose.yaml up -d client --build

.PHONY: client_connect
client_connect: docker_check ## Connect to the running client debugging daemon
	docker exec -it client /bin/bash -c "go run -tags=debug cmd/p1/*.go debug"

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
db_start: docker_check ## Start a detached local postgres and admin instance (this is auto-triggered by compose_and_watch)
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
docker_wipe: docker_check prompt_user ## [WARNING] Remove all the docker containers, images and volumes.
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
docker_loki_check: ## check if the loki docker driver is installed
	if [ `docker plugin ls | grep loki: | wc -l` -eq 0 ]; then make docker_loki_install; fi

.PHONY: clean_mocks
clean_mocks: ## Use `clean_mocks` to delete mocks before recreating them. Also useful to cleanup code that was generated from a different branch
	$(eval modules_dir = "internal/shared/modules")
	find ${modules_dir}/mocks -type f ! -name "mocks.go" -exec rm {} \;
	$(eval p2p_type_mocks_dir = "internal/p2p/types/mocks")
	find ${p2p_type_mocks_dir} -type f ! -name "mocks.go" -exec rm {} \;

.PHONY: mockgen
mockgen: clean_mocks ## Use `mockgen` to generate mocks used for testing purposes of all the modules.
	$(eval modules_dir = "internal/shared/modules")
	go generate ./${modules_dir}
	echo "Mocks generated in ${modules_dir}/mocks"

	$(eval p2p_types_dir = "internal/p2p/types")
	$(eval p2p_type_mocks_dir = "internal/p2p/types/mocks")
	find ${p2p_type_mocks_dir} -type f ! -name "mocks.go" -exec rm {} \;
	go generate ./${p2p_types_dir}
	echo "P2P mocks generated in ${p2p_types_dir}/mocks"

# TODO(team): Tested locally with `protoc` version `libprotoc 3.19.4`. In the near future, only the Dockerfiles will be used to compile protos.

.PHONY: protogen_show
protogen_show: ## A simple `find` command that shows you the generated protobufs.
	find . -name "*.pb.go" | grep -v -e "prototype" -e "vendor"

.PHONY: protogen_clean
protogen_clean: ## Remove all the generated protobufs.
	find . -name "*.pb.go" | grep -v -e "prototype" -e "vendor" | xargs -r rm

.PHONY: protogen_local
protogen_local: go_protoc-go-inject-tag ## Generate go structures for all of the protobufs
	$(eval proto_dir = ".")
	protoc --go_opt=paths=source_relative  -I=./internal/shared/messaging/proto    --go_out=./internal/shared/messaging    	./internal/shared/messaging/proto/*.proto    --experimental_allow_proto3_optional
	protoc --go_opt=paths=source_relative  -I=./internal/shared/codec/proto        --go_out=./internal/shared/codec       	./internal/shared/codec/proto/*.proto        --experimental_allow_proto3_optional
	protoc --go_opt=paths=source_relative  -I=./internal/persistence/indexer/proto --go_out=./internal/persistence/indexer/ ./internal/persistence/indexer/proto/*.proto --experimental_allow_proto3_optional
	protoc --go_opt=paths=source_relative  -I=./internal/persistence/proto         --go_out=./internal/persistence/types  	./internal/persistence/proto/*.proto         --experimental_allow_proto3_optional
	protoc --go_opt=paths=source_relative  -I=./internal/utility/types/proto       --go_out=./internal/utility/types      	./internal/utility/types/proto/*.proto       --experimental_allow_proto3_optional
	protoc --go_opt=paths=source_relative  -I=./internal/consensus/types/proto     --go_out=./internal/consensus/types    	./internal/consensus/types/proto/*.proto     --experimental_allow_proto3_optional
	protoc --go_opt=paths=source_relative  -I=./internal/p2p/raintree/types/proto  --go_out=./internal/p2p/types          	./internal/p2p/raintree/types/proto/*.proto  --experimental_allow_proto3_optional
	protoc --go_opt=paths=source_relative  -I=./internal/p2p/types/proto           --go_out=./internal/p2p/types          	./internal/p2p/types/proto/*.proto           --experimental_allow_proto3_optional
	protoc --go_opt=paths=source_relative  -I=./internal/telemetry/proto           --go_out=./internal/telemetry          	./internal/telemetry/proto/*.proto           --experimental_allow_proto3_optional
	protoc --go_opt=paths=source_relative  -I=./internal/logger/proto              --go_out=./internal/logger             	./internal/logger/proto/*.proto              --experimental_allow_proto3_optional
	protoc --go_opt=paths=source_relative  -I=./internal/rpc/types/proto 		   --go_out=./internal/rpc/types          	./internal/rpc/types/proto/*.proto           --experimental_allow_proto3_optional
	protoc-go-inject-tag -input="./internal/persistence/types/*.pb.go"
	echo "View generated proto files by running: make protogen_show"

.PHONY: protogen_docker_m1
## TECHDEBT: Test, validate & update.
protogen_docker_m1: docker_check
	docker build  -t pocket/proto-generator -f ./build/Dockerfile.m1.proto . && docker run --platform=linux/amd64 -it -v $(CWD)/shared:/usr/src/app/shared pocket/proto-generator

.PHONY: protogen_docker
## TECHDEBT: Test, validate & update.
protogen_docker: docker_check
	docker build -t pocket/proto-generator -f ./build/Dockerfile.proto . && docker run -it -v $(CWD)/:/usr/src/app/ pocket/proto-generator

.PHONY: generate_rpc_openapi
## (Re)generates the RPC server and client infra code from the openapi spec file (./rpc/v1/openapi.yaml)
generate_rpc_openapi: go_oapi-codegen
	oapi-codegen  --config ./internal/rpc/server.gen.config.yml ./internal/rpc/v1/openapi.yaml > ./internal/rpc/server.gen.go
	oapi-codegen  --config ./internal/rpc/client.gen.config.yml ./internal/rpc/v1/openapi.yaml > ./internal/rpc/client.gen.go
	echo "OpenAPI client and server generated"

.PHONY: swagger-ui
swagger-ui: ## Starts a local Swagger UI instance for the RPC API
	echo "Attempting to start Swagger UI at http://localhost:8080\n\n"
	docker run -p 8080:8080 -e SWAGGER_JSON=/v1/openapi.yaml -v $(shell pwd)/rpc/v1:/v1 swaggerapi/swagger-ui

.PHONY: generate_cli_commands_docs
## (Re)generates the CLI commands docs (this is meant to be called by CI)
generate_cli_commands_docs:
	$(eval cli_docs_dir = "cmd/p1/cli/doc/commands")
	rm ${cli_docs_dir}/*.md >/dev/null 2>&1 || true
	cd cmd/p1/cli/docgen && go run .
	echo "CLI commands docs generated in ${cli_docs_dir}"

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

.PHONY: check_cross_module_imports
check_cross_module_imports: ## Lists cross-module imports
	$(eval exclude_common=--exclude=Makefile --exclude-dir=shared --exclude-dir=cmd --exclude-dir=runtime)
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
