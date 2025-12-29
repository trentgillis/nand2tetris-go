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
}

func New(f *os.File) *VmTranslator {
	asmFile, err := os.Create(strings.Replace(f.Name(), ".vm", ".asm", 1))
	if err != nil {
		log.Fatalf("vmtranslator.New: %e\n", err)
	}

	return &VmTranslator{
		vmFile:  f,
		asmFile: asmFile,
	}
}

func (vmt *VmTranslator) Translate() {
	fmt.Fprint(vmt.asmFile, "Translating...")
}
