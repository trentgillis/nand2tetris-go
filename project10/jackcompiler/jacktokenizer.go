package jackcompiler

import (
	"bufio"
	"log"
	"os"
	"strings"
)

type jackTokenizer struct {
	hasMoreTokens bool
	currToken     string
	scanner       *bufio.Scanner
}

func newJackTokenizer(file *os.File, outf *os.File) jackTokenizer {
	scanner := bufio.NewScanner(file)
	return jackTokenizer{
		hasMoreTokens: true,
		scanner:       scanner,
	}
}

func (jt *jackTokenizer) Advance() {
	for jt.scanner.Scan() {
		line := strings.TrimSpace(jt.scanner.Text())
		if idx := strings.Index(line, "//"); idx != -1 {
			line = strings.TrimSpace(line[:idx])
		}
		if len(line) == 0 {
			continue
		}

		jt.currToken = line
		return
	}

	jt.hasMoreTokens = false
	if err := jt.scanner.Err(); err != nil {
		log.Fatal(err)
	}
}
