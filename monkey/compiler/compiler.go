package compiler

import (
	"fmt"
	"monkey/ast"
	"monkey/code"
	"monkey/object"
	"sort"
)

type EmittedInstruction struct {
	Opcode   code.Opcode
	Position int
}

type CopmilationScope struct {
	instructions        code.Instructions
	lastInstruction     EmittedInstruction
	previousInstruction EmittedInstruction
}

type Bytecode struct {
	Instructions code.Instructions
	Constants    []object.Object
}

type Compiler struct {
	constants  []object.Object
	symTable   *SymbolTable
	scopes     []CopmilationScope
	scopeIndex int
}

func New() *Compiler {
	mainScope := CopmilationScope{
		instructions:        code.Instructions{},
		lastInstruction:     EmittedInstruction{},
		previousInstruction: EmittedInstruction{},
	}
	return &Compiler{
		constants:  []object.Object{},
		symTable:   NewSymbolTable(),
		scopes:     []CopmilationScope{mainScope},
		scopeIndex: 0,
	}
}

func NewWithState(s *SymbolTable, constants []object.Object) *Compiler {
	c := New()
	c.symTable = s
	c.constants = constants
	return c
}

func (c *Compiler) Compile(node ast.Node) error {
	switch node := node.(type) {
	case *ast.Program:
		for _, statement := range node.Statements {
			err := c.Compile(statement)
			if err != nil {
				return err
			}
		}
	case *ast.ExpressionStatement:
		err := c.Compile(node.Expression)
		if err != nil {
			return err
		}
		c.emit(code.OpPop)
	case *ast.BlockStatement:
		for _, statement := range node.Statements {
			err := c.Compile(statement)
			if err != nil {
				return err
			}
		}
	case *ast.Boolean:
		if node.Value {
			c.emit(code.OpTrue)
		} else {
			c.emit(code.OpFalse)
		}
	case *ast.IntegerLiteral:
		integer := &object.Integer{Value: node.Value}
		c.emit(code.OpConstant, c.addConstant(integer))
	case *ast.StringLiteral:
		str := &object.String{Value: node.Value}
		c.emit(code.OpConstant, c.addConstant(str))
	case *ast.ArrayLiteral:
		for _, el := range node.Elements {
			err := c.Compile(el)
			if err != nil {
				return err
			}
		}
		c.emit(code.OpArray, len(node.Elements))
	case *ast.HashLiteral:
		keys := []ast.Expression{}
		for k := range node.Pairs {
			keys = append(keys, k)
		}
		sort.Slice(keys, func(i, j int) bool {
			return keys[i].String() < keys[j].String()
		})

		for _, k := range keys {
			err := c.Compile(k)
			if err != nil {
				return err
			}
			err = c.Compile(node.Pairs[k])
			if err != nil {
				return err
			}
		}

		c.emit(code.OpHash, len(node.Pairs)*2)

	case *ast.LetStatement:
		err := c.Compile(node.Value)
		if err != nil {
			return err
		}
		symbol := c.symTable.Define(node.Name.Value)
		if symbol.Scope == GlobalScope {
			c.emit(code.OpSetGlobal, symbol.Index)
		} else {
			c.emit(code.OpSetLocal, symbol.Index)
		}
	case *ast.Identifier:
		sym, ok := c.symTable.Resolve(node.Value)
		if !ok {
			return fmt.Errorf("undefined variable %s", node.Value)
		}
		if sym.Scope == GlobalScope {
			c.emit(code.OpGetGlobal, sym.Index)
		} else {
			c.emit(code.OpGetLocal, sym.Index)
		}
	case *ast.PrefixExpression:
		err := c.Compile(node.Right)
		if err != nil {
			return err
		}
		op, err := getPrefixOperator(node.Operator)
		if err != nil {
			return err
		}
		c.emit(op)
	case *ast.InfixExpression:
		if node.Operator == "<" {
			err := c.Compile(node.Right)
			if err != nil {
				return err
			}
			err = c.Compile(node.Left)
			if err != nil {
				return err
			}
			c.emit(code.OpGreaterThan)
			return nil
		}
		err := c.Compile(node.Left)
		if err != nil {
			return err
		}
		err = c.Compile(node.Right)
		if err != nil {
			return err
		}
		op, err := getInfixOperator(node.Operator)
		if err != nil {
			return err
		}
		c.emit(op)
	case *ast.IfExpression:
		err := c.Compile(node.Condition)
		if err != nil {
			return err
		}
		// NOTE(jan): 42069 is a placeholder, that gets changed later
		jumpNotTruthyPos := c.emit(code.OpJumpNotTruthy, 42069)
		err = c.Compile(node.Consequence)
		if err != nil {
			return err
		}
		if c.lastInstructionIs(code.OpPop) {
			c.removeLastInstruction()
		}
		jmpPos := c.emit(code.OpJump, 42069)

		afterConseqPos := len(c.currentInstructions())
		c.changeOperand(jumpNotTruthyPos, afterConseqPos)
		if node.Alternative == nil {
			c.emit(code.OpNull)
		} else {
			err := c.Compile(node.Alternative)
			if err != nil {
				return err
			}
			if c.lastInstructionIs(code.OpPop) {
				c.removeLastInstruction()
			}
		}
		afterAlternativePos := len(c.currentInstructions())
		c.changeOperand(jmpPos, afterAlternativePos)
	case *ast.IndexExpression:
		err := c.Compile(node.Left)
		if err != nil {
			return err
		}

		err = c.Compile(node.Index)
		if err != nil {
			return err
		}
		c.emit(code.OpIndex)
	case *ast.FunctionLiteral:
		c.enterScope()

		for _, param := range node.Parameters {
			c.symTable.Define(param.Value)
		}

		err := c.Compile(node.Body)
		if err != nil {
			return err
		}
		if c.lastInstructionIs(code.OpPop) {
			c.replaceLastPopWithReturn()
		}
		if !c.lastInstructionIs(code.OpReturnValue) {
			c.emit(code.OpReturn)
		}

		numLocals := c.symTable.numDefinitions
		instructions := c.leaveScope()

		compiledFn := &object.CompiledFunction{
			Instruction:   instructions,
			NumLocals:     numLocals,
			NumParameters: len(node.Parameters),
		}
		c.emit(code.OpConstant, c.addConstant(compiledFn))
	case *ast.ReturnStatement:
		err := c.Compile(node.ReturnValue)
		if err != nil {
			return err
		}
		c.emit(code.OpReturnValue)
	case *ast.CallExpression:
		err := c.Compile(node.Function)
		if err != nil {
			return err
		}
		for _, argument := range node.Arguments {
			err := c.Compile(argument)
			if err != nil {
				return err
			}
		}
		c.emit(code.OpCall, len(node.Arguments))
	}
	return nil
}

