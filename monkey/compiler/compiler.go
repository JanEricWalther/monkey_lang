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

type Compiler struct {
	instructions        code.Instructions
	constants           []object.Object
	lastInstruction     EmittedInstruction
	previousInstruction EmittedInstruction
	symTable            *SymbolTable
}

func New() *Compiler {
	return &Compiler{
		instructions:        code.Instructions{},
		constants:           []object.Object{},
		lastInstruction:     EmittedInstruction{},
		previousInstruction: EmittedInstruction{},
		symTable:            NewSymbolTable(),
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
		c.emit(code.OpSetGlobal, symbol.Index)
	case *ast.Identifier:
		sym, ok := c.symTable.Resolve(node.Value)
		if !ok {
			return fmt.Errorf("undefined variable %s", node.Value)
		}
		c.emit(code.OpGetGlobal, sym.Index)
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
		if c.lastInstructionIsPop() {
			c.removeLastInstruction()
		}
		jmpPos := c.emit(code.OpJump, 42069)

		afterConseqPos := len(c.instructions)
		c.changeOperand(jumpNotTruthyPos, afterConseqPos)
		if node.Alternative == nil {
			c.emit(code.OpNull)
		} else {
			err := c.Compile(node.Alternative)
			if err != nil {
				return err
			}
			if c.lastInstructionIsPop() {
				c.removeLastInstruction()
			}
		}
		afterAlternativePos := len(c.instructions)
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
	}
	return nil
}

func (c *Compiler) Bytecode() *Bytecode {
	return &Bytecode{
		Instructions: c.instructions,
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

func (c *Compiler) addInstruction(ins []byte) int {
	newPos := len(c.instructions)
	c.instructions = append(c.instructions, ins...)
	return newPos
}

func (c *Compiler) setLastInstruction(op code.Opcode, pos int) {
	// previous := c.lastInstruction
	// last := EmittedInstruction{Opcode: op, Position: pos}
	// c.previousInstruction = previous
	// c.lastInstruction = last

	c.previousInstruction, c.lastInstruction = c.lastInstruction, EmittedInstruction{Opcode: op, Position: pos}
}

type Bytecode struct {
	Instructions code.Instructions
	Constants    []object.Object
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

func (c *Compiler) lastInstructionIsPop() bool {
	return c.lastInstruction.Opcode == code.OpPop
}

func (c *Compiler) removeLastInstruction() {
	c.instructions = c.instructions[:c.lastInstruction.Position]
	c.lastInstruction = c.previousInstruction
}

func (c *Compiler) replaceInstruction(pos int, newInstruction []byte) {
	for i, bit := range newInstruction {
		c.instructions[pos+i] = bit
	}
}

func (c *Compiler) changeOperand(opPos int, operand int) {
	op := code.Opcode(c.instructions[opPos])
	newInstruction := code.Make(op, operand)
	c.replaceInstruction(opPos, newInstruction)
}
