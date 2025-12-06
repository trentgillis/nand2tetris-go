package assembler

import (
	"bufio"
	"log"
	"os"
	"strings"
)

const (
	A_INSTRUCTION = iota
	C_INSTRUCTION
	L_INSTRUCTION
)

type Parser struct {
	scanner            *bufio.Scanner
	HasMoreLines       bool
	currentInstruction string
}

func NewParser(f *os.File) Parser {
	return Parser{
		scanner:            bufio.NewScanner(f),
		HasMoreLines:       true,
		currentInstruction: "",
	}
}

func (p *Parser) Advance() {
	for p.scanner.Scan() {
		line := strings.TrimSpace(p.scanner.Text())
		if len(line) == 0 || strings.HasPrefix(line, "//") {
			continue
		}

		p.currentInstruction = p.scanner.Text()
		return
	}

	p.HasMoreLines = false
	if err := p.scanner.Err(); err != nil {
		log.Fatal(err)
	}
}

func (p *Parser) CurrentInstructionType() int {
	if strings.HasPrefix(p.currentInstruction, "@") {
		return A_INSTRUCTION
	}

	if strings.HasPrefix(p.currentInstruction, "(") && strings.HasSuffix(p.currentInstruction, ")") {
		return L_INSTRUCTION
	}

	return C_INSTRUCTION
}
