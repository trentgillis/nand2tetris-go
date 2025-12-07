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
	scanner      *bufio.Scanner
	HasMoreLines bool
	currInst     string
}

func NewParser(f *os.File) Parser {
	return Parser{
		scanner:      bufio.NewScanner(f),
		HasMoreLines: true,
		currInst:     "",
	}
}

func (p *Parser) Advance() {
	for p.scanner.Scan() {
		line := strings.TrimSpace(p.scanner.Text())
		if len(line) == 0 || strings.HasPrefix(line, "//") {
			continue
		}

		p.currInst = line
		return
	}

	p.HasMoreLines = false
	if err := p.scanner.Err(); err != nil {
		log.Fatal(err)
	}
}

func (p *Parser) CurrInstType() int {
	if strings.HasPrefix(p.currInst, "@") {
		return A_INSTRUCTION
	}

	if strings.HasPrefix(p.currInst, "(") && strings.HasSuffix(p.currInst, ")") {
		return L_INSTRUCTION
	}

	return C_INSTRUCTION
}

func (p *Parser) Symbol() string {
	after, _ := strings.CutPrefix(p.currInst, "@")
	return after
}

func (p *Parser) Dest() string {
	if !strings.Contains(p.currInst, "=") {
		return "null"
	}
	return strings.Split(p.currInst, "=")[0]
}

func (p *Parser) Comp() string {
	comp := p.currInst
	if strings.Contains(comp, "=") {
		comp = strings.Split(comp, "=")[1]
	}
	if strings.Contains(comp, ";") {
		comp = strings.Split(comp, ";")[0]
	}

	return comp
}

func (p *Parser) Jump() string {
	if !strings.Contains(p.currInst, ";") {
		return "null"
	}

	return strings.Split(p.currInst, ";")[1]
}
