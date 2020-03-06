BINARY_NAME = brudi
COMMIT_HASH = $(shell git rev-parse --verify HEAD)

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