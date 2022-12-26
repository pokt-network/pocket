
.PHONY: test_all
test_all: ## Run all go unit tests
	go test -p 1 -count=1 ./...

.PHONY: test_all_with_json_coverage
test_all_with_json_coverage: generate_rpc_openapi ## Run all go unit tests, output results & coverage into json & coverage files
	go test -p 1 -json ./... -covermode=count -coverprofile=coverage.out | tee test_results.json | jq

.PHONY: test_race
test_race: ## Identify all unit tests that may result in race conditions
	go test ${VERBOSE_TEST} -race ./...

.PHONY: test_utility
test_utility: ## Run all go utility module unit tests
	go test ${VERBOSE_TEST} -p=1 -count=1  ./utility/...

.PHONY: test_shared
test_shared: ## Run all go unit tests in the shared module
	go test ${VERBOSE_TEST} -p 1 ./shared/...

.PHONY: test_consensus
test_consensus: ## Run all go unit tests in the consensus module
	go test ${VERBOSE_TEST} -count=1 ./consensus/...

.PHONY: test_consensus_concurrent_tests
test_consensus_concurrent_tests: ## Run unit tests in the consensus module that could be prone to race conditions (#192)
	for i in $$(seq 1 100); do go test -timeout 2s -count=1 -run ^TestTinyPacemakerTimeouts$  ./consensus/consensus_tests; done;
	for i in $$(seq 1 100); do go test -timeout 2s -count=1 -run ^TestHotstuff4Nodes1BlockHappyPath$  ./consensus/consensus_tests; done;
	for i in $$(seq 1 100); do go test -timeout 2s -count=1 -race -run ^TestTinyPacemakerTimeouts$  ./consensus/consensus_tests; done;
	for i in $$(seq 1 100); do go test -timeout 2s -count=1 -race -run ^TestHotstuff4Nodes1BlockHappyPath$  ./consensus/consensus_tests; done;

.PHONY: test_hotstuff
test_hotstuff: ## Run all go unit tests related to hotstuff consensus
	go test ${VERBOSE_TEST} ./consensus/consensus_tests -run Hotstuff -failOnExtraMessages=${EXTRA_MSG_FAIL}

.PHONY: test_pacemaker
test_pacemaker: ## Run all go unit tests related to the hotstuff pacemaker
	go test ${VERBOSE_TEST} ./consensus/consensus_tests -run Pacemaker -failOnExtraMessages=${EXTRA_MSG_FAIL}

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
