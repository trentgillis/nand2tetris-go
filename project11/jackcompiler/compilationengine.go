package jackcompiler

import (
	"log"
	"os"
	"slices"
	"strconv"
)

type compilationEngine struct {
	jt        jackTokenizer
	vw        vmWriter
	className string
	classSt   symbolTable
	routineSt symbolTable
	inf       *os.File
	outf      *os.File
}

func newCompilationEngine(inf *os.File, outf *os.File) compilationEngine {
	classSt := newSymbolTable()
	vw := newVmWriter(outf)
	jt := newJackTokenizer(inf, outf)
	jt.advance() // move to the first token
	return compilationEngine{inf: inf, outf: outf, jt: jt, vw: vw, classSt: classSt}
}

func (ce *compilationEngine) process(token string) {
	if ce.jt.currToken != token {
		log.Fatalf("Syntax error at token: %s. Expected: %s\n", ce.jt.currToken, token)
	}
	ce.jt.advance()
}

// Performs syntax analysis and outputs XML class declaration. Entrypoint function for the
// compilation engine and should be called first to begin recursive descent of Jack programs
// 'class' className '{' classVarDec* subroutineDec* '}'
func (ce *compilationEngine) compileClass() {
	ce.process("class")
	ce.className = ce.jt.currToken
	ce.jt.advance()
	ce.process("{")
	for slices.Contains([]string{"static", "field"}, ce.jt.currToken) {
		ce.compileClassVarDec()
	}
	for slices.Contains([]string{"function", "method", "constructor"}, ce.jt.currToken) {
		ce.compileSubroutine()
	}
	ce.process("}")
}

// Performs syntax analysis and outputs XML for class variable declarations
// ('static' | 'field') type varName (',' varName)* ';'
func (ce *compilationEngine) compileClassVarDec() {
	stEntry := stEntry{}

	switch ce.jt.currToken {
	case "static":
		stEntry.index = ce.classSt.staticCount
		stEntry.kind = "static"
		ce.jt.advance()
	case "field":
		stEntry.index = ce.classSt.fieldCount
		stEntry.kind = "field"
		ce.jt.advance()
	default:
		log.Fatalf("Syntax error at token %s. Expected: static or field", ce.jt.currToken)
	}

	stEntry.dataType = ce.jt.currToken
	ce.jt.advance()
	stEntry.name = ce.jt.currToken
	ce.jt.advance()
	ce.classSt.table[stEntry.name] = stEntry
	for ce.jt.currToken == "," {
		ce.process(",")
		stEntry.name = ce.jt.currToken
		stEntry.index += 1
		ce.jt.advance()
		ce.classSt.table[stEntry.name] = stEntry
	}
	ce.process(";")

	if stEntry.kind == "static" {
		ce.classSt.staticCount = stEntry.index + 1
	}
	if stEntry.kind == "field" {
		ce.classSt.fieldCount = stEntry.index + 1
	}
}

// Performs subroutine declaration compilation and outputs appropriate vm code
// ('constructor' | 'function' | 'method') ('void' | type) subroutineName '(' parameterList ')' subroutineBody
func (ce *compilationEngine) compileSubroutine() {
	ce.routineSt = newSymbolTable()

	switch ce.jt.currToken {
	case "constructor":
		// TODO: compile constructors
		ce.process("constructor")
	case "method":
		ce.compileMethod()
	case "function":
		ce.compileFunction()
	default:
		log.Fatalf("Syntax error at token %s. Expected: constructor, method or function", ce.jt.currToken)
	}
}

func (ce *compilationEngine) compileFunction() {
	ce.process("function")
	subroutineName, nVars := ce.compileSubroutineDeclaration()
	ce.vw.writeFunction(ce.className, subroutineName, nVars)
	ce.compileSubroutineBody()
}

func (ce *compilationEngine) compileMethod() {
	ce.process("method")
	ce.routineSt.table["this"] = stEntry{
		name:     "this",
		kind:     "arg",
		dataType: ce.jt.currToken,
		index:    0,
	}
	ce.routineSt.argCount += 1

	subroutineName, nVars := ce.compileSubroutineDeclaration()
	ce.vw.writeFunction(ce.className, subroutineName, nVars)
	ce.vw.writePush(ARGUMENT, 0)
	ce.vw.writePop(POINTER, 0)

	ce.compileSubroutineBody()
}

