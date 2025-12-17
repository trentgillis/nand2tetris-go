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

func New(f *os.File) *Assembler {
	codegen := CodeGen{}
	symtable := NewSymbolTable()

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

func (a *Assembler) Assemble() {
	// Performs first pass of the input file, adding L instruction symbols to the
	// symbol table
	a.populateLAddrs()

	parser := NewParser(a.infile)
	parser.Advance()
	for parser.HasMoreLines {
		switch parser.CurrInstType() {
		case A_INSTRUCTION:
			symbol := parser.Symbol()
			a.processAInst(symbol)
		case C_INSTRUCTION:
			dest := parser.Dest()
			comp := parser.Comp()
			jump := parser.Jump()
			a.processCInst(dest, comp, jump)
		}
		parser.Advance()
	}
}

func (a *Assembler) populateLAddrs() {
	lineNum := 0
	parser := NewParser(a.infile)

	parser.Advance()
	for parser.HasMoreLines {
		switch parser.CurrInstType() {
		case L_INSTRUCTION:
			symbol := parser.Symbol()
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
		if !a.symtable.Contains(symbol) {
			a.symtable.AddVar(symbol)
		}
		value = int64(a.symtable.GetAddr(symbol))
	}

	binStr := padZeros(fmt.Sprintf("%b", value))
	fmt.Fprintf(a.outfile, "%s\n", binStr)
}

func (a *Assembler) processCInst(dest string, comp string, jump string) {
	aBit := "0"
	if strings.Contains(comp, "M") {
		aBit = "1"
	}
	fmt.Fprintf(a.outfile, "111%s%s%s%s\n", aBit, a.codegen.Comp(comp), a.codegen.Dest(dest), a.codegen.Jump(jump))
}

func (a *Assembler) processLInst(symbol string, lineNum int) {
	if a.symtable.Contains(symbol) {
		return
	}
	a.symtable.AddEntry(symbol, lineNum)
}

func padZeros(binStr string) string {
	padAmt := 16 - len(binStr)
	return strings.Repeat("0", padAmt) + binStr
}
