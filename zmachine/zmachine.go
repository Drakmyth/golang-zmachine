package zmachine

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"os"
	"reflect"
	"runtime"
	"strings"
)

type ZMachine struct {
	Header      Header
	Memory      []uint8
	StackFrames []StackFrame
}

type StackFrame struct {
	Counter Address
	Stack   []uint16
	Locals  []uint16
}

type Address uint16

type Instruction struct {
	OpcodeInfo
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

	function_path := strings.Split(runtime.FuncForPC(reflect.ValueOf(instruction.Handler).Pointer()).Name(), ".")
	operation_name := strings.ToUpper(function_path[len(function_path)-1])

	log_strings := make([]string, 0, len(instruction.Operands)+4)
	log_strings = append(log_strings, operation_name)
	for _, operand := range instruction.Operands {
		log_strings = append(log_strings, fmt.Sprintf("$%x", operand.Value))
	}

	if instruction.PerformsStore {
		log_strings = append(log_strings, fmt.Sprintf("$%x", instruction.Store))
	}

	if instruction.PerformsBranch {
		log_strings = append(log_strings, fmt.Sprintf("$%x", instruction.BranchOffset))
	}

	return strings.Join(log_strings, " ")
}

type Operand struct {
	Type  OpType
	Value uint16
}

type OpType uint8

const (
	OPTYPE_LargeConst OpType = 0 // 2 bytes
	OPTYPE_SmallConst OpType = 1 // 1 byte
	OPTYPE_Variable   OpType = 2 // 1 byte
	OPTYPE_Omitted    OpType = 3 // 0 byte
)

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
	zmachine.StackFrames = make([]StackFrame, 0, 1024)

	header := Header{}
	err := binary.Read(bytes.NewBuffer(zmachine.Memory[0:64]), binary.BigEndian, &header)
	if err != nil {
		panic(err)
	}

	zmachine.Header = header
	zmachine.StackFrames = append(zmachine.StackFrames, StackFrame{Counter: header.InitialProgramCounter})
}

func (zmachine ZMachine) CurrentFrame() *StackFrame {
	return &zmachine.StackFrames[len(zmachine.StackFrames)-1]
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

func (zmachine *ZMachine) execute_next_instruction() {
	frame := zmachine.CurrentFrame()
	pc := frame.Counter
	opcode, _ := zmachine.read_byte(pc)

	var instruction Instruction
	var next_address Address
	if is_short_instruction(opcode) {
		instruction, next_address = zmachine.parse_short_instruction(pc)
	} else if is_variable_instruction(opcode) {
		instruction, next_address = zmachine.parse_variable_instruction(pc)
	} else if is_extended_instruction(opcode, zmachine.Header.Version) {
		instruction, next_address = zmachine.parse_extended_instruction(pc)
	} else {
		instruction, next_address = zmachine.parse_long_instruction(pc)
	}

	fmt.Printf("%x: %s\n", pc, instruction)

	instruction.Handler(zmachine, instruction)
	frame.Counter = next_address
}

func (zmachine ZMachine) read_byte(address Address) (uint8, Address) {
	return zmachine.Memory[address], address + 1
}

func (zmachine *ZMachine) write_byte(value uint8, address Address) {
	zmachine.Memory[address] = value
}

func (zmachine ZMachine) read_word(address Address) (uint16, Address) {
	byte1 := zmachine.Memory[address]
	byte2 := zmachine.Memory[address+1]
	return (uint16(byte1) << 8) | uint16(byte2), address + 2
}

func (zmachine *ZMachine) write_word(value uint16, address Address) {
	byte1 := uint8(value >> 8)
	byte2 := uint8(value)
	zmachine.Memory[address] = byte1
	zmachine.Memory[address+1] = byte2
}

func (zmachine *ZMachine) read_variable(index uint8) uint16 {
	if index == 0 {
		frame := zmachine.CurrentFrame()
		// Pop from stack
		value := frame.Stack[len(frame.Stack)-1]
		frame.Stack = frame.Stack[:len(frame.Stack)-1]
		return value
	} else if index > 0 && index < 0x10 {
		// Local variable
		frame := zmachine.CurrentFrame()
		return frame.Locals[index]
	} else {
		// Global variable
		value, _ := zmachine.read_word(Address(uint16(zmachine.Header.GlobalsAddr) + uint16(index)))
		return value
	}
}

func (zmachine *ZMachine) write_variable(value uint16, index uint8) {
	if index == 0 {
		// Push to stack
		frame := zmachine.CurrentFrame()
		frame.Stack = append(frame.Stack, value)
	} else if index > 0 && index < 0x10 {
		// Local variable
		frame := zmachine.CurrentFrame()
		frame.Locals[index] = value
	} else {
		zmachine.write_word(value, Address(uint16(zmachine.Header.GlobalsAddr)+uint16(index)))
	}
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

func is_short_instruction(opcode uint8) bool {
	return opcode>>6 == 0b10
}

func (zmachine ZMachine) parse_short_instruction(address Address) (Instruction, Address) {
	opcode, next_address := zmachine.read_byte(address)
	opinfo, ok := opcodes[opcode]
	if !ok {
		panic(fmt.Sprintf("Unknown opcode: %x", opcode))
	}
	instruction := Instruction{OpcodeInfo: opinfo, Opcode: opcode}

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

	if instruction.PerformsStore {
		var store uint8
		store, next_address = zmachine.read_byte(next_address)
		instruction.Store = store
	}

	// if instruction.PerformsBranch {
	// 	// TODO: Branches
	// }

	// if instruction.HasText {
	// 	// TODO: Text
	// }

	return instruction, next_address
}

func is_variable_instruction(opcode uint8) bool {
	return opcode>>6 == 0b11
}

func (zmachine ZMachine) parse_variable_instruction(address Address) (Instruction, Address) {
	opcode, next_address := zmachine.read_byte(address)
	opinfo, ok := opcodes[opcode]
	if !ok {
		panic(fmt.Sprintf("Unknown opcode: %x", opcode))
	}
	instruction := Instruction{OpcodeInfo: opinfo, Opcode: opcode}

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

	if instruction.PerformsStore {
		var store uint8
		store, next_address = zmachine.read_byte(next_address)
		instruction.Store = store
	}

	// if instruction.PerformsBranch {
	// 	// TODO: Branches
	// }

	// if instruction.HasText {
	// 	// TODO: Text
	// }

	return instruction, next_address
}

func is_extended_instruction(opcode uint8, version uint8) bool {
	return opcode == 0xBE && version >= 5
}

func (zmachine ZMachine) parse_extended_instruction(address Address) (Instruction, Address) {
	panic("unimplemented")
}

func (zmachine ZMachine) parse_long_instruction(address Address) (Instruction, Address) {
	opcode, next_address := zmachine.read_byte(address)
	opinfo, ok := opcodes[opcode]
	if !ok {
		panic(fmt.Sprintf("Unknown opcode: %x", opcode))
	}
	instruction := Instruction{OpcodeInfo: opinfo, Opcode: opcode}

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

	if instruction.PerformsStore {
		var store uint8
		store, next_address = zmachine.read_byte(next_address)
		instruction.Store = store
	}

	// if instruction.ShouldBranch {
	// 	// TODO: Branches
	// }

	// if instruction.HasText {
	// 	// TODO: Text
	// }

	return instruction, next_address
}
