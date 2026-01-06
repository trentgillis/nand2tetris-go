package main

import (
	"jackvmt/vmtranslator"
	"log"
	"os"
)

func main() {
	if len(os.Args) < 2 {
		log.Fatal("Path to vm file or directory for translation was not provided")
	}
	programPath := os.Args[1]
	vmtranslator.Translate(programPath)
}
