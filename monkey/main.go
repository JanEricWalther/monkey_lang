package main

import (
	"fmt"
	"io"
	"monkey/eval"
	"monkey/lexer"
	"monkey/object"
	"monkey/parser"
	"monkey/repl"
	"os"
)

func main() {
	if len(os.Args) > 1 {
		runFile(os.Args[1])
	} else {
		startRepl()
	}
}

func startRepl() {
	fmt.Println("Hello! This is the Monkey programming language!")
	fmt.Println("Feel free to type in commands")
	repl.Start(os.Stdin, os.Stdout)
}

func runFile(filename string) {
	fContent, err := os.ReadFile(filename)
	if err != nil {
		fmt.Printf("Error opening file %s\n", filename)
		os.Exit(1)
	}
	input := string(fContent)

	lexer := lexer.New(input)
	parser := parser.New(lexer)
	program := parser.ParseProgram()

	if len(parser.Errors()) > 0 {
		// printParserErros(os.Stderr, parser.Errors())
		return
	}
	evaluated := eval.Eval(program, object.NewEnv())
	if evaluated != nil {
		io.WriteString(os.Stdout, evaluated.Inspect())
		io.WriteString(os.Stdout, "\n")
	}
}
