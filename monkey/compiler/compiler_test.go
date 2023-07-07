package compiler

import (
	"fmt"
	"monkey/ast"
	"monkey/code"
	"monkey/lexer"
	"monkey/object"
	"monkey/parser"
	"testing"
)

type compilerTestCase struct {
	input                string
	expectedConstants    []interface{}
	expectedInstructions []code.Instructions
}

func TestIntegerArithmetic(t *testing.T) {
	tests := []compilerTestCase{
		{
			input:             "1 + 2",
			expectedConstants: []interface{}{1, 2},
			expectedInstructions: []code.Instructions{
				code.Make(code.OpConstant, 0),
				code.Make(code.OpConstant, 1),
				code.Make(code.OpAdd),
				code.Make(code.OpPop),
			},
		},
		{
			input:             "1; 2",
			expectedConstants: []interface{}{1, 2},
			expectedInstructions: []code.Instructions{
				code.Make(code.OpConstant, 0),
				code.Make(code.OpPop),
				code.Make(code.OpConstant, 1),
				code.Make(code.OpPop),
			},
		},
	}

	runCompilerTests(t, tests)
}

func runCompilerTests(t *testing.T, tests []compilerTestCase) {
	t.Helper()

	for _, test := range tests {
		program := parse(test.input)
		compiler := New()
		err := compiler.Compile(program)
		if err != nil {
			t.Fatalf("compiler error: %s", err)
		}
		bytecode := compiler.Bytecode()
		err = testInstructions(test.expectedInstructions, bytecode.Instructions)
		if err != nil {
			t.Fatalf("testing instructions failed: %s", err)
		}
		err = testConstants(test.expectedConstants, bytecode.Constants)
		if err != nil {
			t.Fatalf("testing constants failed: %s", err)
		}
	}
}

func parse(input string) *ast.Program {
	l := lexer.New(input)
	p := parser.New(l)
	return p.ParseProgram()
}

func testInstructions(expected []code.Instructions, actual code.Instructions) error {
	concatInst := concatInstructions(expected)

	if len(actual) != len(concatInst) {
		return fmt.Errorf("wrong instruction length:\nexpected %q\ngot\t %q", concatInst, actual)
	}

	for i, ins := range concatInst {
		if actual[i] != ins {
			return fmt.Errorf("wrong instruction at %d:\nexpected %q\ngot\t %q", i, ins, actual[i])
		}
	}
	return nil
}

func concatInstructions(instructions []code.Instructions) (out code.Instructions) {
	for _, ins := range instructions {
		out = append(out, ins...)
	}
	return
}

func testConstants(
	expected []interface{},
	actual []object.Object,
) error {
	if len(actual) != len(expected) {
		return fmt.Errorf("wrong number of constants: expected %d, got %d", len(expected), len(actual))
	}

	for i, constant := range expected {
		switch constant := constant.(type) {
		case int:
			err := testIntegerObject(int64(constant), actual[i])

			if err != nil {
				return fmt.Errorf("constant %d: testIntegerObject failed: %s", i, err)
			}
		}
	}
	return nil
}

func testIntegerObject(expected int64, actual object.Object) error {
	result, ok := actual.(*object.Integer)
	if !ok {
		return fmt.Errorf("object ist not of type Integer. got %T (%+v)", actual, actual)
	}
	if result.Value != expected {
		return fmt.Errorf("object has wrong value. got %d, expected %d", result.Value, expected)
	}
	return nil
}