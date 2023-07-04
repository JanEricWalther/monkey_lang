package eval

import (
	"fmt"
	"monkey/ast"
	"monkey/object"
	"monkey/token"
)

func quote(node ast.Node, env *object.Environment) object.Object {
	node = evalUnqouteCalls(node, env)
	return &object.Quote{Node: node}
}

func evalUnqouteCalls(qouted ast.Node, env *object.Environment) ast.Node {
	return ast.Modify(qouted, func(node ast.Node) ast.Node {
		if !isUnqouted(node) {
			return node
		}
		call, ok := node.(*ast.CallExpression)
		if !ok {
			return node
		}
		if len(call.Arguments) != 1 {
			return node
		}
		return convertObjectToASTNode(Eval(call.Arguments[0], env))
	})
}

func isUnqouted(node ast.Node) bool {
	exp, ok := node.(*ast.CallExpression)
	if !ok {
		return false
	}
	return exp.Function.TokenLiteral() == "unquote"
}

func convertObjectToASTNode(obj object.Object) ast.Node {
	switch obj := obj.(type) {
	case *object.Integer:
		t := token.Token{
			Type:    token.INT,
			Literal: fmt.Sprintf("%d", obj.Value),
		}
		return &ast.IntegerLiteral{Token: t, Value: obj.Value}
	case *object.Boolean:
		t := token.Token{
			Type:    token.FALSE,
			Literal: "false",
		}
		if obj.Value {
			t.Type = token.TRUE
			t.Literal = "true"
		}
		return &ast.Boolean{Token: t, Value: obj.Value}
	case *object.Quote:
		return obj.Node
	default:
		// TODO(jan): handle errors
		return nil
	}
}
