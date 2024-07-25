package zmachine

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"os"
	"reflect"
	"runtime"
	"slices"
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

type BranchBehavior uint8

const (
	BB_Normal      BranchBehavior = 0
	BB_ReturnFalse BranchBehavior = 1
	BB_ReturnTrue  BranchBehavior = 2
)

type BranchCondition uint8

const (
	BC_OnFalse BranchCondition = 0
	BC_OnTrue  BranchCondition = 1
)

type Branch struct {
	Address   Address
	Behavior  BranchBehavior
	Condition BranchCondition
}

type Instruction struct {
	InstructionInfo
	Opcode        uint8
	OperandValues []uint16
	StoreVariable uint8
	Branch        Branch
	Text          string
}

func get_variable_string(value uint16) string {
	if value == 0 {
		return "sp"
	} else if value <= 0x10 {
		return fmt.Sprintf("local%d", value-1)
	} else {
		return fmt.Sprintf("g%d", value-0x10)
	}
}

func (instruction Instruction) String() string {
	function_path := strings.Split(runtime.FuncForPC(reflect.ValueOf(instruction.Handler).Pointer()).Name(), ".")
	operation_name := strings.ToUpper(function_path[len(function_path)-1])

	log_strings := make([]string, 0, len(instruction.OperandValues)+4)
	log_strings = append(log_strings, fmt.Sprintf("%02x %s", instruction.Opcode, operation_name))
	for i, operand := range instruction.OperandValues {
		optype := instruction.OperandTypes[i]
		if optype == OT_Variable {
			log_strings = append(log_strings, get_variable_string(operand))
		} else {
			log_strings = append(log_strings, fmt.Sprintf("%x", operand))
		}
	}

	if instruction.StoresResult() {
		log_strings = append(log_strings, fmt.Sprintf("-> %s", get_variable_string(uint16(instruction.StoreVariable))))
	}

	if instruction.Branches() {
		not := ""
		if instruction.Branch.Condition == BC_OnFalse {
			not = "~"
		}
		log_strings = append(log_strings, fmt.Sprintf("?%s%x", not, instruction.Branch.Address))
	}

	return strings.Join(log_strings, " ")
}

func (zmachine ZMachine) get_operand_value(instruction Instruction, index int) uint16 {
	if instruction.OperandTypes[index] == OT_Variable {
		return zmachine.read_variable(uint8(instruction.OperandValues[index]))
	}

	return instruction.OperandValues[index]
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

func (zmachine *ZMachine) init(memory []uint8) error {
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

func (zmachine ZMachine) CurrentFrame() *StackFrame {
	return &zmachine.StackFrames[len(zmachine.StackFrames)-1]
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
		err := zmachine.execute_next_instruction()
		if err != nil {
			return err
		}
	}
}

func (zmachine *ZMachine) execute_next_instruction() error {
	frame := zmachine.CurrentFrame()
	pc := frame.Counter

	instruction, next_address, err := zmachine.parse_instruction(pc)
	if err != nil {
		return err
	}

	fmt.Printf("%x: %s\n", pc, instruction)

	counter_updated, err := instruction.Handler(zmachine, instruction)
	if err != nil {
		return err
	}

	if !counter_updated {
		frame.Counter = next_address
	}
	return err
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
	} else if index < 0x10 {
		// Local variable
		frame := zmachine.CurrentFrame()
		return frame.Locals[index-1]
	} else {
		// Global variable
		value, _ := zmachine.read_word(Address(uint16(zmachine.Header.GlobalsAddr) + uint16(index-0x10)))
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
		frame.Locals[index-1] = value
	} else {
		zmachine.write_word(value, Address(uint16(zmachine.Header.GlobalsAddr)+uint16(index-0x10)))
	}
}

// func (zmachine ZMachine) read_zstring(address Address) {
// 	words := make([]uint16, 0)
// 	zstr_word, next_address := zmachine.read_word(address)
// 	words = append(words, zstr_word)

// 	for zstr_word>>15 == 0 {
// 		zstr_word, next_address = zmachine.read_word(next_address)
// 		words = append(words, zstr_word)
// 	}

// 	zchars := make([]uint8, 0, len(words)*3)
// 	for _, word := range words {
// 		zchars = append(zchars, uint8((word>>10)&0b11111))
// 		zchars = append(zchars, uint8((word>>5)&0b11111))
// 		zchars = append(zchars, uint8(word&0b11111))
// 	}

// 	print("temp")
// }

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

func (zmachine ZMachine) parse_instruction(address Address) (Instruction, Address, error) {
	opcode, next_address := zmachine.read_byte(address)
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
		types_byte, next_address = zmachine.read_byte(next_address)

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
	instruction.OperandValues = make([]uint16, 0, len(instruction.OperandTypes))
	for _, optype := range instruction.OperandTypes {
		var opvalue uint16
		if optype == OT_Large {
			opvalue, next_address = zmachine.read_word(next_address)
		} else {
			var value_byte uint8
			value_byte, next_address = zmachine.read_byte(next_address)
			opvalue = uint16(value_byte)
		}
		instruction.OperandValues = append(instruction.OperandValues, opvalue)
	}

	if instruction.StoresResult() {
		var store_byte uint8
		store_byte, next_address = zmachine.read_byte(next_address)
		instruction.StoreVariable = store_byte
	}

	if instruction.Branches() {
		var branch_byte uint8
		branch_byte, next_address = zmachine.read_byte(next_address)
		instruction.Branch.Behavior = BranchBehavior(branch_byte >> 7)

		var offset uint16
		offset = uint16(branch_byte & 0b00111111)
		if ((branch_byte >> 6) & 0b01) == 0 {
			branch_byte, next_address = zmachine.read_byte(next_address)
			offset = (offset << 8) | uint16(branch_byte)
		}

		switch offset {
		case 0:
			instruction.Branch.Behavior = BB_ReturnFalse
		case 1:
			instruction.Branch.Behavior = BB_ReturnTrue
		default:
			instruction.Branch.Address = Address(uint16(next_address) + offset - 2)
		}
	}

	// if instruction.HasText() {
	//     // TODO: Implement text handling
	// }

	return instruction, next_address, nil
}
