package vmtranslator

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

type codeWriter struct {
	outfile    *os.File
	fname      string
	strBuilder *strings.Builder
	numLabels  int
}

func newCodeWriter(outfile *os.File) codeWriter {
	var b strings.Builder
	fname, _ := strings.CutSuffix(filepath.Base(outfile.Name()), ".asm")
	return codeWriter{
		outfile:    outfile,
		fname:      fname,
		strBuilder: &b,
		numLabels:  0,
	}
}

func (cw *codeWriter) write(commandType int, arg1 string, arg2 string) {
	cw.strBuilder.Reset()
	switch commandType {
	case c_push:
		cw.writePush(arg1, arg2)
	case c_pop:
		cw.writePop(arg1, arg2)
	case c_arithmetic:
		cw.writeArithmetic(arg1)
	}
	cw.outfile.WriteString(cw.strBuilder.String())
}

func (cw *codeWriter) writePush(segment string, index string) {
	switch segment {
	case "constant":
		cw.writePushConstant(index)
	}
}

func (cw *codeWriter) writePop(segment string, index string) {
	switch segment {
	case "register":
		cw.writePopReg(index)
	}
}

func (cw *codeWriter) writeArithmetic(command string) {
	switch command {
	case "add":
		cw.writeAdd()
	case "sub":
		cw.writeSub()
	case "neg":
		cw.writeNeg()
	case "and":
		cw.writeAnd()
	case "or":
		cw.writeOr()
	case "not":
		cw.writeNot()
	case "eq":
		cw.writeEq()
	case "gt":
		cw.writeGt()
	case "lt":
		cw.writeLt()
	}
}

func (cw *codeWriter) writePushConstant(index string) {
	fmt.Fprintf(cw.strBuilder, "@%s\n", index)
	cw.strBuilder.WriteString("D=A\n")
	cw.strBuilder.WriteString("@SP\n")
	cw.strBuilder.WriteString("A=M\n")
	cw.strBuilder.WriteString("M=D\n")
	cw.incrementSp()
}

func (cw *codeWriter) writePopReg(index string) {
	cw.decrementSp()
	cw.strBuilder.WriteString("@SP\n")
	cw.strBuilder.WriteString("A=M\n")
	cw.strBuilder.WriteString("D=M\n")
	fmt.Fprintf(cw.strBuilder, "@R%s\n", index)
	cw.strBuilder.WriteString("M=D\n")
}

func (cw *codeWriter) writeAdd() {
	cw.writePop("register", "13")
	cw.writePop("register", "14")
	cw.strBuilder.WriteString("@R13\n")
	cw.strBuilder.WriteString("D=M\n")
	cw.strBuilder.WriteString("@R14\n")
	cw.strBuilder.WriteString("D=M+D\n")
	cw.strBuilder.WriteString("@SP\n")
	cw.strBuilder.WriteString("A=M\n")
	cw.strBuilder.WriteString("M=D\n")
	cw.incrementSp()
}

func (cw *codeWriter) writeSub() {
	cw.writePop("register", "13")
	cw.writePop("register", "14")
	cw.strBuilder.WriteString("@R13\n")
	cw.strBuilder.WriteString("D=M\n")
	cw.strBuilder.WriteString("@R14\n")
	cw.strBuilder.WriteString("D=M-D\n")
	cw.strBuilder.WriteString("@SP\n")
	cw.strBuilder.WriteString("A=M\n")
	cw.strBuilder.WriteString("M=D\n")
	cw.incrementSp()
}

func (cw *codeWriter) writeNeg() {
	cw.writePop("register", "13")
	cw.strBuilder.WriteString("@R13\n")
	cw.strBuilder.WriteString("D=-M\n")
	cw.strBuilder.WriteString("@SP\n")
	cw.strBuilder.WriteString("A=M\n")
	cw.strBuilder.WriteString("M=D\n")
	cw.incrementSp()
}

func (cw *codeWriter) writeAnd() {
	cw.writePop("register", "13")
	cw.writePop("register", "14")
	cw.strBuilder.WriteString("@R13\n")
	cw.strBuilder.WriteString("D=M\n")
	cw.strBuilder.WriteString("@R14\n")
	cw.strBuilder.WriteString("D=M&D\n")
	cw.strBuilder.WriteString("@SP\n")
	cw.strBuilder.WriteString("A=M\n")
	cw.strBuilder.WriteString("M=D\n")
	cw.incrementSp()
}

func (cw *codeWriter) writeOr() {
	cw.writePop("register", "13")
	cw.writePop("register", "14")
	cw.strBuilder.WriteString("@R13\n")
	cw.strBuilder.WriteString("D=M\n")
	cw.strBuilder.WriteString("@R14\n")
	cw.strBuilder.WriteString("D=M|D\n")
	cw.strBuilder.WriteString("@SP\n")
	cw.strBuilder.WriteString("A=M\n")
	cw.strBuilder.WriteString("M=D\n")
	cw.incrementSp()
}

func (cw *codeWriter) writeNot() {
	cw.writePop("register", "13")
	cw.strBuilder.WriteString("@R13\n")
	cw.strBuilder.WriteString("D=!M\n")
	cw.strBuilder.WriteString("@SP\n")
	cw.strBuilder.WriteString("A=M\n")
	cw.strBuilder.WriteString("M=D\n")
	cw.incrementSp()
}

