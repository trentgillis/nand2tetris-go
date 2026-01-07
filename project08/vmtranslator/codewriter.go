package vmtranslator

import (
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strconv"
	"strings"
)

const (
	tempBase    = 5
	pointerBase = 3
	regR13      = "R13"
	regR14      = "R14"
	regR15      = "R15"
)

type codeWriter struct {
	outfile         *os.File
	fname           string
	strBuilder      *strings.Builder
	numLabels       int
	segmentMappings map[string]string
}

func newCodeWriter(outfile *os.File) codeWriter {
	segmentMappings := map[string]string{
		"local":    "LCL",
		"argument": "ARG",
		"this":     "THIS",
		"that":     "THAT",
	}

	var b strings.Builder
	fname, _ := strings.CutSuffix(filepath.Base(outfile.Name()), ".asm")

	return codeWriter{
		outfile:         outfile,
		fname:           fname,
		strBuilder:      &b,
		numLabels:       0,
		segmentMappings: segmentMappings,
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
	case c_label:
		cw.writeLabel(arg1)
	case c_goto:
		cw.writeGoto(arg1)
	case c_if:
		cw.writeIf(arg1)
	case c_function:
		nVars, err := strconv.Atoi(arg2)
		if err != nil {
			log.Fatal(err)
		}
		cw.writeFunction(arg1, nVars)
	case c_return:
		cw.writeReturn()
	}

	if _, err := cw.outfile.WriteString(cw.strBuilder.String()); err != nil {
		log.Fatal(err)
	}
}

func (cw *codeWriter) writePush(segment string, index string) {
	switch segment {
	case "constant":
		cw.writePushConstant(index)
	case "static":
		cw.writePushStatic(index)
	case "local", "argument", "this", "that":
		cw.writePushSegment(segment, index)
	case "temp":
		cw.writePushTemp(index)
	case "pointer":
		cw.writePushPointer(index)
	}
}

func (cw *codeWriter) writePop(segment string, index string) {
	switch segment {
	case "static":
		cw.writePopStatic(index)
	case "local", "argument", "this", "that":
		cw.writePopSegment(segment, index)
	case "temp":
		cw.writePopTemp(index)
	case "pointer":
		cw.writePopPointer(index)
	}
}

func (cw *codeWriter) writeArithmetic(command string) {
	switch command {
	case "add", "sub", "and", "or":
		cw.writeTwoOpArithmetic(command)
	case "neg", "not":
		cw.writeOneOpArithmetic(command)
	case "eq", "gt", "lt":
		cw.writeLogical(command)
	}
}

func (cw *codeWriter) writeLabel(label string) {
	fmt.Fprintf(cw.strBuilder, "(%s)\n", label)
}

func (cw *codeWriter) writeGoto(label string) {
	fmt.Fprintf(cw.strBuilder, "@%s\n", label)
	cw.strBuilder.WriteString("0;JEQ\n")
}

func (cw *codeWriter) writeIf(label string) {
	cw.strBuilder.WriteString("@SP\n")
	cw.strBuilder.WriteString("AM=M-1\n")
	cw.strBuilder.WriteString("D=M\n")
	fmt.Fprintf(cw.strBuilder, "@%s\n", label)
	cw.strBuilder.WriteString("D;JNE")
}

func (cw *codeWriter) writeFunction(fnName string, nVars int) {
	cw.writeLabel(fnName)
	for range nVars {
		cw.writePushConstant("0")
	}
}

func (cw *codeWriter) writeReturn() {
	// Get a reference to the start of caller's function frame
	cw.strBuilder.WriteString("@LCL\n")
	cw.strBuilder.WriteString("D=M\n")
	cw.strBuilder.WriteString("@frame\n")
	cw.strBuilder.WriteString("M=D\n")

	// Pop top value of the stack into argument 0 for use by the caller
	cw.writePopSegment("argument", "0")

	// Reposition the stack pointer to the appropriate position in the caller (@ARG+1)
	cw.strBuilder.WriteString("@ARG\n")
	cw.strBuilder.WriteString("D=M\n")
	cw.strBuilder.WriteString("@SP\n")
	cw.strBuilder.WriteString("M=D+1\n")

	// Reposition THAT pointer
	cw.strBuilder.WriteString("@1\n")
	cw.strBuilder.WriteString("D=A\n")
	cw.strBuilder.WriteString("@frame\n")
	cw.strBuilder.WriteString("A=M-D\n")
	cw.strBuilder.WriteString("D=M\n")
	cw.strBuilder.WriteString("@THAT\n")
	cw.strBuilder.WriteString("M=D\n")

	// Reposition THIS pointer
	cw.strBuilder.WriteString("@2\n")
	cw.strBuilder.WriteString("D=A\n")
	cw.strBuilder.WriteString("@frame\n")
	cw.strBuilder.WriteString("A=M-D\n")
	cw.strBuilder.WriteString("D=M\n")
	cw.strBuilder.WriteString("@THIS\n")
	cw.strBuilder.WriteString("M=D\n")

	// Reposition ARG pointer
	cw.strBuilder.WriteString("@3\n")
	cw.strBuilder.WriteString("D=A\n")
	cw.strBuilder.WriteString("@frame\n")
	cw.strBuilder.WriteString("A=M-D\n")
	cw.strBuilder.WriteString("D=M\n")
	cw.strBuilder.WriteString("@ARG\n")
	cw.strBuilder.WriteString("M=D\n")

	// Reposition LCL pointer
	cw.strBuilder.WriteString("@4\n")
	cw.strBuilder.WriteString("D=A\n")
	cw.strBuilder.WriteString("@frame\n")
	cw.strBuilder.WriteString("A=M-D\n")
	cw.strBuilder.WriteString("D=M\n")
	cw.strBuilder.WriteString("@LCL\n")
	cw.strBuilder.WriteString("M=D\n")

	// Calculate and goto return address
	cw.strBuilder.WriteString("@5\n")
	cw.strBuilder.WriteString("D=A\n")
	cw.strBuilder.WriteString("@frame\n")
	cw.strBuilder.WriteString("A=M-D\n")
	cw.strBuilder.WriteString("0;JMP\n")
}

