PACKAGE=github.com/srk/graphd

.PHONY: server

server:
	go build -o build/bin/server ${PACKAGE}/cmd/server
