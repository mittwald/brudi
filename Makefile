BINARY_NAME = brudi
COMMIT_HASH = $(shell git rev-parse --verify HEAD)
CURDIR = $(shell pwd)

.PHONY: build test

all: dep test lint build

dep:
	go mod tidy
	go mod download

build:
	go build \
		-ldflags " \
			-s \
			-w \
			-X \
				'github.com/mittwald/brudi/cmd.tag=$(COMMIT_HASH)' \
		" \
		-o $(BINARY_NAME) \
		-a main.go

test:
	go test -v ./...

lintci:
	docker run --rm \
    -v $(CURDIR):/app \
    -w /app \
	-e GOLANGCI_ADDITIONAL_YML=/app/build/ci/.golangci.yml \
	quay.io/mittwald/golangci-lint:0.0.29 \
		golangci-lint run -v --fix  ./...

lint:
	docker run --rm \
		-v $(shell go env GOPATH):/go \
		-v ${CURDIR}:/app -w /app \
		-e GOLANGCI_ADDITIONAL_YML=/app/build/ci/.golangci.yml \
		quay.io/mittwald/golangci-lint:0.0.29 \
			golangci-lint run -v --fix  ./...

goreleaser:
	curl -sL https://git.io/goreleaser | bash -s -- --debug --snapshot --skip-publish --clean

upTestMongo: downTestMongo
	trap 'cd $(CURDIR) && make downTestMongo' 0 1 2 3 6 9 15
	docker-compose --file example/docker-compose/mongo.yml up -d

downTestMongo:
	docker-compose --file example/docker-compose/mongo.yml down -v --remove-orphans

upTestMysql: downTestMysql
	trap 'cd $(CURDIR) && make downTestMysql' 0 1 2 3 6 9 15
	docker-compose --file example/docker-compose/mysql.yml up -d

downTestMysql:
	docker-compose --file example/docker-compose/mysql.yml down -v --remove-orphans

upTestPostgres: downTestPostgres
	trap 'cd $(CURDIR) && make downTestPostgres' 0 1 2 3 6 9 15
	docker-compose --file example/docker-compose/postgresql.yml up -d

downTestPostgres:
	docker-compose --file example/docker-compose/postgresql.yml down -v --remove-orphans

upTestRedis: downTestRedis
	trap 'cd $(CURDIR) && make downTestRedis' 0 1 2 3 6 9 15
	docker-compose --file example/docker-compose/redis.yml up -d

downTestRedis:
	docker-compose --file example/docker-compose/redis.yml down -v --remove-orphans