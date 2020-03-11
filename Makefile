BINARY_NAME = brudi
COMMIT_HASH = $(shell git rev-parse --verify HEAD)
GOLANGCI_LINT_VERSION = v1.23
CURDIR = $(shell pwd)

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

lintpull:
	docker pull -q golangci/golangci-lint:$(GOLANGCI_LINT_VERSION)

lint: lintpull
	docker run --rm -v $(CURDIR):/app -w /app golangci/golangci-lint:$(GOLANGCI_LINT_VERSION) golangci-lint run -v

lintfix: lintpull
	docker run --rm -v $(CURDIR):/app -w /app golangci/golangci-lint:$(GOLANGCI_LINT_VERSION) golangci-lint run -v --fix

goreleaser:
	curl -sL https://git.io/goreleaser | bash -s -- --snapshot --skip-publish --rm-dist

upTestMongo:
	trap 'cd $(CURDIR) && make downTestMongo' 0 1 2 3 6 9 15
	docker-compose --file example/docker-compose/mongo.yml --env-file "nothing" up -d
	docker-compose --file example/docker-compose/mongo.yml --env-file "nothing" logs -f

downTestMongo:
	docker-compose --file example/docker-compose/mongo.yml --env-file "nothing" down -v

upTestMysql:
	trap 'cd $(CURDIR) && make downTestMysql' 0 1 2 3 6 9 15
	docker-compose --file example/docker-compose/mysql.yml --env-file "nothing" up -d
	docker-compose --file example/docker-compose/mysql.yml --env-file "nothing" logs -f

downTestMysql:
	docker-compose --file example/docker-compose/mysql.yml --env-file "nothing" down -v