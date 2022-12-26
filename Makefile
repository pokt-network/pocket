include scripts/build.mk
include scripts/test.mk
include scripts/utils.mk
include scripts/database.mk
include scripts/codegen.mk

CWD ?= CURRENT_WORKING_DIRECTIONRY_NOT_SUPPLIED

.SILENT:

.PHONY: list ## List all make targets
list:
	@${MAKE} -pRrn : -f $(MAKEFILE_LIST) 2>/dev/null | awk -v RS= -F: '/^# File/,/^# Finished Make data base/ {if ($$1 !~ "^[#.]") {print $$1}}' | egrep -v -e '^[^[:alnum:]]' -e '^$@$$' | sort

.PHONY: help ## Prints all the targets in all the Makefiles
.DEFAULT_GOAL := help
help:
	@grep -h -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | awk 'BEGIN {FS = ":.*?## "}; {printf "\033[36m%-30s\033[0m %s\n", $$1, $$2}'

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


.PHONY: go_protoc-go-inject-tag
go_protoc-go-inject-tag: ## Checks if protoc-go-inject-tag is installed
	{ \
	if ! command -v protoc-go-inject-tag >/dev/null; then \
		echo "Install with 'go install github.com/favadi/protoc-go-inject-tag@latest'"; \
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

.PHONY: swagger-ui
swagger-ui: ## Starts a local Swagger UI instance for the RPC API
	echo "Attempting to start Swagger UI at http://localhost:8080\n\n"
	docker run -p 8080:8080 -e SWAGGER_JSON=/v1/openapi.yaml -v $(shell pwd)/rpc/v1:/v1 swaggerapi/swagger-ui

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
