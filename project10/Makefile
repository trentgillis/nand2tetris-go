
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
	./tools/TextComparer.sh ./jack/ArrayTest/MainT.xml ./jack/ArrayTest/output/MainT.xml
	./tools/TextComparer.sh ./jack/Square/MainT.xml ./jack/Square/output/MainT.xml
	./tools/TextComparer.sh ./jack/Square/SquareT.xml ./jack/Square/output/SquareT.xml
	./tools/TextComparer.sh ./jack/Square/SquareGameT.xml ./jack/Square/output/SquareGameT.xml
	./tools/TextComparer.sh ./jack/ExpressionLessSquare/SquareGameT.xml ./jack/ExpressionLessSquare/output/SquareGameT.xml
	./tools/TextComparer.sh ./jack/ExpressionLessSquare/SquareGameT.xml ./jack/ExpressionLessSquare/output/SquareGameT.xml
	./tools/TextComparer.sh ./jack/ExpressionLessSquare/SquareGameT.xml ./jack/ExpressionLessSquare/output/SquareGameT.xml

run: build
	./jackc ./jack/ArrayTest
	./jackc ./jack/Square
	./jackc ./jack/ExpressionLessSquare
