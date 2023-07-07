package code

import (
	"bytes"
	"encoding/binary"
	"fmt"
)

const (
	OpConstant Opcode = iota
	OpAdd
	OpPop
)

var definitions = map[Opcode]*Definition{
	OpConstant: {"OpConstant", []int{2}},
	OpAdd:      {"OpAdd", []int{}},
	OpPop:      {"OpPop", []int{}},
}

type Instructions []byte

func (in Instructions) String() string {
	var out bytes.Buffer

	for i := 0; i < len(in); i++ {
		def, err := Lookup(in[i])
		if err != nil {
			fmt.Fprintf(&out, "ERROR: %s\n", err)
			continue
		}
		operands, read := ReadOperands(def, in[i+1:])
		fmt.Fprintf(&out, "%04d %s\n", i, in.fmtInstruction(def, operands))

		i += read
	}
	return out.String()
}

func (in Instructions) fmtInstruction(def *Definition, operands []int) string {
	opCount := len(def.OpWidths)
	if len(operands) != opCount {
		return fmt.Sprintf("ERROR: operand len %d does not match defined %d\n", len(operands), opCount)
	}
	switch opCount {
	case 0:
		return def.Name
	case 1:
		return fmt.Sprintf("%s %d", def.Name, operands[0])
	}
	return fmt.Sprintf("ERROR: unhandled operandCount for %s\n", def.Name)
}

type Opcode byte

type Definition struct {
	Name     string
	OpWidths []int
}

func Lookup(op byte) (*Definition, error) {
	def, ok := definitions[Opcode(op)]
	if !ok {
		return nil, fmt.Errorf("opcode %d undefined", op)
	}
	return def, nil
}

func Make(op Opcode, operands ...int) []byte {
	def, ok := definitions[op]
	if !ok {
		return []byte{}
	}

	instructionLen := 1
	for _, w := range def.OpWidths {
		instructionLen += w
	}

	instruction := make([]byte, instructionLen)
	instruction[0] = byte(op)

	offset := 1
	for i, o := range operands {
		width := def.OpWidths[i]
		switch width {
		case 2:
			binary.BigEndian.PutUint16(instruction[offset:], uint16(o))
		}
		offset += width
	}
	return instruction
}

func ReadOperands(def *Definition, ins Instructions) (operands []int, offset int) {
	operands = make([]int, len(def.OpWidths))

	for i, width := range def.OpWidths {
		switch width {
		case 2:
			operands[i] = int(ReadUint16(ins[offset:]))
		}
		offset += width
	}
	return
}

func ReadUint16(ins Instructions) uint16 {
	return binary.BigEndian.Uint16(ins)
}