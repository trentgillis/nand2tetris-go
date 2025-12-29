package main

import (
	"jackvmt/vmtranslator"
	"log"
	"os"
	"path/filepath"
)

func main() {
	if len(os.Args) < 2 {
		log.Fatal("Filename was not provided")
	}

	filename := os.Args[1]
	if filepath.Ext(filename) != ".vm" {
		log.Fatal("Input file have the .asm extension")
	}

	f, err := os.Open(filename)
	if err != nil {
		log.Fatal(err)
	}
	defer f.Close()

	vmtranslator.New(f).Translate()
}
