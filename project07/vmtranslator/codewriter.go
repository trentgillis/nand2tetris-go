package vmtranslator

import (
	"fmt"
	"os"
	"strings"
)

type codeWriter struct {
	outfile *os.File
}

func newCodeWriter(outfile *os.File) codeWriter {
	return codeWriter{
		outfile: outfile,
	}
}

func (cw *codeWriter) write(commandType int, arg1 string, arg2 string) {
	var b strings.Builder
	switch commandType {
	case c_push:
		writePush(&b, arg1, arg2)
	case c_pop:
		writePop(&b, arg1, arg2)
	case c_arithmetic:
		writeArithmetic(&b, arg1)
	}
	cw.outfile.WriteString(b.String())
}

func writePush(b *strings.Builder, segment string, index string) {
	switch segment {
	case "constant":
		writePushConstant(b, index)
	}

	b.WriteString("@SP\n")
	b.WriteString("A=M\n")
	b.WriteString("M=D\n")
	b.WriteString("@SP\n")
	b.WriteString("M=M+1\n")
}

func writePushConstant(b *strings.Builder, index string) {
	fmt.Fprintf(b, "@%s\n", index)
	b.WriteString("D=A\n")
}

func writePop(b *strings.Builder, segment string, index string) {
	b.WriteString("@SP\n")
	b.WriteString("M=M-1\n")

	switch segment {
	case "register":
		writePopReg(b, index)
	}
}

func writePopReg(b *strings.Builder, index string) {
	b.WriteString("@SP\n")
	b.WriteString("A=M\n")
	b.WriteString("D=M\n")
	fmt.Fprintf(b, "@R%s\n", index)
	b.WriteString("M=D\n")
}

func writeArithmetic(b *strings.Builder, command string) {
	switch command {
	case "add":
		writeAdd(b)
	}
}

func writeAdd(b *strings.Builder) {
	writePop(b, "register", "13")
	writePop(b, "register", "14")
	b.WriteString("@R13\n")
	b.WriteString("D=M\n")
	b.WriteString("@R14\n")
	b.WriteString("D=D+M\n")
	writePush(b, "", "")
}
