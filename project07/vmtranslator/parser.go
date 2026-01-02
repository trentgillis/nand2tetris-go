package vmtranslator

import (
	"bufio"
	"log"
	"os"
	"strings"
)

const (
	c_arithmetic = iota
	c_push
	c_pop
	c_label
	c_goto
	c_if
	c_function
	c_call
)

type parser struct {
	HasMoreLines bool
	currLine     string
	scanner      *bufio.Scanner
}

func newParser(file *os.File) parser {
	scanner := bufio.NewScanner(file)
	return parser{
		HasMoreLines: true,
		scanner:      scanner,
	}
}

func (p *parser) Advance() {
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

func (p *parser) commandType() int {
	switch cmd := strings.Split(p.currLine, " ")[0]; cmd {
	case "add", "sub", "neg", "eq", "gt", "lt", "and", "or", "not":
		return c_arithmetic
	case "push":
		return c_push
	case "pop":
		return c_pop
	default:
		return 0
	}
}

func (p *parser) arg1() string {
	parts := strings.Split(p.currLine, " ")
	if len(parts) < 2 {
		return ""
	}
	return parts[1]
}

func (p *parser) arg2() string {
	parts := strings.Split(p.currLine, " ")
	if len(parts) < 3 {
		return ""
	}
	return parts[2]
}
