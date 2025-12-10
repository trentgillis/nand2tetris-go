package assembler

import (
	"fmt"
	"log"
	"os"
	"strconv"
	"strings"
)

type Assembler struct {
	outfile *os.File
	parser  Parser
	codegen CodeGen
	st      SymbolTable
}

func New(f *os.File) Assembler {
	p := NewParser(f)
	c := CodeGen{}
	st := NewSymbolTable()

	outfile, err := os.Create(strings.Replace(f.Name(), ".asm", ".hack", 1))
	if err != nil {
		log.Fatal(err)
	}

	return Assembler{
		outfile: outfile,
		parser:  p,
		codegen: c,
		st:      st,
	}
}

func (a Assembler) Assemble() {
	a.parser.Advance()
	for a.parser.HasMoreLines {
		switch a.parser.CurrInstType() {
		case A_INSTRUCTION:
			a.writeAInst()
		case C_INSTRUCTION:
			a.writeCInst()
		case L_INSTRUCTION:
			a.writeLInst()
		}

		a.parser.Advance()
	}
}

func (a Assembler) writeAInst() {
	symbol := a.parser.Symbol()

	var value int64
	value, err := strconv.ParseInt(symbol, 10, 16)
	if err != nil {
		symbolAddr := a.st.GetAddr(symbol)
		binStr := padZeros(fmt.Sprintf("%b", symbolAddr))
		fmt.Fprintf(a.outfile, "%s\n", binStr)
		return
	}

	binStr := padZeros(fmt.Sprintf("%b", value))
	fmt.Fprintf(a.outfile, "%s\n", binStr)
}

func (a Assembler) writeCInst() {
	dest := a.parser.Dest()
	comp := a.parser.Comp()
	jump := a.parser.Jump()
	aBit := "0"
	if strings.Contains(comp, "M") {
		aBit = "1"
	}

	fmt.Fprintf(a.outfile, "111%s%s%s%s\n", aBit, a.codegen.Comp(comp), a.codegen.Dest(dest), a.codegen.Jump(jump))
}

func (a Assembler) writeLInst() {
	symbol := a.parser.Symbol()
	if !a.st.Contains(symbol) {
		a.st.AddEntry(symbol)
	}
}

func padZeros(binStr string) string {
	padAmt := 16 - len(binStr)
	return strings.Repeat("0", padAmt) + binStr
}
