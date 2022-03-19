PACKAGE=github.com/srk/graphd

.PHONY: server

server:
	go build -o build/server ${PACKAGE}/server
