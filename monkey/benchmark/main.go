package main

import (
	"flag"
	"fmt"
	"monkey/compiler"
	"monkey/eval"
	"monkey/lexer"
	"monkey/object"
	"monkey/parser"
	"monkey/vm"
	"time"
)

var engine = flag.String("engine", "vm", "use 'vm' or 'eval'")

var input = `
let fibonacci = fn(x) {
	if (x == 0) {
		return 0;
	}
	if (x == 1) {
		return 1;
	}
	return fibonacci(x - 1) + fibonacci(x - 2);
};
fibonacci(35);
`

func main() {
	flag.Parse()
	var duration time.Duration
	var result object.Object

	l := lexer.New(input)
	p := parser.New(l)
	program := p.ParseProgram()

	if *engine == "vm" {
		comp := compiler.New()
		err := comp.Compile(program)
		if err != nil {
			fmt.Printf("compiler error: %s", err)
			return
		}
		machine := vm.New(comp.Bytecode())
		start := time.Now()
		err = machine.Run()
		if err != nil {
			fmt.Printf("vm error: %s", err)
			return
		}
		duration = time.Since(start)
		result = machine.LastPoppedStackElement()
	} else {
		env := object.NewEnv()
		start := time.Now()
		result = eval.Eval(program, env)
		duration = time.Since(start)
	}

	fmt.Printf(
		"engine: %s, result: %s, duration: %s\n",
		*engine,
		result.Inspect(),
		duration,
	)
}
