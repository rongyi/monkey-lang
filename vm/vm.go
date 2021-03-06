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
	// GlobalSize is global variable array
	GlobalSize = 65536
	// MaxFrames is call frame
	MaxFrames = 1024
)

var (
	// True is the global true object
	True = &object.Boolean{Value: true}
	// False is the global false object
	False = &object.Boolean{Value: false}
	// Null is nil
	Null = &object.Null{}
)

type VM struct {
	constants []object.Object
	// instructions code.Instructions

	stack   []object.Object
	sp      int // always point to next available value, top is stack[sp - 1]
	globals []object.Object

	frames     []*Frame
	frameIndex int // point to next available
}

func (vm *VM) currentFrame() *Frame {
	return vm.frames[vm.frameIndex-1]
}

func (vm *VM) pushFrame(f *Frame) {
	vm.frames[vm.frameIndex] = f
	vm.frameIndex++
}

func (vm *VM) popFrame() *Frame {
	vm.frameIndex--
	return vm.frames[vm.frameIndex]
}

func NewWithGlobalStore(bytecode *compiler.Bytecode, s []object.Object) *VM {
	vm := New(bytecode)
	vm.globals = s

	return vm
}

func New(bytecode *compiler.Bytecode) *VM {
	mainFn := &object.CompiledFunction{
		Instructions: bytecode.Instructions,
	}
	mainClosure := &object.Closure{Fn: mainFn}
	mainFrame := NewFrame(mainClosure, 0)

	frames := make([]*Frame, MaxFrames)
	frames[0] = mainFrame

	return &VM{
		constants: bytecode.Constants,

		stack: make([]object.Object, StackSize),
		sp:    0,

		globals: make([]object.Object, GlobalSize),

		frames:     frames,
		frameIndex: 1,
	}
}

func (vm *VM) StackTop() object.Object {
	if vm.sp == 0 {
		return nil
	}

	return vm.stack[vm.sp-1]
}

func (vm *VM) pop() object.Object {
	// TODO: make it more safe
	o := vm.stack[vm.sp-1]
	vm.sp--

	return o
}

func (vm *VM) Run() error {
	var pc int
	var ins code.Instructions
	var op code.OpCode

	for vm.currentFrame().pc < len(vm.currentFrame().Instructions())-1 {
		vm.currentFrame().pc++

		pc = vm.currentFrame().pc
		ins = vm.currentFrame().Instructions()
		op = code.OpCode(ins[pc])
		switch op {
		case code.OpConstant:
			constIndex := code.ReadUint16(ins[pc+1:])
			vm.currentFrame().pc += 2

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
		case code.OpEqual, code.OpNotEqual, code.OpGreaterThan:
			err := vm.executeComparison(op)
			if err != nil {
				return err
			}
		case code.OpBang:
			err := vm.executeBangOperator()
			if err != nil {
				return err
			}
		case code.OpPrefixMinus:
			err := vm.executeMinuxOperator()
			if err != nil {
				return err
			}
		case code.OpJump:
			pos := int(code.ReadUint16(ins[pc+1:]))
			// 因为for循环里的pc++
			vm.currentFrame().pc = pos - 1
		case code.OpJumpNotTruthy:
			pos := int(code.ReadUint16(ins[pc+1:]))
			// 这个指令长度是3，其实应该是加3，循环里会加1，所以这里少加一个
			vm.currentFrame().pc += 2
			// 这里弹出if条件里的那个值
			condition := vm.pop()
			// not truth 就跳呀
			if !isTruthy(condition) {
				// 减1和上面OpJump少1一个意思
				vm.currentFrame().pc = pos - 1
			}
		case code.OpNull:
			err := vm.push(Null)
			if err != nil {
				return err
			}
		case code.OpSetGlobal:
			globalIndex := code.ReadUint16(ins[pc+1:])
			vm.currentFrame().pc += 2
			vm.globals[globalIndex] = vm.pop()
		case code.OpGetGlobal:
			globalIndex := code.ReadUint16(ins[pc+1:])
			vm.currentFrame().pc += 2
			err := vm.push(vm.globals[globalIndex])
			if err != nil {
				return err
			}
		case code.OpArray:
			numsElements := int(code.ReadUint16(ins[pc+1:]))
			vm.currentFrame().pc += 2
			array := vm.buildArray(vm.sp-numsElements, vm.sp)
			vm.sp = vm.sp - numsElements

			err := vm.push(array)
			if err != nil {
				return err
			}
		case code.OpHash:
			numElements := int(code.ReadUint16(ins[pc+1:]))
			vm.currentFrame().pc += 2

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
		case code.OpCall:
			numArgs := code.ReadUint8(ins[pc+1:])
			// ignore the len(arg) in this instruction
			vm.currentFrame().pc++

			// fn, ok := vm.stack[vm.sp-1-int(numArgs)].(*object.CompiledFunction)
			// if !ok {
			// 	return fmt.Errorf("calling non-function")
			// }
			// // 注意这里如果没有函数参数时是对的
			// // TODO: if there are function args, this need be changed
			// frame := NewFrame(fn, vm.sp)
			// vm.pushFrame(frame)
			// vm.sp = frame.basePointer + fn.NumLocals

			// err := vm.callFunction(int(numArgs))
			// if err != nil {
			// 	return err
			// }
			err := vm.executeCall(int(numArgs))
			if err != nil {
				return err
			}
		case code.OpReturnValue:
			returnValue := vm.pop()
			frame := vm.popFrame()
			// 减一是把这个调用的函数本身也去掉
			vm.sp = frame.basePointer - 1
			// pop the function instruction on the stack
			// vm.pop()
			// push result
			err := vm.push(returnValue)
			if err != nil {
				return err
			}
		case code.OpReturn:
			frame := vm.popFrame()
			// 减一是把这个被调用的函数本身也去掉
			vm.sp = frame.basePointer - 1
			// this function
			// vm.pop()

			err := vm.push(Null)
			if err != nil {
				return err
			}
		case code.OpSetLocal:
			localIndex := code.ReadUint8(ins[pc+1:])
			// 加上操作数长度，然后for循环还会加一
			vm.currentFrame().pc++

			frame := vm.currentFrame()
			// 这里不是stack的玩法，是数组的玩法!!
			vm.stack[frame.basePointer+int(localIndex)] = vm.pop()
		case code.OpGetLocal:
			localIndex := code.ReadUint8(ins[pc+1:])
			// 加上操作数长度，然后for循环还会加一
			vm.currentFrame().pc++
			frame := vm.currentFrame()
			// 数组玩法
			err := vm.push(vm.stack[frame.basePointer+int(localIndex)])
			if err != nil {
				return err
			}
		case code.OpGetBuiltin:
			builtinIndex := code.ReadUint8(ins[pc+1:])

			vm.currentFrame().pc++

			definition := object.Builtins[builtinIndex]
			err := vm.push(definition.Builtin)
			if err != nil {
				return err
			}
		case code.OpClosure:
			constIndex := code.ReadUint16(ins[pc+1:])
			numFree := code.ReadUint8(ins[pc+3:])
			// 其实只需要加上操作数长度即可，本身指令那一个字节在循环头部加上了，一再强调这里
			vm.currentFrame().pc += 3

			err := vm.pushClosure(int(constIndex), int(numFree))
			if err != nil {
				return err
			}
		case code.OpGetFree:
			freeIndex := code.ReadUint8(ins[pc+1:])
			vm.currentFrame().pc++

			currentClosure := vm.currentFrame().cl
			err := vm.push(currentClosure.Free[freeIndex])
			if err != nil {
				return err
			}
		}
	}
	return nil
}

