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

func (vw *vmWriter) writeFunction(className string, subroutineName string, nVars int) {
	fmt.Fprintf(vw.outf, "function %s.%s %d\n", className, subroutineName, nVars)
}
