package vm

import (
	"monkey/code"
	"monkey/object"
)

type Frame struct {
	fn *object.CompiledFunction
	pc int
}

func NewFrame(fn *object.CompiledFunction) *Frame {
	return &Frame{
		fn: fn,
		pc: -1,
	}
}

func (f *Frame) Instructions() code.Instructions {
	return f.fn.Instructions
}
