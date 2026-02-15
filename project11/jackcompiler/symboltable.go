package jackcompiler

type stEntry struct {
	name     string
	dataType string
	kind     string
	index    int
}

type symbolTable struct {
	table       map[string]stEntry
	staticCount int
	fieldCount  int
	argCount    int
	localCount  int
}

// type symbolTable map[string]stEntry

func newSymbolTable() symbolTable {
	return symbolTable{
		table:       make(map[string]stEntry),
		staticCount: 0,
		fieldCount:  0,
		argCount:    0,
		localCount:  0,
	}
}
