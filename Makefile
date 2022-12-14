GOPATH ?= $(shell go env GOPATH)

.PHONY: build
build:
	mkdir -p bin
	GOBIN="$(CURDIR)/bin/" go install ./cmd/...

.PHONY: lint
lint:
	bash scripts/install-lint.sh
	${GOPATH}/bin/golangci-lint run

.PHONY: format
format:
	go fmt ./...

.PHONY: test
test:
	go test -count=1 ./...

.PHONY: clean
clean:
	rm -rf bin
