package compiler

import (
	"fmt"
	"monkey/ast"
	"monkey/code"
	"monkey/lexer"
	"monkey/object"
	"monkey/parser"
)

type Compiler struct {
	instructions        code.Instructions
	constants           []object.Object
	lastInstruction     EmittedInstruction
	previousInstruction EmittedInstruction
}

type EmittedInstruction struct {
	OpCode   code.OpCode
	Position int
}

func New() *Compiler {
	return &Compiler{
		instructions:        code.Instructions{},
		constants:           []object.Object{},
		lastInstruction:     EmittedInstruction{},
		previousInstruction: EmittedInstruction{},
	}
}

func (c *Compiler) Compile(node ast.Node) error {
	switch node := node.(type) {
	case *ast.Program:
		for _, s := range node.Statements {
			err := c.Compile(s)
			if err != nil {
				return err
			}
		}
	case *ast.ExpressionStatement:
		err := c.Compile(node.Expression)
		if err != nil {
			return err
		}
		c.emit(code.OpPop)
	case *ast.InfixExpression:
		if node.Operator == "<" {
			err := c.Compile(node.Right)
			if err != nil {
				return err
			}
			err = c.Compile(node.Left)
			if err != nil {
				return err
			}
			c.emit(code.OpGreaterThan)
			return nil
		}

		err := c.Compile(node.Left)
		if err != nil {
			return err
		}
		err = c.Compile(node.Right)
		if err != nil {
			return err
		}
		switch node.Operator {
		case "+":
			c.emit(code.OpAdd)
		case "-":
			c.emit(code.OpSub)
		case "*":
			c.emit(code.OpMul)
		case "/":
			c.emit(code.OpDiv)
		case ">":
			c.emit(code.OpGreaterThan)
		case "==":
			c.emit(code.OpEqual)
		case "!=":
			c.emit(code.OpNotEqual)
		default:
			return fmt.Errorf("unkown operator %s", node.Operator)
		}
	case *ast.PrefixExpression:
		err := c.Compile(node.Right)
		if err != nil {
			return err
		}
		switch node.Operator {
		case "!":
			c.emit(code.OpBang)
		case "-":
			c.emit(code.OpMinus)
		default:
			return fmt.Errorf("unknown operator %s", node.Operator)
		}
	case *ast.IntegerLiteral:
		integer := &object.Integer{Value: node.Value}
		c.emit(code.OpConstant, c.addConstant(integer))
	case *ast.Boolean:
		if node.Value {
			c.emit(code.OpTrue)
		} else {
			c.emit(code.OpFalse)
		}
	// if statement
	case *ast.IfExpression:
		err := c.Compile(node.Condition)
		if err != nil {
			return err
		}
		jumpNotTruthyPos := c.emit(code.OpJumpNotTruthy, 9999)
		err = c.Compile(node.Consequence)
		if err != nil {
			return err
		}
		if c.lastInstructionIsPop() {
			c.removeLastPop()
		}
		if node.Alternative == nil {
			afterConsequencePos := len(c.instructions)
			// awesome!
			c.changeOperand(jumpNotTruthyPos, afterConsequencePos)
		} else {				// else part
			jumpPos := c.emit(code.OpJump, 9999)
			beginAlterPos := len(c.instructions)
			c.changeOperand(jumpNotTruthyPos, beginAlterPos)
			err := c.Compile(node.Alternative)
			if err != nil {
				return err
			}
			if c.lastInstructionIsPop() {
				c.removeLastPop()
			}
			afterAlterPos := len(c.instructions)
			c.changeOperand(jumpPos, afterAlterPos)
		}
	case *ast.BlockStatement:
		for _, s := range node.Statements {
			err := c.Compile(s)
			if err != nil {
				return err
			}
		}
	}

	return nil
}

func (c *Compiler) emit(op code.OpCode, operands ...int) int {
	ins := code.Make(op, operands...)
	pos := c.addInstruction(ins)

	c.setLastInstruction(op, pos)
	return pos
}

func (c *Compiler) replaceInstructions(pos int, newInstruction []byte) {
	for i := 0; i < len(newInstruction); i++ {
		c.instructions[pos+i] = newInstruction[i]
	}
}

func (c *Compiler) changeOperand(opPos int, operand int) {
	op := code.OpCode(c.instructions[opPos])
	newInstruction := code.Make(op, operand)

	c.replaceInstructions(opPos, newInstruction)
}

func (c *Compiler) setLastInstruction(op code.OpCode, pos int) {
	previous := c.lastInstruction
	last := EmittedInstruction{OpCode: op, Position: pos}

	c.previousInstruction = previous
	c.lastInstruction = last
}

func (c *Compiler) lastInstructionIsPop() bool {
	return c.lastInstruction.OpCode == code.OpPop
}

func (c *Compiler) removeLastPop() {
	c.instructions = c.instructions[:c.lastInstruction.Position]
	c.lastInstruction = c.previousInstruction
}

func (c *Compiler) addConstant(obj object.Object) int {
	c.constants = append(c.constants, obj)
	// return the last index
	return len(c.constants) - 1
}

func (c *Compiler) addInstruction(ins []byte) int {
	posNewInstruction := len(c.instructions)
	c.instructions = append(c.instructions, ins...)
	return posNewInstruction
}

// 目前就俩快，命令的字节码，以及编译时候的constant放在一个pood里
type Bytecode struct {
	Instructions code.Instructions
	// pool里什么都放，int， string等
	Constants []object.Object
}

func (c *Compiler) Bytecode() *Bytecode {
	return &Bytecode{
		Instructions: c.instructions,
		Constants:    c.constants,
	}
}

func parse(input string) *ast.Program {
	l := lexer.New(input)
	p := parser.New(l)
	return p.ParseProgram()
}
