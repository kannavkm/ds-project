PACKAGE=example.com/graphd

.PHONY: server

server:
	go build -o build/bin/server ${PACKAGE}/cmd
