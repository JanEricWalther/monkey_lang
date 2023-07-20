package vm

import (
	"fmt"
	"monkey/code"
	"monkey/compiler"
	"monkey/object"
)

const StackSize = 2 << 10
const GlobalsSize = (2 << 15) - 1
const MaxFrames = 2 << 9

var True = &object.Boolean{Value: true}
var False = &object.Boolean{Value: false}
var Null = &object.Null{}

type VM struct {
	constants   []object.Object
	stack       []object.Object
	sp          int // points to next free slot
	globals     []object.Object
	frames      []*Frame
	framesIndex int
}

func New(bytecode *compiler.Bytecode) *VM {
	mainFn := &object.CompiledFunction{Instruction: bytecode.Instructions}
	mainFrame := NewFrame(mainFn, 0)
	frames := make([]*Frame, MaxFrames)
	frames[0] = mainFrame
	return &VM{
		constants:   bytecode.Constants,
		stack:       make([]object.Object, StackSize),
		sp:          0,
		globals:     make([]object.Object, GlobalsSize),
		frames:      frames,
		framesIndex: 1,
	}
}

func NewWithState(bytecode *compiler.Bytecode, s []object.Object) *VM {
	vm := New(bytecode)
	vm.globals = s
	return vm
}

func (vm *VM) Run() error {
	var ip int
	var ins code.Instructions
	var op code.Opcode

	for vm.currentFrame().ip < len(vm.currentFrame().Instructions())-1 {
		vm.currentFrame().ip += 1
		ip = vm.currentFrame().ip
		ins = vm.currentFrame().Instructions()
		op = code.Opcode(ins[ip])

		switch op {
		case code.OpConstant:
			constIdx := code.ReadUint16(ins[ip+1:])
			vm.currentFrame().ip += 2
			err := vm.push(vm.constants[constIdx])
			if err != nil {
				return err
			}
		case code.OpSetGlobal:
			index := code.ReadUint16(ins[ip+1:])
			vm.currentFrame().ip += 2
			vm.globals[index] = vm.pop()
		case code.OpGetGlobal:
			index := code.ReadUint16(ins[ip+1:])
			vm.currentFrame().ip += 2
			err := vm.push(vm.globals[index])
			if err != nil {
				return err
			}
		case code.OpSetLocal:
			localIdx := code.ReadUint8(ins[ip+1:])
			vm.currentFrame().ip += 1
			frame := vm.currentFrame()
			vm.stack[frame.basePointer+int(localIdx)] = vm.pop()
		case code.OpGetLocal:
			localIdx := code.ReadUint8(ins[ip+1:])
			vm.currentFrame().ip += 1
			frame := vm.currentFrame()
			err := vm.push(vm.stack[frame.basePointer+int(localIdx)])
			if err != nil {
				return err
			}
		case code.OpGetBuiltin:
			builtinIdx := code.ReadUint8(ins[ip+1:])
			vm.currentFrame().ip += 1

			definition := object.Builtins[builtinIdx]

			err := vm.push(definition.Builtin)
			if err != nil {
				return err
			}
		case code.OpArray:
			numElements := int(code.ReadUint16(ins[ip+1:]))
			vm.currentFrame().ip += 2
			array := vm.buildArray(vm.sp-numElements, vm.sp)
			vm.sp = vm.sp - numElements
			err := vm.push(array)
			if err != nil {
				return err
			}
		case code.OpHash:
			numElements := int(code.ReadUint16(ins[ip+1:]))
			vm.currentFrame().ip += 2
			hash, err := vm.buildHash(vm.sp-numElements, vm.sp)
			if err != nil {
				return err
			}
			vm.sp = vm.sp - numElements
			err = vm.push(hash)
			if err != nil {
				return err
			}
		case code.OpIndex:
			index := vm.pop()
			left := vm.pop()

			err := vm.executeIndexExpression(left, index)
			if err != nil {
				return err
			}

		case code.OpTrue:
			err := vm.push(True)
			if err != nil {
				return err
			}
		case code.OpFalse:
			err := vm.push(False)
			if err != nil {
				return err
			}
		case code.OpNull:
			err := vm.push(Null)
			if err != nil {
				return err
			}
		case code.OpBang:
			err := vm.executeBangOperator()
			if err != nil {
				return err
			}
		case code.OpMinus:
			err := vm.executeMinusOperator()
			if err != nil {
				return err
			}
		case code.OpAdd, code.OpSub, code.OpMul, code.OpDiv:
			err := vm.executeBinaryOparation(op)
			if err != nil {
				return err
			}
		case code.OpEqual, code.OpNotEqual, code.OpGreaterThan:
			err := vm.executeComparison(op)
			if err != nil {
				return err
			}
		case code.OpPop:
			vm.pop()
		case code.OpJump:
			pos := int(code.ReadUint16(ins[ip+1:]))
			vm.currentFrame().ip = pos - 1
		case code.OpJumpNotTruthy:
			pos := int(code.ReadUint16(ins[ip+1:]))
			vm.currentFrame().ip += 2
			condition := vm.pop()
			if !isTruthy(condition) {
				vm.currentFrame().ip = pos - 1
			}
		case code.OpCall:
			numArgs := code.ReadUint8(ins[ip+1:])
			vm.currentFrame().ip += 1

			err := vm.executeCall(int(numArgs))
			if err != nil {
				return err
			}
		case code.OpReturnValue:
			returnValue := vm.pop()
			frame := vm.popFrame()
			vm.sp = frame.basePointer - 1

			err := vm.push(returnValue)
			if err != nil {
				return err
			}
		case code.OpReturn:
			frame := vm.popFrame()
			// NOTE(jan): set to basePointer-1 to pop off the top value
			vm.sp = frame.basePointer - 1
			err := vm.push(Null)
			if err != nil {
				return err
			}
		}

	}
	return nil
}

