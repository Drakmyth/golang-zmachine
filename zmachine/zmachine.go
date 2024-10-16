package zmachine

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"os"

	"github.com/Drakmyth/golang-zmachine/zmachine/internal/memory"
	"github.com/Drakmyth/golang-zmachine/zmachine/internal/screen"
)

type ZMachine struct {
	Debug  bool
	Header Header
	Memory []byte
	Stack  memory.Stack[Frame]
}

type Frame struct {
	Counter        memory.Address
	Stack          memory.Stack[memory.Word]
	Locals         []memory.Word
	DiscardReturn  bool
	ReturnVariable Variable
}

func (zmachine *ZMachine) NewFrame() Frame {
	return Frame{ReturnVariable: zmachine.getVariable(0)}
}

func (zmachine *ZMachine) endCurrentFrame(value memory.Word) {
	frame := zmachine.Stack.Pop()
	if !frame.DiscardReturn {
		frame.ReturnVariable.Write(value)
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
	screen.Clear()
	for {
		err := zmachine.executeNextInstruction()
		if err != nil {
			return err
		}
	}
}

func (zmachine ZMachine) readByte(address memory.Address) (byte, memory.Address) {
	return zmachine.Memory[address], address.OffsetBytes(1)
}

func (zmachine *ZMachine) writeByte(value byte, address memory.Address) {
	zmachine.Memory[address] = value
}

func (zmachine ZMachine) readWord(address memory.Address) (memory.Word, memory.Address) {
	high := memory.Word(zmachine.Memory[address])
	low := memory.Word(zmachine.Memory[address.OffsetBytes(1)])
	return (high << 8) | low, address.OffsetWords(1)
}

func (zmachine *ZMachine) writeWord(value memory.Word, address memory.Address) {
	zmachine.Memory[address] = value.HighByte()
	zmachine.Memory[address.OffsetBytes(1)] = value.LowByte()
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