// Compiles generic subroutine declaration code that is shared by methods, functions and constructors and is
// therefore required for the compilation of all of the subroutine types. This includes the subroutine return
// type, name and parameter list.
func (ce *compilationEngine) compileSubroutineDeclaration() (string, int) {
	ce.compileType()
	subroutineName := ce.jt.currToken
	ce.jt.advance()
	ce.process("(")
	nVars := ce.compileParameterList()
	ce.process(")")

	return subroutineName, nVars
}

// Performs syntax analysis and outputs XML for parameter list declaration
// ((type varName) (',' varName)*)?
func (ce *compilationEngine) compileParameterList() int {
	nVars := 0
	stEntry := stEntry{kind: "arg", index: ce.routineSt.argCount}

	for ce.jt.currToken != ")" {
		nVars += 1
		stEntry.dataType = ce.jt.currToken
		ce.jt.advance()
		stEntry.name = ce.jt.currToken
		ce.jt.advance()
		ce.routineSt.table[stEntry.name] = stEntry
		if ce.jt.currToken == "," {
			ce.process(",")
			nVars += 1
			stEntry.index += 1
		}
	}

	return nVars
}

// Performs syntax analysis and outputs XML for subroutine bodies
// '{' varDec* statements '}'
func (ce *compilationEngine) compileSubroutineBody() {
	ce.process("{")
	for ce.jt.currToken == "var" {
		ce.compileVarDec()
	}
	ce.compileStatements()
	ce.process("}")
}

// Performs syntax analysis and outputs XML for variable declarations in a subroutine body
// 'var' type varName (',' type varName)* ';'
func (ce *compilationEngine) compileVarDec() {
	stEntry := stEntry{kind: "var", index: ce.routineSt.varCount}
	ce.jt.advance()
	stEntry.dataType = ce.jt.currToken
	ce.jt.advance()
	stEntry.name = ce.jt.currToken
	ce.jt.advance()
	ce.routineSt.table[stEntry.name] = stEntry
	for ce.jt.currToken == "," {
		ce.process(",")
		stEntry.index += 1
		stEntry.name = ce.jt.currToken
		ce.jt.advance()
		ce.routineSt.table[stEntry.name] = stEntry
	}
	ce.process(";")
	ce.routineSt.varCount = stEntry.index + 1
}

// Performs syntax analysis and outputs XML for one or more statements
// (letStatement | ifStatement | whileStatement | doStatement | returnStatement)*
func (ce *compilationEngine) compileStatements() {
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
}

// Performs syntax analysis and outputs XML for a let statement
// 'let' varName ('[' expression ']')? '=' expression ';'
func (ce *compilationEngine) compileLetStatement() {
	ce.process("let")
	ce.compileCurrentToken()
	if ce.jt.currToken == "[" {
		ce.process("[")
		ce.compileExpression()
		ce.process("]")
	}
	ce.process("=")
	ce.compileExpression()
	ce.process(";")
}

// Performs syntax analysis and outputs XML for an if statement
// 'if' '(' expression ')' '{' statements '}' ('else' '{' statements '}')?
func (ce *compilationEngine) compileIfStatement() {
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
}

// Performs syntax analysis and outputs XML for a while statement
// 'while' '(' expression ')' '{' statements '}'
func (ce *compilationEngine) compileWhileStatement() {
	ce.process("while")
	ce.process("(")
	ce.compileExpression()
	ce.process(")")
	ce.process("{")
	ce.compileStatements()
	ce.process("}")
}

// Performs syntax analysis and outputs XML for a do statement
// 'do' subroutineCall ';'
func (ce *compilationEngine) compileDoStatement() {
	ce.process("do")
	ce.compileSubroutineCall()
	ce.process(";")
}

// Performs syntax analysis and outputs XML for a return statement
// 'return' expression? ';'
func (ce *compilationEngine) compileReturnStatement() {
	ce.process("return")
	ce.vw.writeReturn()
	if ce.jt.currToken != ";" {
		ce.compileExpression()
	}
	ce.process(";")
}

