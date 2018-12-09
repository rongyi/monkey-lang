package vm

import (
	"monkey/code"
	"monkey/object"
)

type Frame struct {
	fn          *object.CompiledFunction
	pc          int
	basePointer int
}

func NewFrame(fn *object.CompiledFunction, basePointer int) *Frame {
	return &Frame{
		fn:          fn,
		pc:          -1,
		basePointer: basePointer,
	}
}

func (f *Frame) Instructions() code.Instructions {
	return f.fn.Instructions
}
