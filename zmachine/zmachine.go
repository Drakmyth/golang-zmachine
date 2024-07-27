package zmachine

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"os"
)

type ZMachine struct {
	Header Header
	Memory []byte
	Stack  Stack[Frame]
}

type Frame struct {
	Counter        Address
	Stack          Stack[word]
	Locals         []word
	DiscardReturn  bool
	ReturnVariable VarNum
}

func (zmachine *ZMachine) endCurrentFrame(value word) {
	frame := zmachine.Stack.pop()
	if !frame.DiscardReturn {
		zmachine.writeVariable(value, frame.ReturnVariable)
	}
}

func (zmachine *ZMachine) init(memory []byte) error {
	zmachine.Memory = memory
	zmachine.Stack = make([]Frame, 0, 1024)

	header := Header{}
	err := binary.Read(bytes.NewBuffer(zmachine.Memory[0:64]), binary.BigEndian, &header)
	if err != nil {
		return err
	}

	zmachine.Header = header
	zmachine.Stack = append(zmachine.Stack, Frame{Counter: header.InitialProgramCounter})
	return err
}

func Load(story_path string) (*ZMachine, error) {
	zmachine := ZMachine{}

	memory, err := os.ReadFile(story_path)
	if err != nil {
		return nil, err
	}

	zmachine.init(memory)
	return &zmachine, nil
}

func (zmachine ZMachine) Run() error {
	for {
		err := zmachine.executeNextInstruction()
		if err != nil {
			return err
		}
	}
}

func (zmachine *ZMachine) executeNextInstruction() error {
	frame := zmachine.Stack.peek()

	instruction, next_address, err := zmachine.readInstruction(frame.Counter)
	if err != nil {
		return err
	}

	fmt.Printf("%x: %s\n", frame.Counter, instruction)

	for i, optype := range instruction.OperandTypes {
		if optype != OT_Variable {
			continue
		}

		instruction.Operands[i] = Operand(zmachine.readVariable(VarNum(instruction.Operands[i])))
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
