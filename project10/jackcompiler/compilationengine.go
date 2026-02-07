package jackcompiler

import (
	"fmt"
	"os"
	"slices"
)

type compilationEngine struct {
	jt   jackTokenizer
	inf  *os.File
	outf *os.File
}

func newCompilationEngine(inf *os.File, outf *os.File) compilationEngine {
	jt := newJackTokenizer(inf, outf)
	jt.advance() // move to the first token
	return compilationEngine{inf: inf, outf: outf, jt: jt}
}

func (ce *compilationEngine) process(token string) {
	fmt.Printf("%s - %s\n", token, ce.jt.currToken)
	if ce.jt.currToken != token {
		// TODO: fatal
		fmt.Printf("Syntax error at token: %s. Expected: %s\n", ce.jt.currToken, token)
	}

	ce.jt.printTokenXML()
	ce.jt.advance()
}

func (ce *compilationEngine) compileClass() {
	fmt.Fprintf(ce.outf, "<class>\n")
	ce.process("class")
	ce.compileTerm() // class name
	ce.process("{")
	for slices.Contains([]string{"static", "field"}, ce.jt.currToken) {
		ce.compileClassVarDec()
	}
	// for slices.Contains([]string{"function", "method", "constructor"}, ce.jt.currToken) {
	// 	ce.compileSubroutine()
	// }
	// ce.process("}")
	fmt.Fprintf(ce.outf, "</class>\n")
}

func (ce *compilationEngine) compileClassVarDec() {
	fmt.Fprintln(ce.outf, "<classVarDec>")

	switch ce.jt.currToken {
	case "static":
		ce.process("static")
	case "field":
		ce.process("field")
	default:
		// TODO: log fatal
		fmt.Printf("Syntax error at token %s", ce.jt.currToken)
	}

	ce.compileTerm() // type
	ce.compileTerm() // var name
	for ce.jt.currToken == "," {
		ce.process(",")
		ce.compileTerm() // var name
	}
	ce.process(";")

	fmt.Fprintln(ce.outf, "</classVarDec>")
}

func (ce *compilationEngine) compileSubroutine() {
	fmt.Fprintln(ce.outf, "<subroutineDec>")
	fmt.Fprintln(ce.outf, "</subroutineDec>")
}

func (ce *compilationEngine) compileTerm() {
	ce.jt.printTokenXML()
	ce.jt.advance()
}
