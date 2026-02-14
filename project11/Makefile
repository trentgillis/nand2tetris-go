
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
	./tools/TextComparer.sh ./jack/Seven/Main.vm ./jack/ArrayTest/output/Main.vm

run: build
	./jackc ./jack/Average
	./jackc ./jack/ComplexArrays
	./jackc ./jack/ConvertToBin
	./jackc ./jack/Pong
	./jackc ./jack/Seven
	./jackc ./jack/Square
