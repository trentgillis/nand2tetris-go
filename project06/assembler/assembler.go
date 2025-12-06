package assembler

import (
	"log"
	"os"
	"strings"
)

type Assembler struct {
	outfile *os.File
	parser  Parser
}

func New(f *os.File) Assembler {
	p := NewParser(f)

	outfile, err := os.Create(strings.Replace(f.Name(), ".asm", ".hack", 1))
	if err != nil {
		log.Fatal(err)
	}

	return Assembler{
		outfile: outfile,
		parser:  p,
	}
}

func (a Assembler) Assemble() {
	for a.parser.HasMoreLines {
		a.parser.Advance()
		switch a.parser.CurrentInstructionType() {
		case A_INSTRUCTION:
			a.writeAInst()
		case C_INSTRUCTION:
			a.writeCInst()
		case L_INSTRUCTION:
			a.writeLInst()
		}
	}
}

func (a Assembler) writeAInst() {
	a.outfile.WriteString("TODO\n")
}

func (a Assembler) writeCInst() {
	a.outfile.WriteString("TODO\n")
}

func (a Assembler) writeLInst() {
	a.outfile.WriteString("TODO\n")
}
