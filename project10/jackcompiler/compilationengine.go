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
	ce.compileSubroutineBody()

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

// Performs syntax analysis and outputs XML for subroutine bodies
// '{' varDec* statements '}'
func (ce *compilationEngine) compileSubroutineBody() {
	fmt.Fprintf(ce.outf, "<subroutineBody>\n")

	ce.process("{")
	for ce.jt.currToken == "var" {
		ce.compileVarDec()
	}
	ce.compileStatements()
	ce.process("}")

	fmt.Fprintf(ce.outf, "</subroutineBody>\n")
}

// Performs syntax analysis and outputs XML for variable declarations in a subroutine body
// 'var' type varName (',' type varName)* ';'
func (ce *compilationEngine) compileVarDec() {
	fmt.Fprintf(ce.outf, "<varDec>\n")

	ce.process("var")
	ce.compileType()
	ce.compileIdentifier()
	for ce.jt.currToken == "," {
		ce.process(",")
		ce.compileIdentifier()
	}
	ce.process(";")

	fmt.Fprintf(ce.outf, "</varDec>\n")
}

// Performs syntax analysis and outputs XML for one or more statements
// (letStatement | ifStatement | whileStatement | doStatement | returnStatement)*
func (ce *compilationEngine) compileStatements() {
	fmt.Fprintf(ce.outf, "<statements>\n")
	for slices.Contains([]string{"let", "if", "while", "do", "return"}, ce.jt.currToken) {
		switch ce.jt.currToken {
		case "let":
			ce.compileLetStatement()
		case "if":
			ce.compileIfStatement()
		case "while":
			ce.compileWhileStatement()
		case "do":
			ce.compileDoStatement()
		case "return":
			ce.compileReturnStatement()
		default:
			log.Fatalf("Syntax error at token %s. Expected: let, if, while, do or return", ce.jt.currToken)
		}
	}
	fmt.Fprintf(ce.outf, "</statements>\n")
}

// Performs syntax analysis and outputs XML for a let statement
// 'let' varName ('[' expression ']')? '=' expression ';'
func (ce *compilationEngine) compileLetStatement() {
	fmt.Fprintf(ce.outf, "<letStatement>\n")

	ce.process("let")
	ce.compileIdentifier()
	if ce.jt.currToken == "[" {
		ce.process("[")
		ce.compileExpression()
		ce.process("]")
	}
	ce.process("=")
	ce.compileExpression()
	ce.process(";")

	fmt.Fprintf(ce.outf, "</letStatement>\n")
}

// Performs syntax analysis and outputs XML for an if statement
// 'if' '(' expression ')' '{' statements '}' ('else' '{' statements '}')?
func (ce *compilationEngine) compileIfStatement() {
	fmt.Fprintf(ce.outf, "<ifStatement>\n")

	ce.process("if")
	ce.process("(")
	ce.compileExpression()
	ce.process(")")
	ce.process("{")
	ce.compileStatements()
	ce.process("}")
	if ce.jt.currToken == "else" {
		ce.process("else")
		ce.process("{")
		ce.compileStatements()
		ce.process("}")
	}

	fmt.Fprintf(ce.outf, "</ifStatement>\n")
}

// Performs syntax analysis and outputs XML for a while statement
// 'while' '(' expression ')' '{' statements '}'
func (ce *compilationEngine) compileWhileStatement() {
	fmt.Fprintf(ce.outf, "<whileStatement>\n")

	ce.process("while")
	ce.process("(")
	ce.compileExpression()
	ce.process(")")
	ce.process("{")
	ce.compileStatements()
	ce.process("}")

	fmt.Fprintf(ce.outf, "</whileStatement>\n")
}

// Performs syntax analysis and outputs XML for a do statement
// 'do' subroutineCall ';'
func (ce *compilationEngine) compileDoStatement() {
	fmt.Fprintf(ce.outf, "<doStatement>\n")

	ce.process("do")
	ce.compileSubroutineCall()
	ce.process(";")

	fmt.Fprintf(ce.outf, "</doStatement>\n")
}

