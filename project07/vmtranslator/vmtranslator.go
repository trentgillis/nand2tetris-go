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
	parser     parser
	codeWriter codeWriter
}

func Translate(f *os.File) {
	vmt := newVmTranslator(f)

	vmt.parser.Advance()
	for vmt.parser.HasMoreLines {
		// TODO: remove when complete; used to debug generated asm in asm output files
		fmt.Fprintf(vmt.asmFile, "// %s\n", vmt.parser.currLine)

		vmt.codeWriter.write(vmt.parser.commandType(), vmt.parser.arg1(), vmt.parser.arg2())
		vmt.parser.Advance()
	}
}

func newVmTranslator(f *os.File) vmTranslator {
	asmFile, err := os.Create(strings.Replace(f.Name(), ".vm", ".asm", 1))
	if err != nil {
		log.Fatalf("vmtranslator.New: %e\n", err)
	}

	parser := newParser(f)
	codeWriter := newCodeWriter(asmFile)

	return vmTranslator{
		vmFile:     f,
		asmFile:    asmFile,
		parser:     parser,
		codeWriter: codeWriter,
	}
}
