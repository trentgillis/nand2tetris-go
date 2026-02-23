package assembler

import (
	"fmt"
	"log"
	"os"
	"strconv"
	"strings"
)

type Assembler struct {
	infile   *os.File
	outfile  *os.File
	codegen  CodeGen
	symtable SymbolTable
}

func newAssembler(f *os.File) *Assembler {
	codegen := CodeGen{}
	symtable := newSymbolTable()

	outfile, err := os.Create(strings.Replace(f.Name(), ".asm", ".hack", 1))
	if err != nil {
		log.Fatal(err)
	}

	return &Assembler{
		infile:   f,
		outfile:  outfile,
		codegen:  codegen,
		symtable: symtable,
	}
}

func Assemble(f *os.File) {
	a := newAssembler(f)

	// Performs first pass of the input file, adding L instruction symbols to the
	// symbol table
	a.populateLAddrs()

	parser := newParser(a.infile)
	parser.Advance()
	for parser.HasMoreLines {
		switch parser.currInstType() {
		case A_INSTRUCTION:
			symbol := parser.symbol()
			a.processAInst(symbol)
		case C_INSTRUCTION:
			dest := parser.dest()
			comp := parser.comp()
			jump := parser.jump()
			a.processCInst(dest, comp, jump)
		}
		parser.Advance()
	}
}

func (a *Assembler) populateLAddrs() {
	lineNum := 0
	parser := newParser(a.infile)

	parser.Advance()
	for parser.HasMoreLines {
		switch parser.currInstType() {
		case L_INSTRUCTION:
			symbol := parser.symbol()
			a.processLInst(symbol, lineNum)
		case A_INSTRUCTION:
			lineNum += 1
		case C_INSTRUCTION:
			lineNum += 1
		}
		parser.Advance()
	}
	// Reset file position for second pass done by Assemble()
	a.infile.Seek(0, 0)
}

func (a *Assembler) processAInst(symbol string) {
	value, err := strconv.ParseInt(symbol, 10, 16)
	if err != nil {
		if !a.symtable.contains(symbol) {
			a.symtable.addVar(symbol)
		}
		value = int64(a.symtable.getAddr(symbol))
	}

	binStr := padZeros(fmt.Sprintf("%b", value))
	fmt.Fprintf(a.outfile, "%s\n", binStr)
}

func (a *Assembler) processCInst(dest string, comp string, jump string) {
	aBit := "0"
	if strings.Contains(comp, "M") {
		aBit = "1"
	}
	fmt.Fprintf(a.outfile, "111%s%s%s%s\n", aBit, a.codegen.comp(comp), a.codegen.dest(dest), a.codegen.jump(jump))
}

func (a *Assembler) processLInst(symbol string, lineNum int) {
	if a.symtable.contains(symbol) {
		return
	}
	a.symtable.addEntry(symbol, lineNum)
}

func padZeros(binStr string) string {
	padAmt := 16 - len(binStr)
	return strings.Repeat("0", padAmt) + binStr
}
