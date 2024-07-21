package zmachine

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"os"
	"strings"
)

type ZMachine struct {
	Counter       Address
	Header        Header
	Memory        []uint8
	DynamicMemory []uint8
	StaticMemory  []uint8
	HighMemory    []uint8
	Stack         []uint16
}

type Instruction struct {
	Opcode       uint8
	Operands     []Operand
	Store        uint8
	BranchOffset uint16
	Text         string
}

func (instruction Instruction) String() string {
	builder := strings.Builder{}
	builder.WriteString(fmt.Sprintf("%x", instruction.Opcode))
	builder.WriteString(" - ")

	operand_strings := make([]string, 0, len(instruction.Operands))
	for _, operand := range instruction.Operands {
		operand_strings = append(operand_strings, fmt.Sprintf("$%x", operand.Value))
	}

	return fmt.Sprintf("%x - %s - b: %x", instruction.Opcode, strings.Join(operand_strings, " "), instruction.BranchOffset)
}

type Operand struct {
	Type  OpType
	Value uint16
}

type OpcodeInfo struct {
	Name   string
	Store  bool
	Branch bool
	Text   bool
}

type OpType uint8
type Address uint16

const (
	OPTYPE_LargeConst OpType = 0 // 2 bytes
	OPTYPE_SmallConst OpType = 1 // 1 byte
	OPTYPE_Variable   OpType = 2 // 1 byte
	OPTYPE_Omitted    OpType = 3 // 0 byte
)

var opcodes = map[uint8]OpcodeInfo{
	0xe0: {"Call", true, false, false}, // call, TODO: In V4 Store should equal false
}

const FLAGS1_None uint8 = 0
const (
	_                 uint8 = 1 << iota
	FLAGS1_StatusLine       // 0 = score/turns, 1 = hours:mins
	FLAGS1_SplitStory
	FLAGS1_Tandy
	FLAGS1_StatusUnavailable
	FLAGS1_ScreenSplit
	FLAGS1_VariableFont
	_
)

const (
	FLAGS1_Colors uint8 = 1 << iota
	FLAGS1_Pictures
	FLAGS1_Boldface
	FLAGS1_Italics
	FLAGS1_Monospace
	FLAGS1_Sounds
	_
	FLAGS1_Timed
)

const FLAGS2_None uint16 = 0
const (
	FLAGS2_Transcript uint16 = 1 << iota
	FLAGS2_ForceMono
	FLAGS2_Redraw
	FLAGS2_Pictures
	FLAGS2_Undo // FLAGS2_TLHSounds
	FLAGS2_Mouse
	FLAGS2_Color
	FLAGS2_Sounds
	FLAGS2_Menus
	_
	FLAGS2_PrintError
	_
	_
	_
	_
	_
)

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

func (zmachine *ZMachine) init(memory []uint8) {
	zmachine.Memory = memory

	header := Header{}
	err := binary.Read(bytes.NewBuffer(zmachine.Memory[0:64]), binary.BigEndian, &header)
	if err != nil {
		panic(err)
	}

	zmachine.Header = header

	zmachine.Counter = header.InitialProgramCounter
	zmachine.DynamicMemory = memory[:header.StaticMemoryAddr]
	zmachine.StaticMemory = memory[header.StaticMemoryAddr:min(len(memory), 0xffff)]
	zmachine.HighMemory = memory[header.HighMemoryAddr:]
}

func Load(story_path string) (*ZMachine, error) {
	zmachine := ZMachine{}

	memory, err := os.ReadFile(story_path)
	if err != nil {
		panic(err)
	}

	zmachine.init(memory)
	return &zmachine, nil
}

func (zmachine ZMachine) Run() {
	for {
		zmachine.execute_next_instruction()
	}
}

func (zmachine ZMachine) read_byte(address Address) (uint8, Address) {
	return zmachine.Memory[address], address + 1
}

func (zmachine ZMachine) read_word(address Address) (uint16, Address) {
	byte1 := zmachine.Memory[address]
	byte2 := zmachine.Memory[address+1]
	return (uint16(byte1) << 8) | uint16(byte2), address + 2
}

