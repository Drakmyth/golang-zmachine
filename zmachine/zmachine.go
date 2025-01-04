package zmachine

import (
	"fmt"

	"github.com/Drakmyth/golang-zmachine/memory"
	"github.com/Drakmyth/golang-zmachine/screen"
	"github.com/Drakmyth/golang-zmachine/stack"
)

type word = uint16

type ZMachine struct {
	Debug  bool
	Memory *memory.Memory
	Stack  stack.Stack[Frame]
}

type Frame struct {
	Counter        memory.Address
	Stack          stack.Stack[word]
	Locals         []word
	DiscardReturn  bool
	ReturnVariable Variable
}

func (zmachine *ZMachine) NewFrame() Frame {
	return Frame{ReturnVariable: zmachine.getVariable(0)}
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

	zmachine := ZMachine{
		Memory: memory,
		Stack:  stack,
	}

	return &zmachine, nil
}

func (zmachine ZMachine) Run() error {
	s := screen.NewScreen()
	for {
		zmachine.executeNextInstruction(s)

		// for ev := range s.Events {
		// 	handleEvent(ev, s)
		// }
	}
}

func (zmachine *ZMachine) executeNextInstruction(screen *screen.Screen) {
	frame := zmachine.Stack.Peek()

	instruction, next_address := zmachine.readInstruction(frame.Counter)

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

	counter_updated := instruction.Handler(zmachine, instruction, screen)

	if !counter_updated {
		frame.Counter = next_address
	}
}
