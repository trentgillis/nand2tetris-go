package vmtranslator

import (
	"fmt"
	"log"
	"os"
	"strings"
)

type vmTranslator struct {
	vmFile     *os.File
	asmFile    *os.File
	parser     Parser
	codeWriter codeWriter
}

func Translate(f *os.File) {
	vmt := newVmTranslator(f)

	vmt.parser.Advance()
	for vmt.parser.HasMoreLines {
		// TODO: remove when complete; used to debug generated asm in asm output files
		fmt.Fprintf(vmt.asmFile, "// %s\n", vmt.parser.currLine)

		vmt.codeWriter.write(vmt.parser.CommandType(), vmt.parser.Arg1(), vmt.parser.Arg2())
		vmt.parser.Advance()
	}
}

func newVmTranslator(f *os.File) vmTranslator {
	asmFile, err := os.Create(strings.Replace(f.Name(), ".vm", ".asm", 1))
	if err != nil {
		log.Fatalf("vmtranslator.New: %e\n", err)
	}

	parser := NewParser(f)
	codeWriter := newCodeWriter(asmFile)

	return vmTranslator{
		vmFile:     f,
		asmFile:    asmFile,
		parser:     parser,
		codeWriter: codeWriter,
	}
}
