BINARY_NAME = brudi
COMMIT_HASH = $(shell git rev-parse --verify HEAD)
GOLANGCI_LINT_VERSION = v1.23

build:
	go build \
		-ldflags " \
			-s \
			-w \
			-X \
				'github.com/mittwald/brudi/cmd.commit=$(COMMIT_HASH)' \
		" \
		-o $(BINARY_NAME) \
		-a main.go

lint:
	docker run --rm -v $$(pwd):/app -w /app golangci/golangci-lint:$(GOLANGCI_LINT_VERSION) golangci-lint run -v

lintfix:
	docker run --rm -v $$(pwd):/app -w /app golangci/golangci-lint:$(GOLANGCI_LINT_VERSION) golangci-lint run -v --fix

goreleaser:
	curl -sL https://git.io/goreleaser | bash -s -- --snapshot --skip-publish --rm-dist