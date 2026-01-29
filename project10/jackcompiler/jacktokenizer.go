package jackcompiler

import (
	"bufio"
	"fmt"
	"log"
	"os"
	"regexp"
	"strings"
)

const (
	TOKEN_KEYWORD      = "keyword"
	TOKEN_SYMBOL       = "symbol"
	TOKEN_IDENTIFIER   = "identifier"
	TOKEN_INT_CONST    = "integerConstant"
	TOKEN_STRING_CONST = "stringConstant"
)

var JACK_SYMBOLS = map[string]string{
	"{": "{",
	"}": "}",
	"(": "(",
	")": ")",
	"[": "[",
	"]": "]",
	".": ".",
	",": ",",
	";": ";",
	"+": "+",
	"-": "-",
	"*": "*",
	"/": "/",
	"&": "&amp;",
	"|": "|",
	"<": "&lt;",
	">": "&gt;",
	"=": "=",
	"~": "~",
}
var JACK_KEYWORDS = map[string]string{
	"class":       "class",
	"method":      "method",
	"function":    "function",
	"constructor": "constructor",
	"int":         "int",
	"boolean":     "boolean",
	"char":        "char",
	"void":        "void",
	"var":         "var",
	"static":      "static",
	"field":       "field",
	"let":         "let",
	"do":          "do",
	"if":          "if",
	"else":        "else",
	"while":       "while",
	"return":      "return",
	"true":        "true",
	"false":       "false",
	"null":        "null",
	"this":        "this",
}

type jackTokenizer struct {
	hasMoreTokens bool
	lineTokens    []string
	currToken     string
	scanner       *bufio.Scanner
	outf          *os.File
}

func newJackTokenizer(file *os.File, outf *os.File) jackTokenizer {
	scanner := bufio.NewScanner(file)
	return jackTokenizer{
		hasMoreTokens: true,
		scanner:       scanner,
		outf:          outf,
	}
}

func (jt *jackTokenizer) advance() {
	if len(jt.lineTokens) == 0 {
		jt.nextLine()
	}
	if jt.hasMoreTokens {
		jt.nextToken()
	}
}

func (jt *jackTokenizer) printTokenXML() {
	token := jt.currToken
	tType := tokenType(jt.currToken)

	if tType == TOKEN_STRING_CONST {
		token = token[1 : len(token)-1]
	}

	if tType == TOKEN_SYMBOL {
		token = JACK_SYMBOLS[token]
	}

	fmt.Fprintf(jt.outf, "<%s> %s </%s>\n", tType, token, tType)
}

func tokenType(token string) string {
	if _, ok := JACK_KEYWORDS[token]; ok {
		return TOKEN_KEYWORD
	}

	if _, ok := JACK_SYMBOLS[token]; ok {
		return TOKEN_SYMBOL
	}

	if strings.HasPrefix(token, "\"") && strings.HasSuffix(token, "\"") {
		return TOKEN_STRING_CONST
	}

	intRegex := regexp.MustCompile(`[0-9]+`)
	if intRegex.MatchString(token) {
		return TOKEN_INT_CONST
	}

	return TOKEN_IDENTIFIER
}

func (jt *jackTokenizer) nextToken() {
	skip := 1
	jt.currToken = string(jt.lineTokens[0])

	// Handles string constant tokens by ensuring we use the entire string constant tokens
	// including any spaces in the string
	if strings.HasPrefix(jt.currToken, "\"") {
		for !strings.HasSuffix(jt.currToken, "\"") {
			jt.currToken = fmt.Sprintf("%s %s", jt.currToken, jt.lineTokens[skip])
			skip += 1
		}
	}

	jt.lineTokens = jt.lineTokens[skip:]
}

func (jt *jackTokenizer) nextLine() {
	multiLineComment := false

	for jt.scanner.Scan() {
		line := strings.TrimSpace(jt.scanner.Text())
		// Remove single line comments
		if idx := strings.Index(line, "//"); idx != -1 {
			line = strings.TrimSpace(line[:idx])
		}
		// Handle multi-line comments
		if strings.HasPrefix(line, "/**") {
			if !strings.HasSuffix(line, "*/") {
				multiLineComment = true
			}
			continue
		}
		if multiLineComment {
			if strings.HasSuffix(line, "*/") {
				multiLineComment = false
			}
			continue
		}
		// Skip empty lines
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
