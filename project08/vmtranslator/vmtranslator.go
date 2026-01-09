package vmtranslator

import (
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strings"
)

type vmTranslator struct {
	vmFilePaths []string
	asmFile     *os.File
	codeWriter  codeWriter
	shouldInit  bool
}

func Translate(programPath string) {
	vmt := newVmTranslator(programPath)
	defer vmt.asmFile.Close()

	if vmt.shouldInit {
		vmt.codeWriter.strBuilder.Reset()
		vmt.codeWriter.writeInit()
	}

	for _, fPath := range vmt.vmFilePaths {
		vmt.translateVmFile(fPath)
	}

	// The Sys.init function handles entering an infinite loop after executionon behalf of
	// our program. If it is not present however, add an end of program loop manually.
	if !vmt.shouldInit {
		fname, _ := strings.CutSuffix(filepath.Base(vmt.asmFile.Name()), ".asm")
		fmt.Fprintf(vmt.asmFile, "(%s.END_LOOP)\n@%s.END_LOOP\n0;JEQ\n", fname, fname)
	}
}

func (vmt *vmTranslator) translateVmFile(vmFilePath string) {
	f, err := os.Open(vmFilePath)
	if err != nil {
		log.Fatalf("vmtranslator.translateVmFile: %e\n", err)
	}
	defer f.Close()

	// Sets the filename attr on our codewrite for use in creating unique symbols
	vmFname, _ := strings.CutSuffix(filepath.Base(vmFilePath), ".vm")
	vmt.codeWriter.setCurrFname(vmFname)

	parser := newParser(f)
	parser.Advance()
	for parser.hasMoreLines {
		vmt.codeWriter.write(parser.commandType(), parser.arg1(), parser.arg2())
		parser.Advance()
	}
}

func newVmTranslator(programPath string) vmTranslator {
	var asmFilePath string
	var vmFilePaths []string
	shouldInit := false

	if strings.HasSuffix(programPath, ".vm") {
		vmFilePaths = append(vmFilePaths, programPath)
		asmFilePath = strings.Replace(programPath, ".vm", ".asm", 1)
	} else {
		vmFilePaths = getVmPathsFromDir(programPath)
		for _, vmFilePath := range vmFilePaths {
			if filepath.Base(vmFilePath) == "Sys.vm" {
				shouldInit = true
			}
		}

		asmFilePath = programPath + fmt.Sprintf("/%s.asm", filepath.Base(programPath))
	}

	asmFile, err := os.Create(asmFilePath)
	if err != nil {
		log.Fatalf("vmtranslator.newVmTranslator: %e\n", err)
	}

	codeWriter := newCodeWriter(asmFile)
	return vmTranslator{
		vmFilePaths: vmFilePaths,
		asmFile:     asmFile,
		codeWriter:  codeWriter,
		shouldInit:  shouldInit,
	}
}

func getVmPathsFromDir(dirPath string) []string {
	dirEntries, err := os.ReadDir(dirPath)
	if err != nil {
		log.Fatal(err)
	}

	vmFilePaths := []string{}
	for _, entry := range dirEntries {
		if !entry.IsDir() && strings.HasSuffix(entry.Name(), ".vm") {
			vmFilePaths = append(vmFilePaths, fmt.Sprintf("%s/%s", dirPath, entry.Name()))
		}
	}
	if len(vmFilePaths) == 0 {
		log.Fatal("vmtranslator.getVmPathsFromDir: input directory contains no vm file for translation")
	}

	return vmFilePaths
}
