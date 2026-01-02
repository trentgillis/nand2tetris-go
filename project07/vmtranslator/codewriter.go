package vmtranslator

import (
	"fmt"
	"os"
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
	fmt.Fprintf(cw.outfile, "ctype: %d, arg1: %s, arg2: %s\n", commandType, arg1, arg2)
}

func writeArithmetic() {}

func writePushPop() {}

func writePush() {}
