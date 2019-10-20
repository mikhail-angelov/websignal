include .env

test:
	go test ./... -v
fmt:
	go fmt ./...
run:
	go run main.go

.PHONY: bin