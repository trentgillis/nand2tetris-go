package main

import (
	"jackc/jackcompiler"
	"log"
	"os"
)

func main() {
	if len(os.Args) < 2 {
		log.Fatal("Path to jack file or program directory was not provided")
	}
	programPath := os.Args[1]
	jackcompiler.Analyze(programPath)
}
