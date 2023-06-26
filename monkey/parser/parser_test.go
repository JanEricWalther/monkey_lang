package parser

import (
	"monkey/ast"
	"monkey/lexer"
	"testing"
)

const letInput = `
	let x = 5;
	let y = 10;
	let foobar = 838383;
`
const returnInput = `
	return 5;
	return 10;
	return 992233;
`

type expected struct {
	Identifier string
}

func TestLetStatements(t *testing.T) {
	l := lexer.New(letInput)
	p := New(l)

	program := p.ParseProgram()
	checkParserErrors(t, p)
	if program == nil {
		t.Fatalf("ParseProgram() returned nil")
	}
	if len(program.Statements) != 3 {
		t.Fatalf("program.Statements does not contain 3 statemenets. got %d", len(program.Statements))
	}

	tests := []expected{
		{"x"},
		{"y"},
		{"foobar"},
	}

	for i, test := range tests {
		stmt := program.Statements[i]
		if !testLetStatements(t, stmt, test.Identifier) {
			return
		}
	}
}

func TestReturnStatements(t *testing.T) {
	l := lexer.New(returnInput)
	p := New(l)
	program := p.ParseProgram()
	checkParserErrors(t, p)

	if len(program.Statements) != 3 {
		t.Fatalf("program.Statements does not contain 3 statemenets. got %d", len(program.Statements))
	}
	for _, statement := range program.Statements {
		returnStatement, ok := statement.(*ast.ReturnStatement)
		if !ok {
			t.Errorf("statement not ReturnStatement. got %q", statement)
			continue
		}
		if returnStatement.TokenLiteral() != "return" {
			t.Errorf("returnStatement TokenLiteral not 'return'. got %q", returnStatement.TokenLiteral())
		}
	}
}

func testLetStatements(t *testing.T, s ast.Statement, name string) bool {
	if s.TokenLiteral() != "let" {
		t.Errorf("s.TokenLiteral not 'let'. got %q", s.TokenLiteral())
		return false
	}

	letStmt, ok := s.(*ast.LetStatement)
	if !ok {
		t.Errorf("s not *ast.LetStatement. got %T", s)
		return false
	}

	if letStmt.Name.Value != name {
		t.Errorf("letStmt.Name.Value not '%s'. got %s", name, letStmt.Name.Value)
		return false
	}

	if letStmt.Name.TokenLiteral() != name {
		t.Errorf("letStmt.Name.TokenLiteral() not '%s'. got %s", name, letStmt.Name.TokenLiteral())
		return false
	}
	return true
}

func checkParserErrors(t *testing.T, p *Parser) {
	errors := p.Errors()
	if len(errors) == 0 {
		return
	}

	t.Errorf("parser has %d errors", len(errors))
	for _, message := range errors {
		t.Errorf("parse error: %q", message)
	}
	t.FailNow()
}
