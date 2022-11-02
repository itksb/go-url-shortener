.PHONY: build run restart vet lint fmt

PID=./cmd/shortener/shortener.pid
APP=cmd/shortener/shortener

all: fmt lint vet build

fmt:
	@echo "Formatting the source code"
	go fmt ./...

lint:
	@echo "Linting the source code"
	# go get -u golang.org/x/lint/golint
	golint ./...

vet:
	@echo "Checking for code issues"
	go vet ./...

build:
	@echo "Cleaning old binaries"
	@rm -f "${APP}"
	@rm -f "${APP}.pid"
	@echo "Building the binaries"
	@go build -o "${APP}" cmd/shortener/main.go

kill:
	@kill `cat ${PID}` || true

run:
	@echo "Run without params"
	@${APP}  & echo $$! > ${PID}
