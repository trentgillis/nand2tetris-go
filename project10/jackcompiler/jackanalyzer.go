package jackcompiler

import (
	"fmt"
	"log"
	"os"
	"path"
	"strings"
)

type jackAnalyzer struct {
	jackFilePaths []string
}

func Analyze(programPath string) {
	jackPaths := getJackPaths(programPath)

	for _, path := range jackPaths {
		analyzeJackFile(path)
	}
}

func analyzeJackFile(jackPath string) {
	f, err := os.Open(jackPath)
	if err != nil {
		log.Fatal(err)
	}
	defer f.Close()

	dir, fileName := path.Split(jackPath)
	outfPath := fmt.Sprintf("%s%s%s", dir, "output/", strings.Replace(fileName, ".jack", "T.xml", 1))
	os.MkdirAll(path.Dir(outfPath), 0755)
	outf, err := os.Create(outfPath)
	if err != nil {
		log.Fatal(err)
	}

	fmt.Fprint(outf, "<tokens>\n")
	jt := newJackTokenizer(f, outf)
	jt.advance()
	for jt.hasMoreTokens {
		jt.printTokenXML()
		jt.advance()
	}
	fmt.Fprint(outf, "</tokens>\n")
}

func getJackPaths(programPath string) []string {
	jackPaths := []string{}

	info, err := os.Stat(programPath)
	if err != nil {
		log.Fatal(err)
	}

	if info.IsDir() {
		jackPaths = getJackPathsFromDir(programPath)
	} else if strings.HasSuffix(programPath, ".jack") {
		jackPaths = append(jackPaths, programPath)
	}

	return jackPaths
}

func getJackPathsFromDir(dirPath string) []string {
	entries, err := os.ReadDir(dirPath)
	if err != nil {
		log.Fatal(err)
	}

	jackFilePaths := []string{}
	for _, entry := range entries {
		if !entry.IsDir() && strings.HasSuffix(entry.Name(), ".jack") {
			jackFilePaths = append(jackFilePaths, fmt.Sprintf("%s/%s", dirPath, entry.Name()))
		}
	}
	if len(jackFilePaths) == 0 {
		log.Fatal("No .jack files in the input directory")
	}

	return jackFilePaths
}
