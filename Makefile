.PHONY: build run restart vet lint fmt test

PID=./cmd/shortener/shortener.pid
APP=cmd/shortener/shortener

all: fmt vet test build

fmt:
	@echo "Formatting the source code"
	go fmt ./...

lint:
	@echo "Linting the source code"
	# go get -u golang.org/x/lint/golint
	golint ./...

vet:
	@echo "Checking for code issues"
	go vet -vettool=$(which statictest)  ./...

build:
	@echo "Cleaning old binaries"
	@rm -f "${APP}"
	@rm -f "${APP}.pid"
	@echo "Building the binaries"
	@go build -o "${APP}" cmd/shortener/main.go

test:
	@echo "Running all tests"
	@go test -mod=mod -v ./internal/handler ./pkg/session/

kill:
	@kill `cat ${PID}` || true

run:
	@echo "Run without params"
	@${APP}  & echo $$! > ${PID}
