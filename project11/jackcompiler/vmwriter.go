package jackcompiler

import (
	"fmt"
	"os"
)

type segment string

const (
	ARGUMENT = "argument"
	LOCAL    = "local"
	STATIC   = "static"
	THIS     = "this"
	THAT     = "that"
	POINTER  = "pointer"
	TEMP     = "temp"
	CONSTANT = "constant"
)

type arithmeticCommand string

const (
	ADD = "add"
	SUB = "sub"
	NEG = "neg"
	EQ  = "eq"
	GT  = "gt"
	LT  = "lt"
	AND = "and"
	OR  = "or"
	NOT = "not"
)

type vmWriter struct {
	outf *os.File
}

func newVmWriter(outf *os.File) vmWriter {
	return vmWriter{
		outf: outf,
	}
}

func (vw *vmWriter) writePush(s segment, i int) {
	fmt.Fprintf(vw.outf, "push %v %d\n", s, i)
}

func (vw *vmWriter) writePop(s string, i int) {
	fmt.Fprintf(vw.outf, "pop %v %d\n", s, i)
}

func (vw *vmWriter) writeArithmetic(command arithmeticCommand) {
	fmt.Fprintf(vw.outf, "%s\n", command)
}

func (vw *vmWriter) writeFunction(className string, subroutineName string, nVars int) {
	fmt.Fprintf(vw.outf, "function %s.%s %d\n", className, subroutineName, nVars)
}

func (vw *vmWriter) writeCall(className string, subroutineName string, nVars int) {
	fmt.Fprintf(vw.outf, "call %s.%s %d\n", className, subroutineName, nVars)
}

func (vw *vmWriter) writeReturn() {
	fmt.Fprintf(vw.outf, "return\n")
}
