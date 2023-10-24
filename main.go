package main

import (
	"flag"
	"fmt"
	"os"
	"os/exec"
	"path"

	"github.com/gmdvv2003/brainfuck-compiler-plus/compiler"
	"github.com/gmdvv2003/brainfuck-compiler-plus/lexer"
	"github.com/gmdvv2003/brainfuck-compiler-plus/parser"
)

var (
	filePath       string
	fileOutputName string
	debug          bool
)

func main() {
	flag.StringVar(&filePath, "file", "", "Path to the file to be compiled")
	flag.StringVar(&fileOutputName, "name", "output", "Output file name")
	flag.BoolVar(&debug, "debug", false, "Prints some debug information about the parser")

	flag.Parse()

	if filePath == "" {
		flag.Usage()
		os.Exit(1)
	}

	// Make sure that the entered file exists
	if _, osStatError := os.Stat(fmt.Sprintf("./%s", filePath)); osStatError != nil {
		if os.IsNotExist(osStatError) {
			panic("file does not exist")
		} else {
			panic(fmt.Sprintf("error while trying to open file: %s", osStatError.Error()))
		}
	}

	if path.Ext(filePath) != ".bf" {
		panic("invalid file extension. Must enter .bf file")
	}

	fileContent, readFileError := os.ReadFile(filePath)
	if readFileError != nil {
		panic(fmt.Sprintf("error while trying to read file: %s", readFileError.Error()))
	}

	outputFile, openFileError := os.OpenFile(fmt.Sprintf("%s.asm", fileOutputName), os.O_CREATE|os.O_WRONLY, 0644)
	if openFileError != nil {
		panic(fmt.Sprintf("error while trying to open file: %s", openFileError.Error()))
	}

	// We don't care about possible errors since it must get called anyway
	defer outputFile.Close()

	// Truncate the file to remove any previous content
	if ok := outputFile.Truncate(0); ok != nil {
		panic(fmt.Sprintf("error while trying to truncate file: %s", ok.Error()))
	}

	if ast, _, ok := parser.Parse(lexer.NewLexer(string(fileContent), &debug), nil); ok == nil {
		compiledSource, compileError := compiler.Compile(ast)
		if compileError != nil {
			panic(fmt.Sprintf("error while trying to compile source: %s", compileError.Error()))
		}

		outputFile.Write(compiledSource)

		// Execute NASM over the compiled source, generating an object file, and then linking it to the executable
		if ok := exec.Command("nasm", "-felf64", fmt.Sprintf("%s.asm", fileOutputName)).Run(); ok != nil {
			panic(fmt.Sprintf("failed to compile the provided .asm code.\nError: %s", ok.Error()))
		}

		if ok := exec.Command("ld", fmt.Sprintf("%s.o", fileOutputName), "-o", fileOutputName).Run(); ok != nil {
			panic(fmt.Sprintf("failed to link the provided .o code.\nError: %s", ok.Error()))
		}
	} else {
		panic(fmt.Sprintf("error while trying to parse source: %s", ok.Error()))
	}
}