func (zmachine ZMachine) get_routine_address(address Address) Address {
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

func is_variable_instruction(opcode uint8) bool {
	return opcode>>6 == 0b11
}

func is_short_instruction(opcode uint8) bool {
	return opcode>>6 == 0b10
}

func (zmachine ZMachine) parse_short_instruction(address Address) (Instruction, Address) {
	opcode, next_address := zmachine.read_byte(address)
	instruction := Instruction{Opcode: opcode}

	optype := OpType((opcode & 0b00110000) >> 4)
	var operand_value uint16
	if optype != OPTYPE_Omitted {
		// 1OP
		if optype == OPTYPE_LargeConst {
			operand_value, next_address = zmachine.read_word(next_address)
		} else {
			var byte uint8
			byte, next_address = zmachine.read_byte(next_address)
			operand_value = uint16(byte)
		}

		instruction.Operands = append(instruction.Operands, Operand{
			Type:  optype,
			Value: operand_value,
		})
	} else {
		// 0OP
	}

	opinfo, ok := opcodes[opcode]
	if !ok {
		panic(fmt.Sprintf("Unknown opcode: %x", opcode))
	}

	if opinfo.Store {
		var store uint8
		store, next_address = zmachine.read_byte(next_address)
		instruction.Store = store
	}

	// if opinfo.Branch {
	// 	// TODO: Branches
	// }

	// if opinfo.Text {
	// 	// TODO: Text
	// }

	return instruction, next_address
}

func (zmachine ZMachine) parse_long_instruction(address Address) (Instruction, Address) {
	opcode, next_address := zmachine.read_byte(address)
	instruction := Instruction{Opcode: opcode}

	optype1 := OPTYPE_SmallConst
	if (opcode&0b01000000)>>6 == 1 {
		optype1 = OPTYPE_Variable
	}

	var byte uint8
	byte, next_address = zmachine.read_byte(next_address)
	instruction.Operands = append(instruction.Operands, Operand{
		Type:  optype1,
		Value: uint16(byte),
	})

	optype2 := OPTYPE_SmallConst
	if (opcode&0b00100000)>>5 == 1 {
		optype2 = OPTYPE_Variable
	}

	byte, next_address = zmachine.read_byte(next_address)
	instruction.Operands = append(instruction.Operands, Operand{
		Type:  optype2,
		Value: uint16(byte),
	})

	opinfo, ok := opcodes[opcode]
	if !ok {
		panic(fmt.Sprintf("Unknown opcode: %x", opcode))
	}

	if opinfo.Store {
		var store uint8
		store, next_address = zmachine.read_byte(next_address)
		instruction.Store = store
	}

	// if opinfo.Branch {
	// 	// TODO: Branches
	// }

	// if opinfo.Text {
	// 	// TODO: Text
	// }

	return instruction, next_address
}

func (zmachine ZMachine) parse_variable_instruction(address Address) (Instruction, Address) {
	opcode, next_address := zmachine.read_byte(address)
	instruction := Instruction{Opcode: opcode}

	// if (opcode&0b00100000)>>5 == 1 {
	var byte uint8
	byte, next_address = zmachine.read_byte(next_address)

	optype4 := OpType(byte & 0b11)
	optype3 := OpType((byte >> 2) & 0b11)
	optype2 := OpType((byte >> 4) & 0b11)
	optype1 := OpType((byte >> 6) & 0b11)

	// TODO: Version 4 and 5 support call_vn2 and call_vs2 which read a second type byte
	types := []OpType{optype1, optype2, optype3, optype4}
	for _, optype := range types {
		if optype == OPTYPE_Omitted {
			break
		}

		operand := Operand{Type: optype}

		var operand_value uint16
		if optype == OPTYPE_LargeConst {
			operand_value, next_address = zmachine.read_word(next_address)
		} else {
			byte, next_address = zmachine.read_byte(next_address)
			operand_value = uint16(byte)
		}

		operand.Value = operand_value
		instruction.Operands = append(instruction.Operands, operand)
	}
	// }

	opinfo, ok := opcodes[opcode]
	if !ok {
		panic(fmt.Sprintf("Unknown opcode: %x", opcode))
	}

	if opinfo.Store {
		var store uint8
		store, next_address = zmachine.read_byte(next_address)
		instruction.Store = store
	}

	// if opinfo.Branch {
	// 	// TODO: Branches
	// }

	// if opinfo.Text {
	// 	// TODO: Text
	// }

	return instruction, next_address
}

func is_extended_instruction(opcode uint8, version uint8) bool {
	return opcode == 0xBE && version >= 5
}

func (zmachine *ZMachine) execute_next_instruction() {
	opcode, _ := zmachine.read_byte(zmachine.Counter)

	var instruction Instruction
	var next_address Address
	if is_short_instruction(opcode) {
		instruction, next_address = zmachine.parse_short_instruction(zmachine.Counter)
	} else if is_variable_instruction(opcode) {
		instruction, next_address = zmachine.parse_variable_instruction(zmachine.Counter)
	} else if is_extended_instruction(opcode, zmachine.Header.Version) {
		// 	opcode, err = buffer.ReadByte()
		// 	if err != nil {
		// 		panic(err)
		// 	}

		//     return parse_extended_instruction(opcode, buffer)
	} else {
		instruction, next_address = zmachine.parse_long_instruction(zmachine.Counter)
	}

	fmt.Println(instruction)
	// TODO: execute instruction
	zmachine.Counter = next_address
}

// func (zmachine ZMachine) call(instruction Instruction) {
// 	routineAddr := zmachine.get_routine_address((Address)(instruction.Operands[0].Value))
// 	byte, next_address := zmachine.read_byte(routineAddr)

// }
