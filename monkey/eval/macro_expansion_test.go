package eval

import (
	"fmt"
	"monkey/ast"
	"monkey/lexer"
	"monkey/object"
	"monkey/parser"
	"testing"
)

func TestDefineMacros(t *testing.T) {
	input := `
	let number = 1;
	let function = fn(x, y) { x + y };
	let mymacro = macro(x, y) { x + y; };
	`
	expectedBody := "(x + y)"

	env := object.NewEnv()
	program := testParseProgram(input)

	DefineMacros(program, env)

	if len(program.Statements) != 2 {
		t.Fatalf("Wrong number of statements. got %d", len(program.Statements))
	}

	_, ok := env.Get("number")
	if ok {
		t.Fatalf("number should not be defined")
	}
	_, ok = env.Get("function")
	if ok {
		t.Fatalf("function should not be defined")
	}

	obj, ok := env.Get("mymacro")
	if !ok {
		t.Fatalf("macro not in environment")
	}

	macro, ok := obj.(*object.Macro)
	if !ok {
		t.Fatalf("object is not macro. got %T (%+v)", obj, obj)
	}

	if len(macro.Parameters) != 2 {
		t.Fatalf("Wrong number of macro parameters. got %d", len(macro.Parameters))
	}

	if macro.Parameters[0].String() != "x" {
		t.Fatalf("paramter is not 'x'. got %q", macro.Parameters[0])
	}
	if macro.Parameters[1].String() != "y" {
		t.Fatalf("paramter is not 'y'. got %q", macro.Parameters[1])
	}

	if macro.Body.String() != expectedBody {
		t.Fatalf("body is not %q. got %q", expectedBody, macro.Body.String())
	}
}

func TestExpandMacros(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{
			`
			let infixExpression = macro() { quote(1 + 2); };
			infixExpression();
			`,
			`(1 + 2)`,
		},
		{
			`
			let reverse = macro(a, b) { quote(unquote(b) - unquote(a)); };
			reverse(2 + 2, 10 - 5);
			`,
			`(10 - 5) - (2 + 2)`,
		},
		{
			`
			let unless = macro(condition, consequence, alternative) {
			    quote(if (!(unquote(condition))) {
			        unquote(consequence);
			    } else {
			        unquote(alternative);
			    });
			};
			unless(10 > 5, print("not greater"), print ("greater"));
			`,
			`if (!(10 > 5)) { print("not greater") } else { print("greater") }`,
		},
	}

	for _, test := range tests {
		expected := testParseProgram(test.expected)
		program := testParseProgram(test.input)

		fmt.Println(program.String())
		env := object.NewEnv()
		DefineMacros(program, env)
		expanded := ExpandMacros(program, env)
		fmt.Println(expanded.String())

		if expanded.String() != expected.String() {
			t.Errorf("not equal. got %q, expected %q", expanded.String(), expected.String())
		}
	}
}

func testParseProgram(input string) *ast.Program {
	l := lexer.New(input)
	p := parser.New(l)
	return p.ParseProgram()
}
