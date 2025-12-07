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

run: build
	./hackassembler ./asm/add/Add.asm
	./hackassembler ./asm/max/MaxL.asm
	./hackassembler ./asm/pong/PongL.asm
	./hackassembler ./asm/rect/RectL.asm
