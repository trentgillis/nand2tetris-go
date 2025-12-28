package assembler

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestAssembler(t *testing.T) {
	testCases := []struct {
		name        string
		asmFilePath string
		cmpFilePath string
	}{
		{
			name:        "Add",
			asmFilePath: "../asm/add/Add.asm",
			cmpFilePath: "../asm/add/AddCmp.hack",
		},
		{
			name:        "Max",
			asmFilePath: "../asm/max/Max.asm",
			cmpFilePath: "../asm/max/MaxCmp.hack",
		},
		{
			name:        "MaxL",
			asmFilePath: "../asm/max/MaxL.asm",
			cmpFilePath: "../asm/max/MaxLCmp.hack",
		},
		{
			name:        "Rect",
			asmFilePath: "../asm/rect/Rect.asm",
			cmpFilePath: "../asm/rect/RectCmp.hack",
		},
		{
			name:        "RectL",
			asmFilePath: "../asm/rect/RectL.asm",
			cmpFilePath: "../asm/rect/RectLCmp.hack",
		},
		{
			name:        "Pong",
			asmFilePath: "../asm/pong/Pong.asm",
			cmpFilePath: "../asm/pong/PongCmp.hack",
		},
		{
			name:        "PongL",
			asmFilePath: "../asm/pong/PongL.asm",
			cmpFilePath: "../asm/pong/PongLCmp.hack",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			inputFile, err := os.Open(tc.asmFilePath)
			if err != nil {
				t.Fatalf("Failed to open input file %s: %v", tc.asmFilePath, err)
			}
			defer inputFile.Close()

			assembler := New(inputFile)
			assembler.Assemble()
			assembler.outfile.Close()

			outputPath := strings.Replace(tc.asmFilePath, ".asm", ".hack", 1)
			generatedContent, err := os.ReadFile(outputPath)
			if err != nil {
				t.Fatalf("Failed to read generated output file %s: %v", outputPath, err)
			}
			expectedContent, err := os.ReadFile(tc.cmpFilePath)
			if err != nil {
				t.Fatalf("Failed to read reference file %s: %v", tc.cmpFilePath, err)
			}

			if string(generatedContent) != string(expectedContent) {
				t.Errorf("Output mismatch for %s:\nGenerated file: %s\nExpected file: %s\n\nGenerated:\n%s\n\nExpected:\n%s",
					tc.name,
					outputPath,
					tc.cmpFilePath,
					string(generatedContent),
					string(expectedContent))
			}
		})
	}
}

func TestAssemblerCleanup(t *testing.T) {
	tmpDir := t.TempDir()
	tmpAsmPath := filepath.Join(tmpDir, "test.asm")
	tmpHackPath := filepath.Join(tmpDir, "test.hack")
	testProgram := `// Simple test program
					@2
					D=A
					@3
					D=D+A
					@0
					M=D
					`
	if err := os.WriteFile(tmpAsmPath, []byte(testProgram), 0644); err != nil {
		t.Fatalf("Failed to create temp test file: %v", err)
	}

	inputFile, err := os.Open(tmpAsmPath)
	if err != nil {
		t.Fatalf("Failed to open temp input file: %v", err)
	}

	assembler := New(inputFile)
	assembler.Assemble()
	inputFile.Close()
	assembler.outfile.Close()

	if _, err := os.Stat(tmpHackPath); os.IsNotExist(err) {
		t.Errorf("Output file was not created: %s", tmpHackPath)
	}
	content, err := os.ReadFile(tmpHackPath)
	if err != nil {
		t.Fatalf("Failed to read generated file: %v", err)
	}
	if len(content) == 0 {
		t.Error("Generated file is empty")
	}
	lines := strings.Split(strings.TrimSpace(string(content)), "\n")
	if len(lines) != 6 {
		t.Errorf("Expected 6 instructions, got %d", len(lines))
	}
}
