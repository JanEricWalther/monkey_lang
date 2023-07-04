package ast

// TODO(jan): add error handling
// TODO(jan): update Token Fields on Parent Nodes
func Modify(node Node, modifier func(Node) Node) Node {
	switch node := node.(type) {
	case *Program:
		for i, statement := range node.Statements {
			// The following line is the key to the whole thing.
			// It is the recursive call to Modify that makes this
			// work. It will continue to call Modify until it
			// reaches the bottom of the AST tree.
			node.Statements[i], _ = Modify(statement, modifier).(Statement)
		}
	case *ExpressionStatement:
		node.Expression, _ = Modify(node.Expression, modifier).(Expression)
	case *InfixExpression:
		node.Left, _ = Modify(node.Left, modifier).(Expression)
		node.Right, _ = Modify(node.Right, modifier).(Expression)
	case *PrefixExpression:
		node.Right, _ = Modify(node.Right, modifier).(Expression)
	case *IndexExpression:
		node.Left, _ = Modify(node.Left, modifier).(Expression)
		node.Index, _ = Modify(node.Index, modifier).(Expression)
	case *IfExpression:
		node.Condition, _ = Modify(node.Condition, modifier).(Expression)
		node.Consequence, _ = Modify(node.Consequence, modifier).(*BlockStatement)
		if node.Alternative != nil {
			node.Alternative, _ = Modify(node.Alternative, modifier).(*BlockStatement)
		}
	case *BlockStatement:
		for i, statement := range node.Statements {
			node.Statements[i], _ = Modify(statement, modifier).(Statement)
		}
	case *FunctionLiteral:
		for i, param := range node.Parameters {
			node.Parameters[i], _ = Modify(param, modifier).(*Identifier)
		}
		node.Body, _ = Modify(node.Body, modifier).(*BlockStatement)
	case *ArrayLiteral:
		for i, elem := range node.Elements {
			node.Elements[i], _ = Modify(elem, modifier).(Expression)
		}
	case *ReturnStatement:
		node.ReturnValue, _ = Modify(node.ReturnValue, modifier).(Expression)
	case *LetStatement:
		node.Value, _ = Modify(node.Value, modifier).(Expression)
	case *HashLiteral:
		newPairs := make(map[Expression]Expression)
		for key, value := range node.Pairs {
			newKey, _ := Modify(key, modifier).(Expression)
			newVal, _ := Modify(value, modifier).(Expression)
			newPairs[newKey] = newVal
		}
		node.Pairs = newPairs
	}
	return modifier(node)
}
