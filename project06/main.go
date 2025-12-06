package main

import (
	"hackassembler/assembler"
	"log"
	"os"
	"path/filepath"
)

func main() {
	if len(os.Args) < 2 {
		log.Fatal("Filename was not provided")
	}

	filename := os.Args[1]
	if filepath.Ext(filename) != ".asm" {
		log.Fatal("Input file have the .asm extension")
	}

	f, err := os.Open(filename)
	if err != nil {
		log.Fatal(err)
	}
	defer f.Close()

	a := assembler.New(f)
	a.Assemble()
}
