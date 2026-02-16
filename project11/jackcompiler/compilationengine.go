package jackcompiler

import (
	"fmt"
	"log"
	"os"
	"slices"
	"strconv"
)

type compilationEngine struct {
	jt         jackTokenizer
	vw         vmWriter
	className  string
	classSt    symbolTable
	routineSt  symbolTable
	ifCount    int
	whileCount int
	inf        *os.File
	outf       *os.File
}

func newCompilationEngine(inf *os.File, outf *os.File) compilationEngine {
	classSt := newSymbolTable()
	vw := newVmWriter(outf)
	jt := newJackTokenizer(inf, outf)
	jt.advance() // move to the first token
	return compilationEngine{inf: inf, outf: outf, jt: jt, vw: vw, classSt: classSt, ifCount: 0, whileCount: 0}
}

func (ce *compilationEngine) process(token string) {
	if ce.jt.currToken != token {
		log.Fatalf("%s - Syntax error at token: %s. Expected: %s\n", ce.outf.Name(), ce.jt.currToken, token)
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
		stEntry.kind = "this"
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
	if stEntry.kind == "this" {
		ce.classSt.fieldCount = stEntry.index + 1
	}
}

// Performs subroutine declaration compilation and outputs appropriate vm code
// ('constructor' | 'function' | 'method') ('void' | type) subroutineName '(' parameterList ')' subroutineBody
func (ce *compilationEngine) compileSubroutine() {
	ce.ifCount = 0
	ce.whileCount = 0
	ce.routineSt = newSymbolTable()

	switch ce.jt.currToken {
	case "constructor":
		ce.compileConstructor()
	case "method":
		ce.compileMethod()
	case "function":
		ce.compileFunction()
	default:
		log.Fatalf("Syntax error at token %s. Expected: constructor, method or function", ce.jt.currToken)
	}
}

func (ce *compilationEngine) compileConstructor() {
	ce.process("constructor")
	subroutineName := ce.compileSubroutineDeclaration()
	ce.compileSubroutineBody(subroutineName, "constructor")
}

func (ce *compilationEngine) compileFunction() {
	ce.process("function")
	subroutineName := ce.compileSubroutineDeclaration()
	ce.compileSubroutineBody(subroutineName, "function")
}

func (ce *compilationEngine) compileMethod() {
	ce.process("method")
	ce.routineSt.table["this"] = stEntry{
		name:     "this",
		kind:     "argument",
		dataType: ce.jt.currToken,
		index:    0,
	}
	ce.routineSt.argCount += 1

	subroutineName := ce.compileSubroutineDeclaration()
	ce.compileSubroutineBody(subroutineName, "method")
}

// Compiles generic subroutine declaration code that is shared by methods, functions and constructors and is
// therefore required for the compilation of all of the subroutine types. This includes the subroutine return
// type, name and parameter list.
func (ce *compilationEngine) compileSubroutineDeclaration() string {
	ce.compileType()
	subroutineName := ce.jt.currToken
	ce.jt.advance()
	ce.process("(")
	ce.compileParameterList()
	ce.process(")")

	return subroutineName
}

// Performs syntax analysis and outputs XML for parameter list declaration
// ((type varName) (',' varName)*)?
func (ce *compilationEngine) compileParameterList() {
	stEntry := stEntry{kind: "argument", index: ce.routineSt.argCount}

	for ce.jt.currToken != ")" {
		stEntry.dataType = ce.jt.currToken
		ce.jt.advance()
		stEntry.name = ce.jt.currToken
		ce.jt.advance()
		ce.routineSt.table[stEntry.name] = stEntry
		if ce.jt.currToken == "," {
			ce.process(",")
			stEntry.index += 1
		}
	}

}

// Performs syntax analysis and outputs XML for subroutine bodies
// '{' varDec* statements '}'
func (ce *compilationEngine) compileSubroutineBody(subroutineName string, subroutineType string) {
	nVars := 0
	ce.process("{")
	for ce.jt.currToken == "var" {
		nVars += ce.compileVarDec()
	}

	ce.vw.writeFunction(ce.className, subroutineName, nVars)
	if subroutineType == "constructor" {
		ce.vw.writePush(CONSTANT, ce.classSt.fieldCount)
		ce.vw.writeCall("Memory", "alloc", 1)
		ce.vw.writePop(POINTER, 0)
	}
	if subroutineType == "method" {
		ce.vw.writePush(ARGUMENT, 0)
		ce.vw.writePop(POINTER, 0)
	}

	ce.compileStatements()
	ce.process("}")
}

// Performs syntax analysis and outputs XML for variable declarations in a subroutine body
// 'var' type varName (',' type varName)* ';'
func (ce *compilationEngine) compileVarDec() int {
	nVars := 1
	stEntry := stEntry{kind: "local", index: ce.routineSt.localCount}
	ce.jt.advance()
	stEntry.dataType = ce.jt.currToken
	ce.jt.advance()
	stEntry.name = ce.jt.currToken
	ce.jt.advance()
	ce.routineSt.table[stEntry.name] = stEntry
	for ce.jt.currToken == "," {
		nVars += 1
		ce.process(",")
		stEntry.index += 1
		stEntry.name = ce.jt.currToken
		ce.jt.advance()
		ce.routineSt.table[stEntry.name] = stEntry
	}
	ce.process(";")
	ce.routineSt.localCount = stEntry.index + 1
	return nVars
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

// Performs compilation of let statements
// 'let' varName ('[' expression ']')? '=' expression ';'
func (ce *compilationEngine) compileLetStatement() {
	isLetArr := false
	ce.process("let")
	identifier, _ := ce.lookupVar(ce.jt.currToken)
	ce.jt.advance()
	if ce.jt.currToken == "[" {
		isLetArr = true
		ce.vw.writePush(segment(identifier.kind), identifier.index)
		ce.process("[")
		ce.compileExpression()
		ce.process("]")
		ce.vw.writeArithmetic(ADD)
	}
	ce.process("=")
	ce.compileExpression()
	if isLetArr {
		ce.vw.writePop(TEMP, 0)
		ce.vw.writePop(POINTER, 1)
		ce.vw.writePush(TEMP, 0)
		ce.vw.writePop(THAT, 0)
	} else {
		ce.vw.writePop(identifier.kind, identifier.index)
	}
	ce.process(";")
}

// Performs syntax analysis and outputs XML for an if statement
// 'if' '(' expression ')' '{' statements '}' ('else' '{' statements '}')?
func (ce *compilationEngine) compileIfStatement() {
	ifTrueLabel := fmt.Sprintf("IF_TRUE%d", ce.ifCount)
	ifFalseLabel := fmt.Sprintf("IF_FALSE%d", ce.ifCount)
	ifEndLabel := fmt.Sprintf("IF_END%d", ce.ifCount)
	ce.ifCount += 1

	ce.process("if")
	ce.process("(")
	ce.compileExpression()
	ce.process(")")
	ce.process("{")
	ce.vw.writeIf(ifTrueLabel)
	ce.vw.writeGoto(ifFalseLabel)
	ce.vw.writeLabel(ifTrueLabel)
	ce.compileStatements()
	ce.process("}")
	if ce.jt.currToken == "else" {
		ce.vw.writeGoto(ifEndLabel)
		ce.vw.writeLabel(ifFalseLabel)
		ce.process("else")
		ce.process("{")
		ce.compileStatements()
		ce.process("}")
		ce.vw.writeLabel(ifEndLabel)
	} else {
		ce.vw.writeLabel(ifFalseLabel)
	}
}

// Performs syntax analysis and outputs XML for a while statement
// 'while' '(' expression ')' '{' statements '}'
func (ce *compilationEngine) compileWhileStatement() {
	whileExpLabel := fmt.Sprintf("WHILE_EXP%d", ce.whileCount)
	whileEndLabel := fmt.Sprintf("WHILE_END%d", ce.whileCount)
	ce.whileCount += 1

	ce.vw.writeLabel(whileExpLabel)
	ce.process("while")
	ce.process("(")
	ce.compileExpression()
	ce.vw.writeArithmetic(NOT)
	ce.vw.writeIf(whileEndLabel)
	ce.process(")")
	ce.process("{")
	ce.compileStatements()
	ce.vw.writeGoto(whileExpLabel)
	ce.process("}")
	ce.vw.writeLabel(whileEndLabel)
}

// Performs syntax analysis and outputs XML for a do statement
// 'do' subroutineCall ';'
func (ce *compilationEngine) compileDoStatement() {
	ce.process("do")
	ce.compileExpression()
	ce.vw.writePop(TEMP, 0)
	ce.process(";")
}

// Performs syntax analysis and outputs XML for a return statement
// 'return' expression? ';'
func (ce *compilationEngine) compileReturnStatement() {
	ce.process("return")
	if ce.jt.currToken == ";" {
		ce.vw.writePush(CONSTANT, 0)
	} else {
		ce.compileExpression()
	}
	ce.vw.writeReturn()
	ce.process(";")
}

// Performs syntax analysis and outputs XML for a subroutine call
// subroutineName '(' expressionList ')' | (className | varName) '.' subroutineName '(' expressionList ')'
func (ce *compilationEngine) compileSubroutineCall() {
	var arg1, arg2 string
	nVars := 0

	arg1 = ce.jt.currToken
	ce.jt.advance()
	if ce.jt.currToken == "." {
		ce.process(".")

		classEntry, ok := ce.lookupVar(arg1)
		if ok {
			ce.vw.writePush(segment(classEntry.kind), classEntry.index)
			arg1 = classEntry.dataType
			nVars += 1
		}

		arg2 = ce.jt.currToken
		ce.jt.advance()
	} else {
		ce.vw.writePush(POINTER, 0)
		nVars += 1
		arg2 = arg1
		arg1 = ce.className
	}
	ce.process("(")
	nVars += ce.compileExpressionList()
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
		ce.process("(")
		ce.compileExpression()
		ce.process(")")
	} else if slices.Contains([]string{"-", "~"}, ce.jt.currToken) {
		ce.compileUnaryOp()
	} else if len(ce.jt.lineTokens) > 0 && (ce.jt.lineTokens[0] == "." || ce.jt.lineTokens[0] == "(") {
		ce.compileSubroutineCall()
	} else if len(ce.jt.lineTokens) > 0 && ce.jt.lineTokens[0] == "[" {
		ce.compileCurrentToken()
		ce.process("[")
		ce.compileExpression()
		ce.process("]")
	} else {
		switch tt := tokenType(ce.jt.currToken); tt {
		case TOKEN_INT_CONST:
			ce.compileIntConstant()
		case TOKEN_STRING_CONST:
			ce.compileStringLiteral()
		case TOKEN_KEYWORD:
			ce.compileKeyword()
		default:
			identifier, _ := ce.lookupVar(ce.jt.currToken)
			ce.vw.writePush(segment(identifier.kind), identifier.index)
		}
		ce.jt.advance()
	}
}

func (ce *compilationEngine) compileKeyword() {
	switch ce.jt.currToken {
	case "true":
		ce.vw.writePush(CONSTANT, 0)
		ce.vw.writeArithmetic(NOT)
	case "null,", "false":
		ce.vw.writePush(CONSTANT, 0)
	case "this":
		ce.vw.writePush(POINTER, 0)
	}
}

func (ce *compilationEngine) compileIntConstant() {
	val, err := strconv.Atoi(ce.jt.currToken)
	if err != nil {
		log.Fatal(err)
	}
	ce.vw.writePush(CONSTANT, val)
}

func (ce *compilationEngine) compileStringLiteral() {
	str := ce.jt.currToken[1 : len(ce.jt.currToken)-1]
	ce.vw.writePush(CONSTANT, len(str))
	ce.vw.writeCall("String", "new", 1)
	for _, c := range str {
		ce.vw.writePush(CONSTANT, int(c))
		ce.vw.writeCall("String", "appendChar", 2)
	}
}

func (ce *compilationEngine) compileOp(op string) {
	switch op {
	case "+":
		ce.vw.writeArithmetic(ADD)
	case "-":
		ce.vw.writeArithmetic(SUB)
	case "*":
		ce.vw.writeCall("Math", "multiply", 2)
	case "/":
		ce.vw.writeCall("Math", "divide", 2)
	case "&":
		ce.vw.writeArithmetic(AND)
	case "|":
		ce.vw.writeArithmetic(OR)
	case ">":
		ce.vw.writeArithmetic(GT)
	case "<":
		ce.vw.writeArithmetic(LT)
	case "=":
		ce.vw.writeArithmetic(EQ)
	}
}

func (ce *compilationEngine) compileUnaryOp() {
	op := ce.jt.currToken
	ce.jt.advance()
	ce.compileTerm()
	switch op {
	case "-":
		ce.vw.writeArithmetic(NEG)
	case "~":
		ce.vw.writeArithmetic(NOT)
	}
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

func (ce *compilationEngine) lookupVar(varName string) (stEntry, bool) {
	varEntry, ok := ce.routineSt.table[varName]
	if !ok {
		varEntry, ok = ce.classSt.table[varName]
	}

	return varEntry, ok
}

func (ce *compilationEngine) compileCurrentToken() {
	ce.jt.printTokenXML()
	ce.jt.advance()
}
