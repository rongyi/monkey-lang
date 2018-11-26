package code

import (
	"encoding/binary"
	"fmt"
)

type Instructions []byte

// Instruction is just []byte, they are used too often, so we
// don't want cast to/from byte slice often

// OpCode
type OpCode byte

const (
	OpConstant OpCode = iota
)

type Definition struct {
	Name string
	// 1. len(OperandWidth) ==> operand count
	// 2. OperandWidth[i] ==> operand width(size)
	OperandWidth []int
}

var definitions = map[OpCode]*Definition{
	OpConstant: &Definition{
		Name:         "OpConstant",
		OperandWidth: []int{2},
	},
}

func Lookup(op byte) (*Definition, error) {
	def, ok := definitions[OpCode(op)]
	if !ok {
		return nil, fmt.Errorf("opcode %d undefined", op)
	}
	return def, nil
}

func Make(op OpCode, operands ...int) []byte {
	def, ok := definitions[op]
	if !ok {
		return []byte{}
	}

	// operator single byte
	instructionLen := 1
	for _, w := range def.OperandWidth {
		instructionLen += w
	}

	instruction := make([]byte, instructionLen)
	instruction[0] = byte(op)
	offset := 1
	for i, o := range operands {
		// 第i个操作数定义的宽度
		width := def.OperandWidth[i]
		switch width {
		case 2:
			binary.BigEndian.PutUint16(instruction[offset:], uint16(o))
		}
		offset += width
	}
	return instruction
}
