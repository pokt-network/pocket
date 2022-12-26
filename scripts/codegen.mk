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

.PHONY: go_oapi-codegen
go_oapi-codegen: ## Checks if oapi-codegen is installed
	{ \
	if ! command -v oapi-codegen >/dev/null; then \
		echo "Install with 'go install github.com/deepmap/oapi-codegen/cmd/oapi-codegen@v1.11.0'"; \
	fi; \
	}


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
