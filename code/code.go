package code

import (
	"bytes"
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

	OpAdd
	OpSub
	OpMul
	OpDiv

	OpPop

	OpTrue
	OpFalse
	OpEqual
	OpNotEqual
	OpGreaterThan

	OpPrefixMinus
	OpBang

	OpJumpNotTruthy
	OpJump

	OpNull

	OpGetGlobal
	OpSetGlobal
	OpGetLocal
	OpSetLocal

	OpArray
	OpHash

	// OpIndex is index operation for array or hash
	OpIndex

	OpCall
	OpReturnValue
	OpReturn

	OpGetBuiltin

	OpClosure
	OpGetFree
)

// Definition 其实主要用于取操作数
type Definition struct {
	Name string
	// 1. len(OperandWidth) ==> operand count
	// 2. OperandWidth[i] ==> operand width(size)
	OperandWidth []int
}

// 指令对应的操作数信息
var definitions = map[OpCode]*Definition{
	OpConstant: &Definition{
		Name:         "OpConstant",
		OperandWidth: []int{2},
	},
	OpAdd: &Definition{
		Name:         "OpAdd",
		OperandWidth: []int{},
	},
	OpSub: &Definition{
		Name:         "OpSub",
		OperandWidth: []int{},
	},
	OpMul: &Definition{
		Name:         "OpMul",
		OperandWidth: []int{},
	},
	OpDiv: &Definition{
		Name:         "OpDiv",
		OperandWidth: []int{},
	},
	OpPop: &Definition{
		Name:         "OpPop",
		OperandWidth: []int{},
	},
	OpTrue: &Definition{
		Name:         "OpTrue",
		OperandWidth: []int{},
	},
	OpFalse: &Definition{
		Name:         "OpFalse",
		OperandWidth: []int{},
	},

	OpEqual: &Definition{
		Name:         "OpEqual",
		OperandWidth: []int{},
	},
	OpNotEqual: &Definition{
		Name:         "OpNotEqual",
		OperandWidth: []int{},
	},
	OpGreaterThan: &Definition{
		Name:         "OpGreaterThan",
		OperandWidth: []int{},
	},

	OpPrefixMinus: &Definition{
		Name:         "OpPrefixMinus",
		OperandWidth: []int{},
	},
	OpBang: &Definition{
		Name:         "OpBang",
		OperandWidth: []int{},
	},
	OpJumpNotTruthy: &Definition{
		Name:         "OpJumpNotTruthy",
		OperandWidth: []int{2},
	},
	OpJump: &Definition{
		Name:         "OpJump",
		OperandWidth: []int{2},
	},
	OpNull: &Definition{
		Name:         "OpNull",
		OperandWidth: []int{},
	},
	OpSetGlobal: &Definition{
		Name:         "OpSetGlobal",
		OperandWidth: []int{2},
	},
	OpGetGlobal: &Definition{
		Name:         "OpGetGlobal",
		OperandWidth: []int{2},
	},
	OpArray: &Definition{
		Name:         "OpArray",
		OperandWidth: []int{2},
	},
	OpHash: &Definition{
		Name:         "OpHash",
		OperandWidth: []int{2},
	},
	OpIndex: &Definition{
		Name:         "OpIndex",
		OperandWidth: []int{},
	},
	OpCall: &Definition{
		Name:         "OpCall",
		OperandWidth: []int{1},
	},
	OpReturnValue: &Definition{
		Name:         "OpReturnValue",
		OperandWidth: []int{},
	},
	OpReturn: &Definition{
		Name:         "OpReturn",
		OperandWidth: []int{},
	},
	OpGetLocal: &Definition{
		Name:         "OpGetLocal",
		OperandWidth: []int{1},
	},
	OpSetLocal: &Definition{
		Name: "OpSetLocal",
		// local变量最多256个哦
		OperandWidth: []int{1},
	},
	OpGetBuiltin: &Definition{
		Name: "OpGetBuiltin",
		// 内建函数就那几个
		OperandWidth: []int{1},
	},
	OpClosure: &Definition{
		Name: "OpClosure",
		// 1. constant index: real function
		// 2. free variable count
		OperandWidth: []int{2, 1},
	},
	OpGetFree: &Definition{
		Name:         "OpGetFree",
		OperandWidth: []int{1},
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
	// operator
	instruction[0] = byte(op)
	offset := 1
	for i, o := range operands {
		// 第i个操作数定义的宽度
		width := def.OperandWidth[i]
		switch width {
		case 2:
			binary.BigEndian.PutUint16(instruction[offset:], uint16(o))
		case 1:
			instruction[offset] = byte(o)
		}
		offset += width
	}
	return instruction
}

func (ins Instructions) String() string {
	var out bytes.Buffer
	i := 0
	for i < len(ins) {
		def, err := Lookup(ins[i])
		if err != nil {
			fmt.Fprintf(&out, "ERROR: %s\n", err)
			continue
		}
		// jump the first operator
		operands, read := ReadOperands(def, ins[i+1:])
		fmt.Fprintf(&out, "%04d %s\n", i, ins.fmtInstruction(def, operands))
		// also count the operator
		i += 1 + read
	}
	return out.String()
}

func (ins Instructions) fmtInstruction(def *Definition, operands []int) string {
	operandCount := len(def.OperandWidth)
	if len(operands) != operandCount {
		return fmt.Sprintf("ERROR: operand len %d does not match defined %d\n", len(operands), operandCount)
	}
	switch operandCount {
	case 0:
		return def.Name
	case 1:
		return fmt.Sprintf("%s %d", def.Name, operands[0])
	case 2:
		return fmt.Sprintf("%s %d %d", def.Name, operands[0], operands[1])
	}
	return fmt.Sprintf("ERROR: unhandled operandCount for %s\n", def.Name)
}

// ReadOperands read all operands，注意ins已经把对应的前面操作符略过了(+1了)
func ReadOperands(def *Definition, ins Instructions) ([]int, int) {
	operands := make([]int, len(def.OperandWidth))
	offset := 0
	for i, width := range def.OperandWidth {
		switch width {
		case 2:
			operands[i] = int(ReadUint16(ins[offset:]))
		case 1:
			operands[i] = int(ReadUint8(ins[offset:]))
		}
		offset += width
	}
	return operands, offset
}

func ReadUint16(ins Instructions) uint16 {
	return binary.BigEndian.Uint16(ins)
}

func ReadUint8(ins Instructions) uint8 {
	return uint8(ins[0])
}
