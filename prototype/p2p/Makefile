CWD ?= UNDEFINEDCWD

compile-p2p:
	go build -race -a -o ./build/dist/node ./cmd/main.go
	chmod +x ./build/dist/node

run-p2p:
	./build/dist/node -address=$(ADDR)

start-stack:
	docker-compose -f ./build/docker/stack.yml --project-directory $(CWD) up --build --force-recreate

stop-stack:
	docker-compose -f ./build/docker/stack.yml --project-directory $(CWD) down
