package vmtranslator

import (
	"fmt"
	"log"
	"os"
	"path/filepath"
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
	defer vmt.asmFile.Close()

	vmt.parser.Advance()
	for vmt.parser.hasMoreLines {
		vmt.codeWriter.write(vmt.parser.commandType(), vmt.parser.arg1(), vmt.parser.arg2())
		vmt.parser.Advance()
	}

	// Write infinite loop to the end of the program
	fname, _ := strings.CutSuffix(filepath.Base(vmt.asmFile.Name()), ".asm")
	fmt.Fprintf(vmt.asmFile, "(%s.END_LOOP)\n@%s.END_LOOP\n0;JEQ\n", fname, fname)
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
