package main

import (
	"bufio"
	"fmt"
	"os"
	"path/filepath"
	"slices"
	"sort"
	"strings"
)

// Setting to skip GoLang packages for a simpler diagram
var skipGoPackages = true

var imports = map[string][]string{}

func main() {
	if err := searchDirectory("./cmd/"); err != nil {
		fmt.Println(err)
		return
	}

	// Get and sort packages
	packages := []string{}
	for category := range imports {
		packages = append(packages, category)
	}
	sort.Slice(packages, func(i, j int) bool {
		return packages[i] < packages[j]
	})

	writeToFile("./doc/Packages.md", generateMermaid(packages))
}

func writeToFile(path, content string) error {
	file, err := os.Create(path)
	if err != nil {
		return fmt.Errorf("failed to create file %s: %s", path, err)
	}

	_, err = file.WriteString(content)
	if err != nil {
		return fmt.Errorf("failed to write to file %s: %s", path, err)
	}
	return nil
}

func generateMermaid(packages []string) string {
	var markdown strings.Builder

	markdown.WriteString("# Packages\n\n")
	markdown.WriteString("```mermaid\n")
	markdown.WriteString("flowchart LR\n")

	for _, pack := range packages {

		packageImports := imports[pack]
		sort.Slice(packageImports, func(i, j int) bool {
			return packageImports[i] < packageImports[j]
		})

		for _, packageImport := range packageImports {
			markdown.WriteString("    ")
			markdown.WriteString(strings.ReplaceAll(pack, "/", "_"))
			markdown.WriteString("[")
			markdown.WriteString(pack)
			markdown.WriteString("] --> ")
			markdown.WriteString(strings.ReplaceAll(packageImport, "/", "_"))
			markdown.WriteString("[")
			markdown.WriteString(packageImport)
			markdown.WriteString("]\n")
		}
		markdown.WriteString("\n")
	}

	markdown.WriteString("```")

	return markdown.String()
}

func searchDirectory(path string) error {
	files, err := os.ReadDir(path)
	if err != nil {
		return fmt.Errorf("failed reading dir %s: %s", path, err)
	}

	for _, file := range files {
		// Iterate sub directories
		if file.IsDir() {
			err := searchDirectory(filepath.Join(path, file.Name()))
			if err != nil {
				return err
			}
			continue
		}

		if !strings.HasSuffix(file.Name(), ".go") {
			continue
		}
		// Read all .go files
		err = readGoFile(path, file.Name())
		if err != nil {
			return err
		}
	}
	return nil
}

func readGoFile(path, filename string) error {
	file, err := os.Open(filepath.Join(path, filename))
	if err != nil {
		return fmt.Errorf("failed to open file %s: %s", path, err)
	}
	defer file.Close()

	isImportSection := false
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line == "" {
			continue
		}

		if isImportSection && line == ")" {
			return nil
		}

		if isImportSection {
			if !strings.HasPrefix(line, "\"") {
				line = line[strings.Index(line, "\""):]
			}
			line = strings.TrimPrefix(line, "\"")
			addImport(path, line[:strings.Index(line, "\"")])
			continue
		}

		if strings.HasPrefix(line, "import (") {
			isImportSection = true
			continue
		}

		if strings.HasPrefix(line, "import \"") {
			line = strings.Replace(line, "import \"", "", 1)
			addImport(path, line[:strings.Index(line, "\"")])
			return nil
		}
	}
	return nil
}

func addImport(curPackage, usedPackage string) {
	if skipGoPackages && !strings.HasPrefix(usedPackage, "GoPHP/cmd/") {
		return
	}

	if strings.HasPrefix(curPackage, "cmd/") {
		curPackage = strings.Replace(curPackage, "cmd/", "GoPHP/cmd/", 1)
	}

	if _, found := imports[curPackage]; !found {
		imports[curPackage] = []string{}
	} else {
		if slices.Contains(imports[curPackage], usedPackage) {
			return
		}
	}
	imports[curPackage] = append(imports[curPackage], usedPackage)

}