func (c *Compiler) Bytecode() *Bytecode {
	return &Bytecode{
		Instructions: c.currentInstructions(),
		Constants:    c.constants,
	}
}

func (c *Compiler) addConstant(obj object.Object) int {
	c.constants = append(c.constants, obj)
	return len(c.constants) - 1
}

// returns current instrution position
func (c *Compiler) emit(op code.Opcode, operands ...int) int {
	ins := code.Make(op, operands...)
	pos := c.addInstruction(ins)
	c.setLastInstruction(op, pos)
	return pos
}

func (c *Compiler) currentInstructions() code.Instructions {
	return c.scopes[c.scopeIndex].instructions
}

func (c *Compiler) addInstruction(ins []byte) int {
	newPos := len(c.currentInstructions())
	updatedInstructions := append(c.currentInstructions(), ins...)

	c.scopes[c.scopeIndex].instructions = updatedInstructions
	return newPos
}

func (c *Compiler) setLastInstruction(op code.Opcode, pos int) {
	previous := c.scopes[c.scopeIndex].lastInstruction
	last := EmittedInstruction{Opcode: op, Position: pos}
	c.scopes[c.scopeIndex].previousInstruction = previous
	c.scopes[c.scopeIndex].lastInstruction = last

}

func getInfixOperator(opString string) (op code.Opcode, err error) {
	switch opString {
	case "+":
		return code.OpAdd, nil
	case "-":
		return code.OpSub, nil
	case "*":
		return code.OpMul, nil
	case "/":
		return code.OpDiv, nil
	case ">":
		return code.OpGreaterThan, nil
	case "==":
		return code.OpEqual, nil
	case "!=":
		return code.OpNotEqual, nil
	}
	return 0, fmt.Errorf("unkown operator %s", opString)
}

func getPrefixOperator(opString string) (op code.Opcode, err error) {
	switch opString {
	case "-":
		return code.OpMinus, nil
	case "!":
		return code.OpBang, nil
	}
	return 0, fmt.Errorf("unkown operator %s", opString)
}

func (c *Compiler) lastInstructionIs(op code.Opcode) bool {
	if len(c.currentInstructions()) == 0 {
		return false
	}
	return c.scopes[c.scopeIndex].lastInstruction.Opcode == op
}

func (c *Compiler) removeLastInstruction() {
	last := c.scopes[c.scopeIndex].lastInstruction
	previous := c.scopes[c.scopeIndex].previousInstruction

	old := c.currentInstructions()
	new := old[:last.Position]

	c.scopes[c.scopeIndex].instructions = new
	c.scopes[c.scopeIndex].lastInstruction = previous
}

func (c *Compiler) replaceInstruction(pos int, newInstruction []byte) {
	ins := c.currentInstructions()
	for i, bit := range newInstruction {
		ins[pos+i] = bit
	}
}

func (c *Compiler) replaceLastPopWithReturn() {
	lastPos := c.scopes[c.scopeIndex].lastInstruction.Position
	c.replaceInstruction(lastPos, code.Make(code.OpReturnValue))

	c.scopes[c.scopeIndex].lastInstruction.Opcode = code.OpReturnValue
}

func (c *Compiler) changeOperand(opPos int, operand int) {
	op := code.Opcode(c.currentInstructions()[opPos])
	newInstruction := code.Make(op, operand)
	c.replaceInstruction(opPos, newInstruction)
}

func (c *Compiler) enterScope() {
	scope := CopmilationScope{
		instructions:        code.Instructions{},
		lastInstruction:     EmittedInstruction{},
		previousInstruction: EmittedInstruction{},
	}
	c.scopes = append(c.scopes, scope)
	c.scopeIndex += 1
	c.symTable = NewEnclosedSymbolTable(c.symTable)
}

func (c *Compiler) leaveScope() code.Instructions {
	instructions := c.currentInstructions()
	c.scopes = c.scopes[:len(c.scopes)-1]
	c.scopeIndex -= 1
	c.symTable = c.symTable.Outer
	return instructions
}
