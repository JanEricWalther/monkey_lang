package main

import (
	"fmt"
	"monkey/repl"
	"os"
)

func main() {
	fmt.Println("Hello! This is the Monkey programming language!")
	fmt.Println("Feel free to type in commands")
	repl.Start(os.Stdin, os.Stdout)
}
