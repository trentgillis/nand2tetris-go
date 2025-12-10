package assembler

type SymbolTable struct {
	currAddr int
	symbols  map[string]int
}

func NewSymbolTable() SymbolTable {
	return SymbolTable{
		currAddr: 16,
		symbols: map[string]int{
			"R0":     0,
			"R1":     1,
			"R2":     2,
			"R3":     3,
			"R4":     4,
			"R5":     5,
			"R6":     6,
			"R7":     7,
			"R8":     8,
			"R9":     9,
			"R10":    10,
			"R11":    11,
			"R12":    12,
			"R13":    13,
			"R14":    14,
			"R15":    15,
			"SP":     0,
			"LCL":    1,
			"ARG":    2,
			"THIS":   3,
			"THAT":   4,
			"SCREEN": 16384,
			"KBD":    24576,
		},
	}
}

func (st *SymbolTable) GetAddr(symbol string) int {
	addr, ok := st.symbols[symbol]
	if !ok {
		addr = st.AddEntry(symbol)
	}

	return addr
}

func (st *SymbolTable) Contains(symbol string) bool {
	_, ok := st.symbols[symbol]
	return ok
}

func (st *SymbolTable) AddEntry(symbol string) int {
	st.symbols[symbol] = st.currAddr
	st.currAddr += 1
	return st.currAddr
}