func (cw *codeWriter) writePushConstant(index string) {
	fmt.Fprintf(cw.strBuilder, "@%s\n", index)
	cw.strBuilder.WriteString("D=A\n")
	cw.strBuilder.WriteString("@SP\n")
	cw.strBuilder.WriteString("A=M\n")
	cw.strBuilder.WriteString("M=D\n")
	cw.strBuilder.WriteString("@SP\n")
	cw.strBuilder.WriteString("M=M+1\n")
}

func (cw *codeWriter) writePushSegment(segment string, index string) {
	memVar, _ := cw.segmentMappings[segment]

	fmt.Fprintf(cw.strBuilder, "@%s\n", memVar)
	cw.strBuilder.WriteString("D=M\n")
	fmt.Fprintf(cw.strBuilder, "@%s\n", index)
	cw.strBuilder.WriteString("A=D+A\n")
	cw.strBuilder.WriteString("D=M\n")
	cw.strBuilder.WriteString("@SP\n")
	cw.strBuilder.WriteString("A=M\n")
	cw.strBuilder.WriteString("M=D\n")
	cw.strBuilder.WriteString("@SP\n")
	cw.strBuilder.WriteString("M=M+1\n")
}

func (cw *codeWriter) writePushStatic(index string) {
	fmt.Fprintf(cw.strBuilder, "@%s.%s\n", cw.fname, index)
	cw.strBuilder.WriteString("D=M\n")
	cw.strBuilder.WriteString("@SP\n")
	cw.strBuilder.WriteString("A=M\n")
	cw.strBuilder.WriteString("M=D\n")
	cw.strBuilder.WriteString("@SP\n")
	cw.strBuilder.WriteString("M=M+1\n")
}

func (cw *codeWriter) writePushTemp(index string) {
	fmt.Fprintf(cw.strBuilder, "@%d\n", tempBase)
	cw.strBuilder.WriteString("D=A\n")
	fmt.Fprintf(cw.strBuilder, "@%s\n", index)
	cw.strBuilder.WriteString("A=D+A\n")
	cw.strBuilder.WriteString("D=M\n")
	cw.strBuilder.WriteString("@SP\n")
	cw.strBuilder.WriteString("A=M\n")
	cw.strBuilder.WriteString("M=D\n")
	cw.strBuilder.WriteString("@SP\n")
	cw.strBuilder.WriteString("M=M+1\n")
}

func (cw *codeWriter) writePushPointer(index string) {
	fmt.Fprintf(cw.strBuilder, "@%d\n", pointerBase)
	cw.strBuilder.WriteString("D=A\n")
	fmt.Fprintf(cw.strBuilder, "@%s\n", index)
	cw.strBuilder.WriteString("A=D+A\n")
	cw.strBuilder.WriteString("D=M\n")
	cw.strBuilder.WriteString("@SP\n")
	cw.strBuilder.WriteString("A=M\n")
	cw.strBuilder.WriteString("M=D\n")
	cw.strBuilder.WriteString("@SP\n")
	cw.strBuilder.WriteString("M=M+1\n")
}

func (cw *codeWriter) writePopSegment(segment string, index string) {
	memVar, _ := cw.segmentMappings[segment]

	fmt.Fprintf(cw.strBuilder, "@%s\n", memVar)
	cw.strBuilder.WriteString("D=M\n")
	fmt.Fprintf(cw.strBuilder, "@%s\n", index)
	cw.strBuilder.WriteString("D=D+A\n")
	fmt.Fprintf(cw.strBuilder, "@%s\n", regR15)
	cw.strBuilder.WriteString("M=D\n")
	cw.strBuilder.WriteString("@SP\n")
	cw.strBuilder.WriteString("AM=M-1\n")
	cw.strBuilder.WriteString("D=M\n")
	fmt.Fprintf(cw.strBuilder, "@%s\n", regR15)
	cw.strBuilder.WriteString("A=M\n")
	cw.strBuilder.WriteString("M=D\n")
}

