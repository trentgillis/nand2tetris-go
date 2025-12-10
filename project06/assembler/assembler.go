package assembler

import (
	"fmt"
	"log"
	"os"
	"strconv"
	"strings"
)

type Assembler struct {
	infile  *os.File
	outfile *os.File
	codegen CodeGen
	st      SymbolTable
}

func New(f *os.File) Assembler {
	c := CodeGen{}
	st := NewSymbolTable()

	outfile, err := os.Create(strings.Replace(f.Name(), ".asm", ".hack", 1))
	if err != nil {
		log.Fatal(err)
	}

	return Assembler{
		infile:  f,
		outfile: outfile,
		codegen: c,
		st:      st,
	}
}

func (a Assembler) Assemble() {
	a.GenerateLAddrs()

	parser := NewParser(a.infile)
	parser.Advance()
	for parser.HasMoreLines {
		switch parser.CurrInstType() {
		case A_INSTRUCTION:
			symbol := parser.Symbol()
			a.writeAInst(symbol)
		case C_INSTRUCTION:
			dest := parser.Dest()
			comp := parser.Comp()
			jump := parser.Jump()
			a.writeCInst(dest, comp, jump)
		}

		parser.Advance()
	}
}

func (a Assembler) GenerateLAddrs() {
	lineNum := 0
	parser := NewParser(a.infile)

	parser.Advance()
	for parser.HasMoreLines {
		switch parser.CurrInstType() {
		case L_INSTRUCTION:
			symbol := parser.Symbol()
			fmt.Println(symbol)
			fmt.Println(lineNum)
			a.processLInstruction(symbol, lineNum)
		case A_INSTRUCTION:
			lineNum += 1
		case C_INSTRUCTION:
			lineNum += 1
		}

		parser.Advance()
	}

	a.infile.Seek(0, 0)
	fmt.Println(a.st)
}

func (a Assembler) writeAInst(symbol string) {
	var value int64
	value, err := strconv.ParseInt(symbol, 10, 16)
	if err != nil {
		if !a.st.Contains(symbol) {
			a.st.AddVar(symbol)
		}
		symbolAddr := a.st.GetAddr(symbol)
		binStr := padZeros(fmt.Sprintf("%b", symbolAddr))
		fmt.Fprintf(a.outfile, "%s\n", binStr)
		return
	}

	binStr := padZeros(fmt.Sprintf("%b", value))
	fmt.Fprintf(a.outfile, "%s\n", binStr)
}

func (a Assembler) writeCInst(dest string, comp string, jump string) {
	aBit := "0"
	if strings.Contains(comp, "M") {
		aBit = "1"
	}
	fmt.Fprintf(a.outfile, "111%s%s%s%s\n", aBit, a.codegen.Comp(comp), a.codegen.Dest(dest), a.codegen.Jump(jump))
}

func (a Assembler) processLInstruction(symbol string, lineNum int) {
	if a.st.Contains(symbol) {
		return
	}
	a.st.AddEntry(symbol, lineNum)
}

func padZeros(binStr string) string {
	padAmt := 16 - len(binStr)
	return strings.Repeat("0", padAmt) + binStr
}
