package jackcompiler

import (
	"bufio"
	"log"
	"os"
	"regexp"
	"strings"
)

var JACK_SYMBOLS = []string{"{", "}", "(", ")", "[", "]", ".", ",", ";", "+", "-", "*", "/", "&", "|", "<", ">", "=", "~"}

type jackTokenizer struct {
	hasMoreTokens bool
	lineTokens    []string
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

func (jt *jackTokenizer) advance() {
	if len(jt.lineTokens) == 0 {
		jt.nextLine()
	} else {
		jt.nextToken()
	}
}

func (jt *jackTokenizer) nextToken() {
	jt.currToken = string(jt.lineTokens[0])
	jt.lineTokens = jt.lineTokens[1:]
}

func (jt *jackTokenizer) nextLine() {

	for jt.scanner.Scan() {
		line := strings.TrimSpace(jt.scanner.Text())
		if idx := strings.Index(line, "//"); idx != -1 {
			line = strings.TrimSpace(line[:idx])
		}
		if len(line) == 0 {
			continue
		}

		jt.lineTokens = getLineTokens(line)
		return
	}

	jt.hasMoreTokens = false
	if err := jt.scanner.Err(); err != nil {
		log.Fatal(err)
	}
}

func getLineTokens(line string) []string {
	lineTokens := []string{}
	re := regexp.MustCompile(`({|}|\(|\)|\[|\]|\.|,|;|\+|-|\*|/|&|\||<|>|=|~)`)

	for l := range strings.SplitSeq(re.ReplaceAllString(line, " $1 "), " ") {
		if len(l) > 0 {
			lineTokens = append(lineTokens, strings.TrimSpace(l))
		}
	}

	return lineTokens
}
