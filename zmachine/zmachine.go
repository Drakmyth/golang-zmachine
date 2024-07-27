package zmachine

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"os"
	"slices"
)

type ZMachine struct {
	Header      Header
	Memory      []byte
	StackFrames Stack[StackFrame]
}

type StackFrame struct {
	Counter        Address
	Stack          Stack[word]
	Locals         []word
	DiscardReturn  bool
	ReturnVariable VarNum
}

func (zmachine *ZMachine) endCurrentFrame(value word) {
	frame := zmachine.StackFrames[0]
	if !frame.DiscardReturn {
		zmachine.writeVariable(value, frame.ReturnVariable)
	}
	zmachine.StackFrames.pop()
}

// const FLAGS1_None uint8 = 0
// const (
// 	_                 uint8 = 1 << iota
// 	FLAGS1_StatusLine       // 0 = score/turns, 1 = hours:mins
// 	FLAGS1_SplitStory
// 	FLAGS1_Tandy
// 	FLAGS1_StatusUnavailable
// 	FLAGS1_ScreenSplit
// 	FLAGS1_VariableFont
// 	_
// )

// const (
// 	FLAGS1_Colors uint8 = 1 << iota
// 	FLAGS1_Pictures
// 	FLAGS1_Boldface
// 	FLAGS1_Italics
// 	FLAGS1_Monospace
// 	FLAGS1_Sounds
// 	_
// 	FLAGS1_Timed
// )

// const FLAGS2_None uint16 = 0
// const (
// 	FLAGS2_Transcript uint16 = 1 << iota
// 	FLAGS2_ForceMono
// 	FLAGS2_Redraw
// 	FLAGS2_Pictures
// 	FLAGS2_Undo // FLAGS2_TLHSounds
// 	FLAGS2_Mouse
// 	FLAGS2_Color
// 	FLAGS2_Sounds
// 	FLAGS2_Menus
// 	_
// 	FLAGS2_PrintError
// 	_
// 	_
// 	_
// 	_
// 	_
// )

type Header struct {
	Version               uint8
	Flags1                uint8
	ReleaseNumber         uint16
	HighMemoryAddr        Address
	InitialProgramCounter Address
	DictionaryAddr        Address
	ObjectsAddr           Address
	GlobalsAddr           Address
	StaticMemoryAddr      Address
	Flags2                uint16
	Serial                [6]byte
	AbbreviationsAddr     Address
	FileLength            uint16
	Checksum              uint16
	Interpreter           Interpreter
	Screen                Screen
	Font                  Font
	RoutinesAddr          Address
	StaticStringsAddr     Address
	BackgroundColor       uint8
	ForegroundColor       uint8
	TermCharsAddr         Address
	Stream3Width          uint16
	StandardRev           uint16
	AlphabetAddr          Address
	HeaderExtensionAddr   Address
}

type Interpreter struct {
	Number   uint8
	Revision uint8
}

type Screen struct {
	Height      uint8
	Width       uint8
	WidthUnits  uint8
	HeightUnits uint8
}

type Font struct {
	Height      uint8
	Width       uint8
	HeightUnits uint8
	WidthUnits  uint8
}

func (zmachine *ZMachine) init(memory []byte) error {
	zmachine.Memory = memory
	zmachine.StackFrames = make([]StackFrame, 0, 1024)

	header := Header{}
	err := binary.Read(bytes.NewBuffer(zmachine.Memory[0:64]), binary.BigEndian, &header)
	if err != nil {
		return err
	}

	zmachine.Header = header
	zmachine.StackFrames = append(zmachine.StackFrames, StackFrame{Counter: header.InitialProgramCounter})
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
	frame := &zmachine.StackFrames[0]
	pc := frame.Counter

	instruction, next_address, err := zmachine.readInstruction(pc)
	if err != nil {
		return err
	}

	fmt.Printf("%x: %s\n", pc, instruction)

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

// // func (zmachine ZMachine) read_zstring(address Address) {
// // 	words := make([]uint16, 0)
// // 	zstr_word, next_address := zmachine.read_word(address)
// // 	words = append(words, zstr_word)

// // 	for zstr_word>>15 == 0 {
// // 		zstr_word, next_address = zmachine.read_word(next_address)
// // 		words = append(words, zstr_word)
// // 	}

// // 	zchars := make([]uint8, 0, len(words)*3)
// // 	for _, word := range words {
// // 		zchars = append(zchars, uint8((word>>10)&0b11111))
// // 		zchars = append(zchars, uint8((word>>5)&0b11111))
// // 		zchars = append(zchars, uint8(word&0b11111))
// // 	}

// // 	print("temp")
// // }

func (zmachine ZMachine) getRoutineAddress(address Address) Address {
	switch zmachine.Header.Version {
	case 1, 2, 3:
		return address * 2
	case 4, 5:
		return address * 4
	case 6, 7:
		return address*4 + zmachine.Header.RoutinesAddr*8
	case 8:
		return address * 8
	}

	panic("Unknown version")
}

func (zmachine ZMachine) readInstruction(address Address) (Instruction, Address, error) {
	opcode, next_address := zmachine.readOpcode(address)
	inst_info, ok := opcodes[opcode]
	if !ok {
		return Instruction{}, 0, fmt.Errorf("unknown opcode: %x", opcode)
	}
	instruction := Instruction{InstructionInfo: inst_info, Opcode: opcode}

	if instruction.Form == IF_Extended {
		return Instruction{}, 0, fmt.Errorf("unimplemented: extended form instruction")
	}

	// Determine Variable Form operand types
	if instruction.Form == IF_Variable {
		var types_byte uint8
		types_byte, next_address = zmachine.readByte(next_address)

		for shift := 0; shift <= 6; shift += 2 {
			operand_type := OperandType((types_byte >> shift) & 0b11)
			if operand_type != OT_Omitted {
				instruction.OperandTypes = append(instruction.OperandTypes, operand_type)
				// Types are parsed last to first, so we need to reverse them
				slices.Reverse(instruction.OperandTypes)
			}
		}
	}

	// Parse Operands
	instruction.Operands = make([]Operand, 0, len(instruction.OperandTypes))
	for _, optype := range instruction.OperandTypes {
		var operand Operand
		operand, next_address = zmachine.readOperand(optype, next_address)
		instruction.Operands = append(instruction.Operands, operand)
	}

	if instruction.StoresResult() {
		var store_varnum VarNum
		store_varnum, next_address = zmachine.readVarNum(next_address)
		instruction.StoreVariable = store_varnum
	}

	if instruction.Branches() {
		var branch Branch
		branch, next_address = zmachine.readBranch(next_address)
		instruction.Branch = branch
	}

	// if instruction.HasText() {
	//     // TODO: Implement text handling
	// }

	return instruction, next_address, nil
}
