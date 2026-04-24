APP = cmd/app/main.go

all: fmt
	go run $(APP)

fmt:
	go fmt ./...

.PHONY: all fmt