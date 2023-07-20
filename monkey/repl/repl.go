package repl

import (
	"bufio"
	"fmt"
	"io"
	"monkey/compiler"
	"monkey/eval"
	"monkey/lexer"
	"monkey/object"
	"monkey/parser"
	"monkey/vm"
)

const PROMPT = "monkey > "
const MONKEY_FACE = `            __,__
   .--.  .-"     "-.  .--.
  / .. \/  .-. .-.  \/ .. \
 | |  '|  /   Y   \  |'  | |
 | \   \  \ 0 | 0 /  /   / |
  \ '- ,\.-"""""""-./, -' /
   ''-' /_   ^ ^   _\ '-''
       |  \._   _./  |
       \   \ '~' /   /
        '._ '-=-' _.'
           '-----'
`

func Start(in io.Reader, out io.Writer) {
	scanner := bufio.NewScanner(in)
	env := object.NewEnv()
	macroEnv := object.NewEnv()

	for {
		fmt.Fprint(out, PROMPT)
		scanned := scanner.Scan()
		if !scanned {
			return
		}

		line := scanner.Text()
		lexer := lexer.New(line)
		parser := parser.New(lexer)
		program := parser.ParseProgram()

		if len(parser.Errors()) > 0 {
			printParserErros(out, parser.Errors())
			continue
		}

		eval.DefineMacros(program, macroEnv)
		expanded := eval.ExpandMacros(program, macroEnv)
		evaluated := eval.Eval(expanded, env)
		if evaluated != nil {
			io.WriteString(out, evaluated.Inspect())
			io.WriteString(out, "\n")
		}
	}
}

func StartCompiled(in io.Reader, out io.Writer) {
	scanner := bufio.NewScanner(in)
	constants := []object.Object{}
	globals := make([]object.Object, vm.GlobalsSize)
	symTable := compiler.NewSymbolTable()
	for i, v := range object.Builtins {
		symTable.DefineBuiltin(i, v.Name)
	}

	for {
		fmt.Fprint(out, PROMPT)
		scanned := scanner.Scan()
		if !scanned {
			return
		}

		line := scanner.Text()
		lexer := lexer.New(line)
		parser := parser.New(lexer)
		program := parser.ParseProgram()

		if len(parser.Errors()) > 0 {
			printParserErros(out, parser.Errors())
			continue
		}

		compiler := compiler.NewWithState(symTable, constants)
		err := compiler.Compile(program)
		if err != nil {
			fmt.Fprintf(out, "Woops Compilation failed:\n%s\n", err)
			continue
		}
		code := compiler.Bytecode()
		constants = code.Constants

		machine := vm.NewWithState(compiler.Bytecode(), globals)
		err = machine.Run()
		if err != nil {
			fmt.Fprintf(out, "Woops Executing bytecode failed:\n%s\n", err)
			continue
		}
		stackTop := machine.LastPoppedStackElement()
		io.WriteString(out, stackTop.Inspect())
		io.WriteString(out, "\n")
	}
}

func printParserErros(out io.Writer, errors []string) {
	io.WriteString(out, MONKEY_FACE)
	io.WriteString(out, "Whoops!\nparser errors:\n")
	for _, message := range errors {
		io.WriteString(out, "\t"+message+"\n")
	}
}