func (cw *codeWriter) writePopStatic(index string) {
	cw.strBuilder.WriteString("@SP\n")
	cw.strBuilder.WriteString("AM=M-1\n")
	cw.strBuilder.WriteString("D=M\n")
	fmt.Fprintf(cw.strBuilder, "@%s.%s\n", cw.fname, index)
	cw.strBuilder.WriteString("M=D\n")
}

func (cw *codeWriter) writePopTemp(index string) {
	fmt.Fprintf(cw.strBuilder, "@%d\n", tempBase)
	cw.strBuilder.WriteString("D=A\n")
	fmt.Fprintf(cw.strBuilder, "@%s\n", index)
	cw.strBuilder.WriteString("D=D+A\n")
	fmt.Fprintf(cw.strBuilder, "@%s\n", regR15)
	cw.strBuilder.WriteString("M=D\n")
	cw.strBuilder.WriteString("@SP\n")
	cw.strBuilder.WriteString("AM=M-1\n")
	cw.strBuilder.WriteString("D=M\n")
	fmt.Fprintf(cw.strBuilder, "@%s\n", regR15)
	cw.strBuilder.WriteString("A=M\n")
	cw.strBuilder.WriteString("M=D\n")
}

func (cw *codeWriter) writePopPointer(index string) {
	fmt.Fprintf(cw.strBuilder, "@%d\n", pointerBase)
	cw.strBuilder.WriteString("D=A\n")
	fmt.Fprintf(cw.strBuilder, "@%s\n", index)
	cw.strBuilder.WriteString("D=D+A\n")
	fmt.Fprintf(cw.strBuilder, "@%s\n", regR15)
	cw.strBuilder.WriteString("M=D\n")
	cw.strBuilder.WriteString("@SP\n")
	cw.strBuilder.WriteString("AM=M-1\n")
	cw.strBuilder.WriteString("D=M\n")
	fmt.Fprintf(cw.strBuilder, "@%s\n", regR15)
	cw.strBuilder.WriteString("A=M\n")
	cw.strBuilder.WriteString("M=D\n")
}

func (cw *codeWriter) writeOneOpArithmetic(command string) {
	opMap := map[string]string{
		"neg": "-",
		"not": "!",
	}
	op, _ := opMap[command]

	cw.strBuilder.WriteString("@SP\n")
	cw.strBuilder.WriteString("A=M-1\n")
	fmt.Fprintf(cw.strBuilder, "M=%sM\n", op)
}

func (cw *codeWriter) writeTwoOpArithmetic(command string) {
	opMap := map[string]string{
		"add": "+",
		"sub": "-",
		"and": "&",
		"or":  "|",
	}
	op, _ := opMap[command]

	cw.strBuilder.WriteString("@SP\n")
	cw.strBuilder.WriteString("AM=M-1\n")
	cw.strBuilder.WriteString("D=M\n")
	cw.strBuilder.WriteString("A=A-1\n")
	fmt.Fprintf(cw.strBuilder, "M=M%sD\n", op)
}

func (cw *codeWriter) writeLogical(command string) {
	cw.numLabels += 1
	jmpMap := map[string]string{
		"eq": "JEQ",
		"lt": "JLT",
		"gt": "JGT",
	}
	jmp, _ := jmpMap[command]

	cw.strBuilder.WriteString("@SP\n")
	cw.strBuilder.WriteString("AM=M-1\n")
	cw.strBuilder.WriteString("D=M\n")
	cw.strBuilder.WriteString("A=A-1\n")
	cw.strBuilder.WriteString("D=M-D\n")
	fmt.Fprintf(cw.strBuilder, "@%s.%s.%d\n", cw.fname, strings.ToUpper(command), cw.numLabels)
	fmt.Fprintf(cw.strBuilder, "D;%s\n", jmp)
	cw.strBuilder.WriteString("@SP\n")
	cw.strBuilder.WriteString("A=M-1\n")
	cw.strBuilder.WriteString("M=0\n") // false
	fmt.Fprintf(cw.strBuilder, "@%s.%s_END.%d\n", cw.fname, strings.ToUpper(command), cw.numLabels)
	cw.strBuilder.WriteString("0;JEQ\n")
	fmt.Fprintf(cw.strBuilder, "(%s.%s.%d)\n", cw.fname, strings.ToUpper(command), cw.numLabels)
	cw.strBuilder.WriteString("@SP\n")
	cw.strBuilder.WriteString("A=M-1\n")
	cw.strBuilder.WriteString("M=-1\n") // true
	fmt.Fprintf(cw.strBuilder, "(%s.%s_END.%d)\n", cw.fname, strings.ToUpper(command), cw.numLabels)
}
