package main

import (
	"fmt"
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
	// repl.Start(os.Stdin, os.Stdout)
	repl.StartCompiled(os.Stdin, os.Stdout)
}

func runFile(filename string) {
	fContent, err := os.ReadFile(filename)
	if err != nil {
		fmt.Printf("Error opening file %s\n", filename)
		os.Exit(1)
	}
	input := string(fContent)
	macroEnv := object.NewEnv()

	lexer := lexer.New(input)
	parser := parser.New(lexer)
	program := parser.ParseProgram()

	if len(parser.Errors()) > 0 {
		fmt.Println("Error parsing program")
		for _, msg := range parser.Errors() {
			fmt.Println(msg)
		}
		return
	}
	eval.DefineMacros(program, macroEnv)
	expanded := eval.ExpandMacros(program, macroEnv)
	eval.Eval(expanded, object.NewEnv())
}