// Performs syntax analysis and outputs XML for a subroutine call
// subroutineName '(' expressionList ')' | (className | varName) '.' subroutineName '(' expressionList ')'
func (ce *compilationEngine) compileSubroutineCall() {
	var arg1, arg2 string
	var nVars int

	arg1 = ce.jt.currToken
	ce.jt.advance()
	if ce.jt.currToken == "." {
		ce.process(".")
		arg2 = ce.jt.currToken
		ce.jt.advance()
	}
	ce.process("(")
	nVars = ce.compileExpressionList()
	ce.process(")")

	ce.vw.writeCall(arg1, arg2, nVars)
}

// Performs syntax analysis and outputs XML for an expression list
// (expression (',' expression)*)?
func (ce *compilationEngine) compileExpressionList() int {
	nVars := 0
	if ce.jt.currToken != ")" {
		nVars += 1
		ce.compileExpression()
		for ce.jt.currToken == "," {
			nVars += 1
			ce.process(",")
			ce.compileExpression()
		}
	}

	return nVars
}

// Performs syntax analysis and outputs XML for an expression
// term (op term)*
func (ce *compilationEngine) compileExpression() {
	ce.compileTerm()
	for slices.Contains([]string{"+", "-", "*", "/", "&", "|", ">", "<", "="}, ce.jt.currToken) {
		op := ce.jt.currToken
		ce.jt.advance()
		ce.compileTerm()
		ce.compileOp(op)
	}
}

// Compiles and outputs vm code for a term
// integerConstant | stringConstant | keywordConstant | varName | varName'[' expression ']' |
// '(' expression ')' | (unaryOp term) | subroutineCall
func (ce *compilationEngine) compileTerm() {
	if ce.jt.currToken == "(" {
		// Handle expression wrapped in parens
		ce.process("(")
		ce.compileExpression()
		ce.process(")")
	} else if len(ce.jt.lineTokens) > 0 && (ce.jt.lineTokens[0] == "." || ce.jt.lineTokens[0] == "(") {
		// Handle subroutine call case with lookahead
		ce.compileSubroutineCall()
	} else if len(ce.jt.lineTokens) > 0 && ce.jt.lineTokens[0] == "[" {
		// TODO: compile array access
		ce.compileCurrentToken()
		ce.process("[")
		ce.compileExpression()
		ce.process("]")
	} else if slices.Contains([]string{"-", "~"}, ce.jt.currToken) {
		ce.compileUnaryOp()
		ce.compileTerm()
	} else {
		if tokenType(ce.jt.currToken) == TOKEN_INT_CONST {
			val, err := strconv.Atoi(ce.jt.currToken)
			if err != nil {
				log.Fatal(err)
			}
			ce.vw.writePush(CONSTANT, val)
		} else if tokenType(ce.jt.currToken) == TOKEN_STRING_CONST {
			// TODO: write string const
		} else {
			// TODO: lookup and write identifier
			identifier, ok := ce.routineSt.table[ce.jt.currToken]
			if !ok {
				identifier = ce.classSt.table[ce.jt.currToken]
			}
			ce.vw.writePush(segment(identifier.kind), identifier.index)
		}
		ce.jt.advance()
	}
}

func (ce *compilationEngine) compileOp(op string) {
	// "+", "-", "*", "/", "&", "|", ">", "<", "="
	switch op {
	case "+":
		ce.vw.writeArithmetic(ADD)
	case "-":
		ce.vw.writeArithmetic(SUB)
	case "*":
		ce.vw.writeCall("Math", "multiply", 2)
	}
}

func (ce *compilationEngine) compileUnaryOp() {
	ce.process(ce.jt.currToken)
}

func (ce *compilationEngine) compileType() {
	switch ce.jt.currToken {
	case "void":
		ce.process("void")
	case "int":
		ce.process("int")
	case "char":
		ce.process("char")
	case "boolean":
		ce.process("boolean")
	default:
		// Type is a className, simply advance to the next token
		ce.jt.advance()
	}
}

func (ce *compilationEngine) compileCurrentToken() {
	ce.jt.printTokenXML()
	ce.jt.advance()
}
