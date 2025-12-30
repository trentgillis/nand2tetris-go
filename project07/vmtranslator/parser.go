package vmtranslator

import (
	"bufio"
	"log"
	"os"
	"strings"
)

const (
	C_ARITHMETIC = iota
	C_PUSH
	C_POP
	C_LABEL
	C_GOTO
	C_IF
	C_FUNCTION
	C_CALL
)

type Parser struct {
	HasMoreLines bool
	currLine     string
	scanner      *bufio.Scanner
}

func NewParser(file *os.File) Parser {
	scanner := bufio.NewScanner(file)
	return Parser{
		HasMoreLines: true,
		scanner:      scanner,
	}
}

func (p *Parser) Advance() {
	for p.scanner.Scan() {
		line := strings.TrimSpace(p.scanner.Text())
		if len(line) == 0 || strings.HasPrefix(line, "//") {
			continue
		}

		p.currLine = line
		return
	}

	p.HasMoreLines = false
	if err := p.scanner.Err(); err != nil {
		log.Fatal(err)
	}
}

func (p *Parser) CommandType() int {
	switch cmd := strings.Split(p.currLine, " ")[0]; cmd {
	case "add", "sub", "neg", "eq", "gt", "lt", "and", "or", "not":
		return C_ARITHMETIC
	case "push":
		return C_PUSH
	case "pop":
		return C_POP
	default:
		return 0
	}
}

func (p *Parser) Arg1() string {
	parts := strings.Split(p.currLine, " ")
	if len(parts) < 2 {
		return ""
	}
	return parts[1]
}

func (p *Parser) Arg2() string {
	parts := strings.Split(p.currLine, " ")
	if len(parts) < 3 {
		return ""
	}
	return parts[2]
}
