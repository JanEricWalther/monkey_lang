package code

import "testing"

func TestMake(t *testing.T) {
	tests := []struct {
		op       Opcode
		operands []int
		expected []byte
	}{
		{
			OpConstant,
			[]int{65534},
			[]byte{byte(OpConstant), 255, 254},
		},
		{
			OpAdd,
			[]int{},
			[]byte{byte(OpAdd)},
		},
		{
			OpGetLocal,
			[]int{255},
			[]byte{byte(OpGetLocal), 255},
		},
	}

	for _, test := range tests {
		instruction := Make(test.op, test.operands...)

		if len(instruction) != len(test.expected) {
			t.Errorf("instruction has wrong length. got %d, expected %d", len(instruction), len(test.expected))
		}
		for i, b := range test.expected {
			if instruction[i] != test.expected[i] {
				t.Errorf("wrong byte at position %d. got %d, expected %d", i, instruction[i], b)
			}
		}
	}
}

func TestInstructionsString(t *testing.T) {
	instructions := []Instructions{
		Make(OpAdd),
		Make(OpGetLocal, 1),
		Make(OpConstant, 2),
		Make(OpConstant, 65535),
	}

	expected := `0000 OpAdd
0001 OpGetLocal 1
0003 OpConstant 2
0006 OpConstant 65535
`

	concatInst := Instructions{}
	for _, ins := range instructions {
		concatInst = append(concatInst, ins...)
	}
	if concatInst.String() != expected {
		t.Errorf("instructions wrongly formatted.\nexpected %q\ngot\t %q", expected, concatInst.String())
	}
}

func TestReadOperands(t *testing.T) {
	tests := []struct {
		op       Opcode
		operands []int
		byteRead int
	}{
		{
			OpConstant,
			[]int{65535},
			2,
		},
		{
			OpGetLocal,
			[]int{255},
			1,
		},
	}

	for _, test := range tests {
		instruction := Make(test.op, test.operands...)
		def, err := Lookup(byte(test.op))
		if err != nil {
			t.Fatalf("definition not found: %q", err)
		}
		operandsRead, n := ReadOperands(def, instruction[1:])
		if n != test.byteRead {
			t.Fatalf("wrong number of bytes expected %d, got %d", test.byteRead, n)
		}
		for i, expected := range test.operands {
			if operandsRead[i] != expected {
				t.Errorf("operand wrong. expected %d, got %d", expected, operandsRead[i])
			}
		}
	}
}
