default: build

SETENV=
ifeq ($(OS),Windows_NT)
	SETENV=set
endif

build:
	go build -v ./...

fix:
	go fix -v ./...

install: build
	go install -v -ldflags "-X main.version=dev-$(shell date +%Y%m%d-%H%M%S)" ./...
lint:
	golangci-lint run

generate:
	go generate ./...

fmt:
	gofmt -s -w -e .

test:
	go test -v -cover -tags=all -timeout=900s -parallel=4 ./...

eval:
	go run ./cmd/sap-iac-eval -live -provider all -judge codex -timeout 30m

.PHONY: build fix install lint generate fmt test eval