// Performs syntax analysis and outputs XML for a return statement
// 'return' expression? ';'
func (ce *compilationEngine) compileReturnStatement() {
	fmt.Fprintf(ce.outf, "<returnStatement>\n")

	ce.process("return")
	if ce.jt.currToken != ";" {
		ce.compileExpression()
	}
	ce.process(";")

	fmt.Fprintf(ce.outf, "</returnStatement>\n")
}

// Performs syntax analysis and outputs XML for a subroutine call
// subroutineName '(' expressionList ')' | (className | varName) '.' subroutineName '(' expressionList ')'
func (ce *compilationEngine) compileSubroutineCall() {
	ce.compileIdentifier()
	if ce.jt.currToken == "." {
		ce.process(".")
		ce.compileIdentifier()
	}
	ce.process("(")
	ce.compileExpressionList()
	ce.process(")")
}

// Performs syntax analysis and outputs XML for an expression list
// (expression (',' expression)*)?
func (ce *compilationEngine) compileExpressionList() {
	fmt.Fprintf(ce.outf, "<expressionList>\n")

	if ce.jt.currToken != ")" {
		ce.compileExpression()
		for ce.jt.currToken == "," {
			ce.process(",")
			ce.compileExpression()
		}
	}

	fmt.Fprintf(ce.outf, "</expressionList>\n")
}

// Performs syntax analysis and outputs XML for an expression
// term (op term)*
// TODO: handle expressions
func (ce *compilationEngine) compileExpression() {
	fmt.Fprintf(ce.outf, "<expression>\n")

	ce.compileTerm()
	for slices.Contains([]string{"+", "-", "*", "/", "&", "|", ">", "<", "="}, ce.jt.currToken) {
		ce.compileOp()
		ce.compileTerm()
	}

	fmt.Fprintf(ce.outf, "</expression>\n")
}

// Performs syntax analysis and outputs XML for an term
// integerConstant | stringConstant | keywordConstant | varName | varName'[' expression ']' |
// '(' expression ')' | (unaryOp term) | subroutineCall
// TODO: handle terms
func (ce *compilationEngine) compileTerm() {
	fmt.Fprintf(ce.outf, "<term>\n")

	if ce.jt.currToken == "(" {
		// Handle expression wrapped in parens
		ce.process("(")
		ce.compileExpression()
		ce.process(")")
	} else if len(ce.jt.lineTokens) > 0 && (ce.jt.lineTokens[0] == "." || ce.jt.lineTokens[0] == "(") {
		// Handle subroutine call case with lookahead
		ce.compileSubroutineCall()
	} else if len(ce.jt.lineTokens) > 0 && ce.jt.lineTokens[0] == "[" {
		// Handle array access with lookahead
		ce.compileIdentifier()
		ce.process("[")
		ce.compileExpression()
		ce.process("]")
	} else if slices.Contains([]string{"-", "~"}, ce.jt.currToken) {
		// Handle unary op term
		ce.compileUnaryOp()
		ce.compileTerm()
	} else {
		// Handle every else as an identifier. This includes all constants, keywords and varNames
		ce.compileIdentifier()
	}

	fmt.Fprintf(ce.outf, "</term>\n")
}

func (ce *compilationEngine) compileOp() {
	switch ce.jt.currToken {
	case "+":
		ce.process("+")
	case "-":
		ce.process("-")
	case "*":
		ce.process("*")
	case "/":
		ce.process("/")
	case "&":
		ce.process("&")
	case "|":
		ce.process("|")
	case "<":
		ce.process("<")
	case ">":
		ce.process(">")
	case "=":
		ce.process("=")
	}
}

func (ce *compilationEngine) compileUnaryOp() {
	switch ce.jt.currToken {
	case "-":
		ce.process("-")
	case "~":
		ce.process("~")
	}
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
