PACKAGE=example.com/graphd
BIN=build/bin/

all: zero alpha

zero:
	go build -o ${BIN}/zero ${PACKAGE}/cmd/zero

alpha:
	go build -o ${BIN}/alpha ${PACKAGE}/cmd/alpha

