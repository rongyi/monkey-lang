package vm

import (
	"monkey/code"
	"monkey/object"
)

type Frame struct {
	// fn          *object.CompiledFunction
	cl          *object.Closure
	pc          int
	basePointer int
}

func NewFrame(cl *object.Closure, basePointer int) *Frame {
	return &Frame{
		cl:          cl,
		pc:          -1,
		basePointer: basePointer,
	}
}

func (f *Frame) Instructions() code.Instructions {
	return f.cl.Fn.Instructions
}
