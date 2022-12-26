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
# DISCUSS_IN_THIS_COMMIT - SHOULD NEVER BE COMMITTED TO MASTER. It is a way for the reviewer of a PR to start / reply to a discussion.
# TODO_IN_THIS_COMMIT    - SHOULD NEVER BE COMMITTED TO MASTER. It is a way to start the review process while non-critical changes are still in progress
TODO_KEYWORDS = -e "TODO" -e "TECHDEBT" -e "IMPROVE" -e "DISCUSS" -e "INCOMPLETE" -e "INVESTIGATE" -e "CLEANUP" -e "HACK" -e "REFACTOR" -e "CONSIDERATION" -e "TODO_IN_THIS_COMMIT" -e "DISCUSS_IN_THIS_COMMIT" -e "CONSOLIDATE" -e "DEPRECATE" -e "ADDTEST"

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

.PHONY: check_cross_module_imports
check_cross_module_imports: ## Lists cross-module imports
	$(eval exclude_common=--exclude=Makefile --exclude-dir=shared --exclude-dir=cmd --exclude-dir=runtime)
	echo "persistence:\n"
	grep ${exclude_common} --exclude-dir=persistence -r "github.com/pokt-network/pocket/internal/persistence" || echo "✅ OK!"
	echo "-----------------------"
	echo "utility:\n"
	grep ${exclude_common} --exclude-dir=utility -r "github.com/pokt-network/pocket/internal/utility" || echo "✅ OK!"
	echo "-----------------------"
	echo "consensus:\n"
	grep ${exclude_common} --exclude-dir=consensus -r "github.com/pokt-network/pocket/internal/consensus" || echo "✅ OK!"
	echo "-----------------------"
	echo "telemetry:\n"
	grep ${exclude_common} --exclude-dir=telemetry -r "github.com/pokt-network/pocket/internal/telemetry" || echo "✅ OK!"
	echo "-----------------------"
	echo "p2p:\n"
	grep ${exclude_common} --exclude-dir=p2p -r "github.com/pokt-network/pocket/internal/p2p" || echo "✅ OK!"
	echo "-----------------------"
	echo "runtime:\n"
	grep ${exclude_common} --exclude-dir=runtime -r "github.com/pokt-network/pocket/internal/runtime" || echo "✅ OK!"
