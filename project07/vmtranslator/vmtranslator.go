package vmtranslator

import (
	"fmt"
	"log"
	"os"
	"strings"
)

type VmTranslator struct {
	vmFile  *os.File
	asmFile *os.File
	parser  Parser
}

func New(f *os.File) *VmTranslator {
	asmFile, err := os.Create(strings.Replace(f.Name(), ".vm", ".asm", 1))
	if err != nil {
		log.Fatalf("vmtranslator.New: %e\n", err)
	}

	parser := NewParser(f)

	return &VmTranslator{
		vmFile:  f,
		asmFile: asmFile,
		parser:  parser,
	}
}

func (vmt *VmTranslator) Translate() {
	vmt.parser.Advance()
	for vmt.parser.HasMoreLines {
		arg1 := vmt.parser.Arg1()
		arg2 := vmt.parser.Arg2()
		commandType := vmt.parser.CommandType()
		fmt.Fprintf(vmt.asmFile, "Arg1(): %s, Arg2(): %s, CommandType(): %d\n", arg1, arg2, commandType)

		vmt.parser.Advance()
	}
}