func (vm *VM) StackTop() object.Object {
	if vm.sp == 0 {
		return nil
	}
	return vm.stack[vm.sp-1]
}

func (vm *VM) LastPoppedStackElement() object.Object {
	return vm.stack[vm.sp]
}

func (vm *VM) push(o object.Object) error {
	if vm.sp >= StackSize {
		return fmt.Errorf("stack overflow")
	}
	vm.stack[vm.sp] = o
	vm.sp += 1
	return nil
}

func (vm *VM) pop() object.Object {
	o := vm.stack[vm.sp-1]
	vm.sp -= 1
	return o
}

func (vm *VM) executeBinaryOparation(op code.Opcode) error {

	right := vm.pop()
	left := vm.pop()

	leftType := left.Type()
	rightType := right.Type()
	switch {
	case leftType == object.INTEGER_OBJ && leftType == rightType:
		return vm.executeBinaryIntegerOparation(op, left, right)
	case leftType == object.STRING_OBJ && leftType == rightType:
		return vm.executeBinaryStringOparation(op, left, right)
	default:
		return fmt.Errorf("unsupported types for binary operation: %s %s", leftType, rightType)
	}
}

func (vm *VM) executeBinaryIntegerOparation(op code.Opcode, left, right object.Object) error {
	leftVal := left.(*object.Integer).Value
	rightVal := right.(*object.Integer).Value

	var result int64

	switch op {
	case code.OpAdd:
		result = leftVal + rightVal
	case code.OpSub:
		result = leftVal - rightVal
	case code.OpMul:
		result = leftVal * rightVal
	case code.OpDiv:
		result = leftVal / rightVal
	default:
		return fmt.Errorf("unknown integer operator: %d", op)
	}

	return vm.push(&object.Integer{Value: result})
}

func (vm *VM) executeBinaryStringOparation(op code.Opcode, left, right object.Object) error {
	if op != code.OpAdd {
		return fmt.Errorf("unknown string operator: %d", op)
	}
	leftVal := left.(*object.String).Value
	rightVal := right.(*object.String).Value

	return vm.push(&object.String{Value: leftVal + rightVal})
}

func (vm *VM) executeComparison(op code.Opcode) error {
	right := vm.pop()
	left := vm.pop()

	if left.Type() == object.INTEGER_OBJ && right.Type() == object.INTEGER_OBJ {
		return vm.executeIntegerComparison(op, left, right)
	}
	switch op {
	case code.OpEqual:
		return vm.push(nativeBoolToBooleanObject(left == right))
	case code.OpNotEqual:
		return vm.push(nativeBoolToBooleanObject(left != right))
	default:
		return fmt.Errorf("unknown operator: %d (%s %s)", op, left.Type(), right.Type())
	}
}

func (vm *VM) executeIntegerComparison(op code.Opcode, left, right object.Object) error {
	leftVal := left.(*object.Integer).Value
	rightVal := right.(*object.Integer).Value

	switch op {
	case code.OpEqual:
		return vm.push(nativeBoolToBooleanObject(leftVal == rightVal))
	case code.OpNotEqual:
		return vm.push(nativeBoolToBooleanObject(leftVal != rightVal))
	case code.OpGreaterThan:
		return vm.push(nativeBoolToBooleanObject(leftVal > rightVal))
	default:
		return fmt.Errorf("onknown operator: %d", op)
	}
}

