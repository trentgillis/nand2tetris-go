.DEFAULT_GOAL := run

.PHONY:fmt vet build

fmt:
	go fmt ./...

vet: fmt
	go vet ./...

build: clean vet
	go build

clean:
	go clean

test:
	go test -v ./...

run: build
	./jackvmt ./vm/StackArithmetic/SimpleAdd/SimpleAdd.vm
	./jackvmt ./vm/StackArithmetic/StackTest/StackTest.vm