func (cw *codeWriter) writeEq() {
	cw.numLabels += 1

	cw.writePop("register", "13")
	cw.writePop("register", "14")
	cw.strBuilder.WriteString("@R13\n")
	cw.strBuilder.WriteString("D=M\n")
	cw.strBuilder.WriteString("@R14\n")
	cw.strBuilder.WriteString("D=D-M\n")
	fmt.Fprintf(cw.strBuilder, "@%s.EQ.%d\n", cw.fname, cw.numLabels)
	cw.strBuilder.WriteString("D;JEQ\n")
	fmt.Fprintf(cw.strBuilder, "@%s.NE.%d\n", cw.fname, cw.numLabels)
	cw.strBuilder.WriteString("D;JNE\n")
	fmt.Fprintf(cw.strBuilder, "(%s.EQ.%d)\n", cw.fname, cw.numLabels)
	cw.strBuilder.WriteString("D=-1\n")
	fmt.Fprintf(cw.strBuilder, "@%s.EQ_END.%d\n", cw.fname, cw.numLabels)
	cw.strBuilder.WriteString("0;JEQ\n")
	fmt.Fprintf(cw.strBuilder, "(%s.NE.%d)\n", cw.fname, cw.numLabels)
	cw.strBuilder.WriteString("D=0\n")
	fmt.Fprintf(cw.strBuilder, "(%s.EQ_END.%d)\n", cw.fname, cw.numLabels)
	cw.strBuilder.WriteString("@SP\n")
	cw.strBuilder.WriteString("A=M\n")
	cw.strBuilder.WriteString("M=D\n")
	cw.incrementSp()
}

func (cw *codeWriter) writeLt() {
	cw.numLabels += 1

	cw.writePop("register", "13")
	cw.writePop("register", "14")
	cw.strBuilder.WriteString("@R13\n")
	cw.strBuilder.WriteString("D=M\n")
	cw.strBuilder.WriteString("@R14\n")
	cw.strBuilder.WriteString("D=M-D\n")
	fmt.Fprintf(cw.strBuilder, "@%s.LT.%d\n", cw.fname, cw.numLabels)
	cw.strBuilder.WriteString("D;JLT\n")
	fmt.Fprintf(cw.strBuilder, "@%s.GTE.%d\n", cw.fname, cw.numLabels)
	cw.strBuilder.WriteString("0;JEQ\n")
	fmt.Fprintf(cw.strBuilder, "(%s.LT.%d)\n", cw.fname, cw.numLabels)
	cw.strBuilder.WriteString("D=-1\n")
	fmt.Fprintf(cw.strBuilder, "@%s.LT_END.%d\n", cw.fname, cw.numLabels)
	cw.strBuilder.WriteString("0;JEQ\n")
	fmt.Fprintf(cw.strBuilder, "(%s.GTE.%d)\n", cw.fname, cw.numLabels)
	cw.strBuilder.WriteString("D=0\n")
	fmt.Fprintf(cw.strBuilder, "(%s.LT_END.%d)\n", cw.fname, cw.numLabels)
	cw.strBuilder.WriteString("@SP\n")
	cw.strBuilder.WriteString("A=M\n")
	cw.strBuilder.WriteString("M=D\n")
	cw.incrementSp()
}

func (cw *codeWriter) writeGt() {
	cw.numLabels += 1

	cw.writePop("register", "13")
	cw.writePop("register", "14")
	cw.strBuilder.WriteString("@R13\n")
	cw.strBuilder.WriteString("D=M\n")
	cw.strBuilder.WriteString("@R14\n")
	cw.strBuilder.WriteString("D=M-D\n")
	fmt.Fprintf(cw.strBuilder, "@%s.GT.%d\n", cw.fname, cw.numLabels)
	cw.strBuilder.WriteString("D;JGT\n")
	fmt.Fprintf(cw.strBuilder, "@%s.LTE.%d\n", cw.fname, cw.numLabels)
	cw.strBuilder.WriteString("0;JEQ\n")
	fmt.Fprintf(cw.strBuilder, "(%s.GT.%d)\n", cw.fname, cw.numLabels)
	cw.strBuilder.WriteString("D=-1\n")
	fmt.Fprintf(cw.strBuilder, "@%s.GT_END.%d\n", cw.fname, cw.numLabels)
	cw.strBuilder.WriteString("0;JEQ\n")
	fmt.Fprintf(cw.strBuilder, "(%s.LTE.%d)\n", cw.fname, cw.numLabels)
	cw.strBuilder.WriteString("D=0\n")
	fmt.Fprintf(cw.strBuilder, "(%s.GT_END.%d)\n", cw.fname, cw.numLabels)
	cw.strBuilder.WriteString("@SP\n")
	cw.strBuilder.WriteString("A=M\n")
	cw.strBuilder.WriteString("M=D\n")
	cw.incrementSp()
}

func (cw *codeWriter) incrementSp() {
	cw.strBuilder.WriteString("@SP\n")
	cw.strBuilder.WriteString("M=M+1\n")
}

func (cw *codeWriter) decrementSp() {
	cw.strBuilder.WriteString("@SP\n")
	cw.strBuilder.WriteString("M=M-1\n")
}