func (vm *VM) executeBangOperator() error {
	op := vm.pop()
	switch op {
	case True:
		return vm.push(False)
	case False, Null:
		return vm.push(True)
	default:
		return vm.push(False)
	}
}

func (vm *VM) executeMinusOperator() error {
	op := vm.pop()
	if op.Type() != object.INTEGER_OBJ {
		return fmt.Errorf("unsupported type for negation: %s", op.Type())
	}
	value := op.(*object.Integer).Value
	return vm.push(&object.Integer{Value: -value})
}

func (vm *VM) executeIndexExpression(left, index object.Object) error {
	switch {
	case left.Type() == object.ARRAY_OBJ && index.Type() == object.INTEGER_OBJ:
		return vm.executeArrayIndex(left, index)
	case left.Type() == object.HASH_OBJ:
		return vm.executeHashIndex(left, index)
	default:
		return fmt.Errorf("index operator not supported: %s", left.Type())
	}
}

func (vm *VM) executeArrayIndex(left, index object.Object) error {
	arrayObj := left.(*object.Array)
	i := index.(*object.Integer).Value
	max := int64(len(arrayObj.Elements) - 1)
	if i < 0 || i > max {
		return vm.push(Null)
	}
	return vm.push(arrayObj.Elements[i])
}

func (vm *VM) executeHashIndex(left, index object.Object) error {
	hashObj := left.(*object.Hash)
	key, ok := index.(object.Hashable)
	if !ok {
		return fmt.Errorf("unusable as hash key: %s", index.Type())
	}
	pair, ok := hashObj.Pairs[key.HashKey()]
	if !ok {
		return vm.push(Null)
	}
	return vm.push(pair.Value)
}

func (vm *VM) buildArray(start, end int) object.Object {
	elements := make([]object.Object, end-start)
	for i := start; i < end; i++ {
		elements[i-start] = vm.stack[i]
	}
	return &object.Array{Elements: elements}
}

func (vm *VM) buildHash(start, end int) (object.Object, error) {
	hashPairs := make(map[object.HashKey]object.HashPair)
	for i := start; i < end; i += 2 {
		key := vm.stack[i]
		value := vm.stack[i+1]

		pair := object.HashPair{Key: key, Value: value}
		hashKey, ok := key.(object.Hashable)
		if !ok {
			return nil, fmt.Errorf("unusable as hash key: %s", key.Type())
		}
		hashPairs[hashKey.HashKey()] = pair
	}
	return &object.Hash{Pairs: hashPairs}, nil
}

func (vm *VM) currentFrame() *Frame {
	return vm.frames[vm.framesIndex-1]
}

func (vm *VM) pushFrame(f *Frame) {
	vm.frames[vm.framesIndex] = f
	vm.framesIndex += 1
}

func (vm *VM) popFrame() *Frame {
	vm.framesIndex -= 1
	return vm.frames[vm.framesIndex]
}

func (vm *VM) executeCall(numArgs int) error {
	callee := vm.stack[vm.sp-1-numArgs]
	switch callee := callee.(type) {
	case *object.CompiledFunction:
		return vm.callFunction(callee, numArgs)
	case *object.Builtin:
		return vm.callBuiltin(callee, numArgs)
	default:
		return fmt.Errorf("calling non-function and non-built-in")
	}
}

func (vm *VM) callFunction(fn *object.CompiledFunction, numArgs int) error {
	if numArgs != fn.NumParameters {
		return fmt.Errorf("wrong number of arguments: expected %d, got %d", fn.NumParameters, numArgs)
	}
	frame := NewFrame(fn, vm.sp-numArgs)
	vm.pushFrame(frame)
	vm.sp = frame.basePointer + fn.NumLocals
	return nil
}

func (vm *VM) callBuiltin(fn *object.Builtin, numArgs int) error {
	args := vm.stack[vm.sp-numArgs : vm.sp]
	result := fn.Fn(args...)
	vm.sp -= numArgs + 1

	if result != nil {
		vm.push(result)
	} else {
		vm.push(Null)
	}
	return nil
}

func nativeBoolToBooleanObject(b bool) *object.Boolean {
	if b {
		return True
	}
	return False
}

func isTruthy(obj object.Object) bool {
	switch obj := obj.(type) {
	case *object.Boolean:
		return obj.Value
	case *object.Integer:
		return obj.Value != 0
	case *object.Null:
		return false
	default:
		return true
	}
}
