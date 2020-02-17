include .env

test:
	go test ./... -v
fmt:
	go fmt ./...
run:
	go run -race main.go

link-lib:
	ln -s ${PWD}/node_modules/lit-html ${PWD}/static/libs
.PHONY: bin