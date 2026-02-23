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
	HasMoreLines bool
	scanner      *bufio.Scanner
	currInst     string
}

func newParser(f *os.File) Parser {
	return Parser{
		HasMoreLines: true,
		scanner:      bufio.NewScanner(f),
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

func (p *Parser) currInstType() int {
	if strings.HasPrefix(p.currInst, "@") {
		return A_INSTRUCTION
	}
	if strings.HasPrefix(p.currInst, "(") && strings.HasSuffix(p.currInst, ")") {
		return L_INSTRUCTION
	}
	return C_INSTRUCTION
}

func (p *Parser) symbol() string {
	var after string
	if strings.HasPrefix(p.currInst, "@") {
		after, _ = strings.CutPrefix(p.currInst, "@")
	}
	if strings.HasPrefix(p.currInst, "(") {
		after, _ = strings.CutPrefix(p.currInst, "(")
		after, _ = strings.CutSuffix(after, ")")
	}
	return after
}

func (p *Parser) dest() string {
	if !strings.Contains(p.currInst, "=") {
		return "null"
	}
	return strings.Split(p.currInst, "=")[0]
}

func (p *Parser) comp() string {
	comp := p.currInst
	if strings.Contains(comp, "=") {
		comp = strings.Split(comp, "=")[1]
	}
	if strings.Contains(comp, ";") {
		comp = strings.Split(comp, ";")[0]
	}
	return comp
}

func (p *Parser) jump() string {
	if !strings.Contains(p.currInst, ";") {
		return "null"
	}
	return strings.Split(p.currInst, ";")[1]
}
