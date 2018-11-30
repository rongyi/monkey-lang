package vm

import (
	"fmt"
	"monkey/code"
	"monkey/compiler"
	"monkey/object"
)

const (
	// StackSize is the default stack size
	StackSize = 2048
)

type VM struct {
	constants    []object.Object
	instructions code.Instructions

	stack []object.Object
	sp    int // always point to next available value, top is stack[sp - 1]
}

func New(bytecode *compiler.Bytecode) *VM {
	return &VM{
		instructions: bytecode.Instructions,
		constants:    bytecode.Constants,
		stack:        make([]object.Object, StackSize),
		sp:           0,
	}
}

func (vm *VM) StackTop() object.Object {
	if vm.sp == 0 {
		return nil
	}

	return vm.stack[vm.sp-1]
}

func (vm *VM) pop() object.Object {
	o := vm.stack[vm.sp-1]
	vm.sp--

	return o
}

func (vm *VM) Run() error {
	for pc := 0; pc < len(vm.instructions); pc++ {
		op := code.OpCode(vm.instructions[pc])
		switch op {
		case code.OpConstant:
			constIndex := code.ReadUint16(vm.instructions[pc+1:])
			pc += 2

			err := vm.push(vm.constants[constIndex])
			if err != nil {
				return err
			}
		case code.OpAdd, code.OpSub, code.OpMul, code.OpDiv:
			err := vm.executeBinaryOperation(op)
			if err != nil {
				return err
			}
		case code.OpPop:
			vm.pop()
		}
	}
	return nil
}

func (vm *VM) executeBinaryOperation(op code.OpCode) error {
	right := vm.pop()
	left := vm.pop()

	leftType := left.Type()
	rightType := right.Type()
	if leftType == object.INTEGER_OBJ && rightType == object.INTEGER_OBJ {
		return vm.executeBinaryIntegerOperation(op, left, right)
	}
	return fmt.Errorf("unsupported types for binary operation: %s %s", leftType, rightType)
}

func (vm *VM) executeBinaryIntegerOperation(op code.OpCode, left, right object.Object) error {
	var ret int64
	leftValue := left.(*object.Integer).Value
	rightValue := right.(*object.Integer).Value

	switch op {
	case code.OpAdd:
		ret = leftValue + rightValue
	case code.OpSub:
		ret = leftValue - rightValue
	case code.OpMul:
		ret = leftValue * rightValue
	case code.OpDiv:
		ret = leftValue / rightValue
	default:
		return fmt.Errorf("unkown integer operation: %d", op)
	}
	return vm.push(&object.Integer{Value: ret})
}

func (vm *VM) push(o object.Object) error {
	if vm.sp >= StackSize {
		return fmt.Errorf("stack overflow")
	}
	vm.stack[vm.sp] = o
	vm.sp++

	return nil
}

func (vm *VM) LastPoppedStackElem() object.Object {
	return vm.stack[vm.sp]
}
