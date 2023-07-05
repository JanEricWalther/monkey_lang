package vm

import (
	"fmt"
	"monkey/code"
	"monkey/compiler"
	"monkey/object"
)

const StackSize = 2 * 1024

type VM struct {
	constants    []object.Object
	instructions code.Instructions
	stack        []object.Object
	sp           int // points to next free slot
}

func New(bytecode *compiler.Bytecode) *VM {
	return &VM{
		instructions: bytecode.Instructions,
		constants:    bytecode.Constants,
		stack:        make([]object.Object, StackSize),
		sp:           0,
	}
}

func (vm *VM) Run() error {
	for instP := 0; instP < len(vm.instructions); instP++ {
		op := code.Opcode(vm.instructions[instP])

		switch op {
		case code.OpConstant:
			constIdx := code.ReadUint16(vm.instructions[instP+1:])
			instP += 2
			err := vm.push(vm.constants[constIdx])
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

func (vm *VM) push(o object.Object) error {
	if vm.sp >= StackSize {
		return fmt.Errorf("stack overflow")
	}
	vm.stack[vm.sp] = o
	vm.sp += 1
	return nil
}