func (vm *VM) pushClosure(constIndex, numFree int) error {
	constant := vm.constants[constIndex]
	fn, ok := constant.(*object.CompiledFunction)
	if !ok {
		return fmt.Errorf("not a function: %+v", constant)
	}
	free := make([]object.Object, numFree)
	for i := 0; i < numFree; i++ {
		free[i] = vm.stack[vm.sp-numFree+i]
	}

	vm.sp = vm.sp - numFree

	closure := &object.Closure{Fn: fn, Free: free}

	return vm.push(closure)
}

func (vm *VM) executeCall(numArgs int) error {
	callee := vm.stack[vm.sp-1-numArgs]
	switch callee := callee.(type) {
	case *object.Closure:
		return vm.callClosure(callee, numArgs)
	case *object.Builtin:
		return vm.callBuiltin(callee, numArgs)
	default:
		return fmt.Errorf("calling non-function and non-built-in")
	}
}

func (vm *VM) callBuiltin(builtin *object.Builtin, numArgs int) error {
	args := vm.stack[vm.sp-numArgs : vm.sp]
	ret := builtin.Fn(args...)

	vm.sp = vm.sp - numArgs - 1

	if ret != nil {
		vm.push(ret)
	} else {
		vm.push(Null)
	}

	return nil
}

