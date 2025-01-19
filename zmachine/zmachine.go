package zmachine

import (
	"bytes"
	"fmt"

	"github.com/Drakmyth/golang-zmachine/assert"
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
	frame, err := zmachine.Stack.Pop()
	assert.NoError(err, "Error popping frame stack")

	if !frame.DiscardReturn {
		frame.ReturnVariable.Write(value)
	}
}

func Load(story_path string) (*ZMachine, error) {
	m, err := memory.NewMemoryFromFile(story_path, func(m *memory.Memory) {
		// TODO: Initialize IROM
	})
	assert.NoError(err, "Error loading story")

	stack := append(make([]Frame, 0, 1024), Frame{Counter: m.GetInitialProgramCounter()})

	version := m.GetVersion()

	alphabetAddress := memory.Address(m.ReadWord(memory.Addr_ROM_A_AlphabetTable))

	ctrlchars := zstring.GetDefaultCtrlCharMapping(version)
	var charset zstring.Charset
	if alphabetAddress == 0 {
		alphabet := zstring.GetDefaultAlphabet(m.GetVersion())
		charset, err = zstring.NewStaticCharset(alphabet, ctrlchars)
		assert.NoError(err, "Error instantiating static charset")
	} else {
		alphabetHandler := func() []rune { return bytes.Runes(m.GetBytes(memory.Addr_ROM_A_AlphabetTable, 78)) }
		charset, err = zstring.NewDynamicCharset(alphabetHandler, ctrlchars)
		assert.NoError(err, "Error instantiating dynamic charset")
	}

	zmachine := ZMachine{
		Memory:  m,
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
	frame, err := zmachine.Stack.Peek()
	assert.NoError(err, "Error peeking instruction stack")

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

func (zmachine ZMachine) GetObjectShortName(o Object) string {
	parser := zstring.NewParser(zmachine.Charset, zmachine.Memory.GetAbbreviation)
	name, err := parser.Parse(o.GetShortName())
	assert.NoError(err, "Error parsing object short name")
	return name
}
