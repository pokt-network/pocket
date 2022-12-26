
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