func (vm *VM) callClosure(cl *object.Closure, numArgs int) error {
	// sp总是指向下一个可用的地方，所以减一
	// fn, ok := vm.stack[vm.sp-1-numArgs].(*object.CompiledFunction)
	// if !ok {
	// 	return fmt.Errorf("calling non-function")
	// }
	if numArgs != cl.Fn.NumParameters {
		return fmt.Errorf("wrong number of arguments: want=%d, got=%d", cl.Fn.NumParameters, numArgs)
	}

	frame := NewFrame(cl, vm.sp-numArgs)
	vm.pushFrame(frame)

	// this hole for local binding
	vm.sp = frame.basePointer + cl.Fn.NumLocals

	return nil
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

func (vm *VM) executeArrayIndex(array, index object.Object) error {
	arrayObject := array.(*object.Array)
	i := index.(*object.Integer).Value
	max := int64(len(arrayObject.Elements) - 1)
	if i < 0 || i > max {
		return vm.push(Null)
	}
	return vm.push(arrayObject.Elements[i])
}

func (vm *VM) executeHashIndex(hash, index object.Object) error {
	hashObject := hash.(*object.Hash)
	key, ok := index.(object.Hashable)
	if !ok {
		return fmt.Errorf("unusuable as hash key: %s", index.Type())
	}
	pair, ok := hashObject.Pairs[key.HashKey()]
	if !ok {
		return vm.push(Null)
	}
	return vm.push(pair.Value)
}

// buildHash build a map from stack, 这里的index和 buildArray一样，可以看下那个函数的说明
func (vm *VM) buildHash(startIndex, endIndex int) (object.Object, error) {
	hashedPairs := make(map[object.HashKey]object.HashPair)
	for i := startIndex; i < endIndex; i += 2 {
		key := vm.stack[i]
		value := vm.stack[i+1]

		pair := object.HashPair{Key: key, Value: value}

		hashKey, ok := key.(object.Hashable)
		if !ok {
			return nil, fmt.Errorf("unusuable as hash key: %s", key.Type())
		}
		hashedPairs[hashKey.HashKey()] = pair
	}
	return &object.Hash{Pairs: hashedPairs}, nil
}

// buildArray build array, [startIndex, endIndex)
func (vm *VM) buildArray(startIndex, endIndex int) object.Object {
	elements := make([]object.Object, endIndex-startIndex)

	// 从小于这里可以看出 endIndex 是不包含进来的，这个 vm的sp总是指向下一个可用的地方一致
	// 注意这个细节
	for i := startIndex; i < endIndex; i++ {
		elements[i-startIndex] = vm.stack[i]
	}

	return &object.Array{
		Elements: elements,
	}
}

func isTruthy(obj object.Object) bool {
	switch obj := obj.(type) {
	case *object.Boolean:
		return obj.Value
	case *object.Null:
		return false
	default:
		return true
	}
}

func (vm *VM) executeMinuxOperator() error {
	operand := vm.pop()
	if operand.Type() != object.INTEGER_OBJ {
		return fmt.Errorf("unsupported type for negation: %s", operand.Type())
	}
	value := operand.(*object.Integer).Value

	return vm.push(&object.Integer{Value: -value})
}

func (vm *VM) executeBangOperator() error {
	operand := vm.pop()
	switch operand {
	case True:
		return vm.push(False)
	case False:
		return vm.push(True)
	case Null:
		return vm.push(True)
	default:
		return vm.push(False)
	}
}

func (vm *VM) executeComparison(op code.OpCode) error {
	right := vm.pop()
	left := vm.pop()

	if left.Type() == object.INTEGER_OBJ || right.Type() == object.INTEGER_OBJ {
		return vm.executeIntegerComparison(op, left, right)
	}
	switch op {
	case code.OpEqual:
		return vm.push(nativeBoolToBooleanObject(right == left))
	case code.OpNotEqual:
		return vm.push(nativeBoolToBooleanObject(right != left))
	default:
		return fmt.Errorf("unknown operator: %d (%s %s)", op, left.Type(), right.Type())
	}
}

func (vm *VM) executeIntegerComparison(op code.OpCode, left, right object.Object) error {
	leftValue := left.(*object.Integer).Value
	rightValue := right.(*object.Integer).Value

	switch op {
	case code.OpEqual:
		return vm.push(nativeBoolToBooleanObject(rightValue == leftValue))
	case code.OpNotEqual:
		return vm.push(nativeBoolToBooleanObject(rightValue != leftValue))
	case code.OpGreaterThan:
		return vm.push(nativeBoolToBooleanObject(leftValue > rightValue))
	default:
		return fmt.Errorf("unknown operator: %d", op)
	}
}

func nativeBoolToBooleanObject(input bool) *object.Boolean {
	if input {
		return True
	}
	return False
}

func (vm *VM) executeBinaryOperation(op code.OpCode) error {
	right := vm.pop()
	left := vm.pop()

	leftType := left.Type()
	rightType := right.Type()
	switch {
	case leftType == object.INTEGER_OBJ && rightType == object.INTEGER_OBJ:
		return vm.executeBinaryIntegerOperation(op, left, right)
	case leftType == object.STRING_OBJ && rightType == object.STRING_OBJ:
		return vm.executeBinaryStringOperation(op, left, right)
	default:
		return fmt.Errorf("unsupported types for binary operation: %s %s", leftType, rightType)
	}
}

func (vm *VM) executeBinaryStringOperation(op code.OpCode, left, right object.Object) error {
	if op != code.OpAdd {
		return fmt.Errorf("unknown string operation: %d(string only support concatenation)", op)
	}
	leftValue := left.(*object.String).Value
	rightValue := right.(*object.String).Value
	return vm.push(&object.String{Value: leftValue + rightValue})
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
