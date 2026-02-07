package jackcompiler

import (
	"fmt"
	"log"
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
	if ce.jt.currToken != token {
		log.Fatalf("Syntax error at token: %s. Expected: %s\n", ce.jt.currToken, token)
	}

	ce.jt.printTokenXML()
	ce.jt.advance()
}

// Performs syntax analysis and outputs XML for class declaration. Entrypoint function for the
// compilation engine and should be called first to begin recursive descent of Jack programs
// 'class' className '{' classVarDec* subroutineDec* '}'
func (ce *compilationEngine) compileClass() {
	fmt.Fprintf(ce.outf, "<class>\n")

	ce.process("class")
	ce.compileIdentifier() // className
	ce.process("{")
	for slices.Contains([]string{"static", "field"}, ce.jt.currToken) {
		ce.compileClassVarDec()
	}
	for slices.Contains([]string{"function", "method", "constructor"}, ce.jt.currToken) {
		ce.compileSubroutine()
	}
	ce.process("}")

	fmt.Fprintf(ce.outf, "</class>\n")
}

// Performs syntax analysis and outputs XML for class variable declarations
// ('static' | 'field') type varName (',' varName)* ';'
func (ce *compilationEngine) compileClassVarDec() {
	fmt.Fprintf(ce.outf, "<classVarDec>\n")

	switch ce.jt.currToken {
	case "static":
		ce.process("static")
	case "field":
		ce.process("field")
	default:
		log.Fatalf("Syntax error at token %s. Expected: static or field", ce.jt.currToken)
	}

	ce.compileType()       // type
	ce.compileIdentifier() // varName
	for ce.jt.currToken == "," {
		ce.process(",")
		ce.compileIdentifier() // varName
	}
	ce.process(";")

	fmt.Fprintf(ce.outf, "</classVarDec>\n")
}

// Performs syntax analysis and outputs XML for subroutine declarations
// ('constructor' | 'function' | 'method') ('void' | type) subroutineName '(' parameterList ')' subroutineBody
func (ce *compilationEngine) compileSubroutine() {
	fmt.Fprintf(ce.outf, "<subroutineDec>\n")

	switch ce.jt.currToken {
	case "constructor":
		ce.process("constructor")
	case "method":
		ce.process("method")
	case "function":
		ce.process("function")
	default:
		// TODO: log fatal
		log.Fatalf("Syntax error at token %s. Expected: constructor, method or function", ce.jt.currToken)
	}

	ce.compileType()
	ce.compileIdentifier() // subroutineName
	ce.process("(")
	ce.compileParameterList()
	ce.process(")")

	fmt.Fprintf(ce.outf, "</subroutineDec>\n")
}

// Performs syntax analysis and outputs XML for parameter list declaration
// ((type varName) (',' varName)*)?
func (ce *compilationEngine) compileParameterList() {
	fmt.Fprintf(ce.outf, "<parameterList>\n")

	for ce.jt.currToken != ")" {
		ce.compileType()
		ce.compileIdentifier()
		if ce.jt.currToken == "," {
			ce.process(",")
		}
	}

	fmt.Fprintf(ce.outf, "</parameterList>\n")
}

func (ce *compilationEngine) compileSubroutineBody() {
	fmt.Fprintf(ce.outf, "<subroutineBody>\n")
	fmt.Fprintf(ce.outf, "</subroutineBody>\n")
}

func (ce *compilationEngine) compileType() {
	switch ce.jt.currToken {
	case "int":
		ce.process("int")
	case "char":
		ce.process("char")
	case "boolean":
		ce.process("boolean")
	default:
		ce.compileIdentifier()
	}
}

func (ce *compilationEngine) compileIdentifier() {
	ce.jt.printTokenXML()
	ce.jt.advance()
}
