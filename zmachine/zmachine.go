package zmachine

import (
	"fmt"

	"github.com/Drakmyth/golang-zmachine/memory"
	"github.com/Drakmyth/golang-zmachine/stack"
	"github.com/Drakmyth/golang-zmachine/zmachine/internal/screen"
	"github.com/Drakmyth/golang-zmachine/zstring"
)

type word = uint16

type ZMachine struct {
	Debug   bool
	Memory  *memory.Memory
	Stack   stack.Stack[Frame]
	Charset zstring.Charset
}

type Frame struct {
	Counter        memory.Address
	Stack          stack.Stack[word]
	Locals         []word
	DiscardReturn  bool
	ReturnVariable Variable
}

func (zmachine *ZMachine) endCurrentFrame(value word) {
	frame := zmachine.Stack.Pop()
	if !frame.DiscardReturn {
		frame.ReturnVariable.Write(value)
	}
}

func Load(story_path string) (*ZMachine, error) {
	memory := memory.NewMemory(story_path, func(m *memory.Memory) {
		// TODO: Initialize IROM
	})

	stack := append(make([]Frame, 0, 1024), Frame{Counter: memory.GetInitialProgramCounter()})

	version := memory.GetVersion()
	alphabet := memory.GetAlphabet()
	ctrlchars := zstring.GetDefaultCtrlCharMapping(version)
	charset := zstring.NewCharset(alphabet, ctrlchars)

	zmachine := ZMachine{
		Memory:  memory,
		Stack:   stack,
		Charset: charset,
	}

	return &zmachine, nil
}

func (zmachine ZMachine) Run() error {
	screen.Clear()
	for {
		err := zmachine.executeNextInstruction()
		if err != nil {
			return err
		}
	}
}

func (zmachine *ZMachine) executeNextInstruction() error {
	frame := zmachine.Stack.Peek()

	instruction, next_address, err := zmachine.readInstruction(frame.Counter)
	if err != nil {
		return err
	}

	if zmachine.Debug {
		fmt.Printf("%x: %s\n", frame.Counter, instruction)
	}

	for i, optype := range instruction.OperandTypes {
		if optype != OT_Variable {
			continue
		}

		variable := zmachine.getVariable(VarNum(instruction.Operands[i]))
		instruction.Operands[i] = Operand(variable.Read())
	}

	counter_updated, err := instruction.Handler(zmachine, instruction)
	if err != nil {
		return err
	}

	if !counter_updated {
		frame.Counter = next_address
	}
	return err
}

func (zmachine ZMachine) GetObjectShortName(o memory.Object) string {
	parser := zstring.NewParser(zmachine.Charset, zmachine.Memory.GetAbbreviation)
	name := parser.Parse(o.GetShortName())
	return name
}
