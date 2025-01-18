package zmachine

import (
	"fmt"
	"reflect"
	"runtime"
	"slices"
	"strings"

	"github.com/Drakmyth/golang-zmachine/memory"
)

type InstructionForm uint8

const (
	IF_Short    InstructionForm = 0
	IF_Long     InstructionForm = 1
	IF_Variable InstructionForm = 2
	IF_Extended InstructionForm = 3
)

type InstructionMeta uint8

const (
	IM_None   = 0
	IM_Store  = 1
	IM_Branch = 2
	IM_Text   = 4
)

type OperandType uint8

const (
	OT_Large    OperandType = 0
	OT_Small    OperandType = 1
	OT_Variable OperandType = 2
	OT_Omitted  OperandType = 3
)

type InstructionHandler func(*ZMachine, Instruction) (bool, error)

type InstructionInfo struct {
	Form         InstructionForm
	Meta         InstructionMeta
	OperandTypes []OperandType
	Handler      InstructionHandler
}

func (info InstructionInfo) StoresResult() bool {
	return info.Meta&IM_Store == IM_Store
}

func (info InstructionInfo) Branches() bool {
	return info.Meta&IM_Branch == IM_Branch
}

func (info InstructionInfo) HasText() bool {
	return info.Meta&IM_Text == IM_Text
}

type Instruction struct {
	InstructionInfo
	Address       memory.Address
	NextAddress   memory.Address
	Opcode        Opcode
	Operands      []Operand
	StoreVariable Variable
	Branch        Branch
	Text          string
}

func (instruction Instruction) String() string {
	function_path := strings.Split(runtime.FuncForPC(reflect.ValueOf(instruction.Handler).Pointer()).Name(), ".")
	operation_name := strings.ToUpper(function_path[len(function_path)-1])

	log_strings := make([]string, 0, len(instruction.Operands)+4)
	log_strings = append(log_strings, fmt.Sprintf("%02x %s", instruction.Opcode, operation_name))
	for i, operand := range instruction.Operands {
		optype := instruction.OperandTypes[i]
		switch optype {
		case OT_Variable:
			log_strings = append(log_strings, fmt.Sprint(operand.asVarNum()))
		case OT_Small:
			log_strings = append(log_strings, fmt.Sprintf("#%02x", operand.asByte()))
		case OT_Large:
			log_strings = append(log_strings, fmt.Sprintf("%x", operand.asWord()))
		}
	}

	if instruction.StoresResult() {
		log_strings = append(log_strings, fmt.Sprintf("-> %s", fmt.Sprint(instruction.StoreVariable)))
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

func (zmachine ZMachine) readInstruction(address memory.Address) (Instruction, memory.Address, error) {
	opcode, next_address := zmachine.readOpcode(address)
	inst_info, ok := opcodes[opcode]
	if !ok {
		return Instruction{}, 0, fmt.Errorf("unknown opcode: %02x", opcode)
	}
	instruction := Instruction{InstructionInfo: inst_info, Opcode: opcode, Address: address}

	if instruction.Form == IF_Extended {
		return Instruction{}, 0, fmt.Errorf("unimplemented: extended form instruction")
	}

	// Determine Variable Form operand types
	if instruction.Form == IF_Variable {
		var types_byte uint8
		types_byte, next_address = zmachine.Memory.ReadByteNext(next_address)

		for shift := 0; shift <= 6; shift += 2 {
			operand_type := OperandType((types_byte >> shift) & 0b11)
			if operand_type != OT_Omitted {
				instruction.OperandTypes = append(instruction.OperandTypes, operand_type)
			}
		}
		// Types are parsed last to first, so we need to reverse them
		slices.Reverse(instruction.OperandTypes)
	}

	// Parse Operands
	instruction.Operands = make([]Operand, 0, len(instruction.OperandTypes))
	for _, optype := range instruction.OperandTypes {
		var operand Operand
		operand, next_address = zmachine.readOperand(optype, next_address)
		instruction.Operands = append(instruction.Operands, operand)
	}

	if instruction.StoresResult() {
		var store_varnum Variable
		store_varnum, next_address = zmachine.readVariable(next_address)
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

	instruction.NextAddress = next_address
	return instruction, next_address, nil
}

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
	Address   memory.Address
	Behavior  BranchBehavior
	Condition BranchCondition
}

func (zmachine ZMachine) readBranch(address memory.Address) (Branch, memory.Address) {
	branch := Branch{}
	branch_byte, next_address := zmachine.Memory.ReadByteNext(address)
	branch.Condition = BranchCondition(branch_byte >> 7)

	var offset word
	offset = word(branch_byte & 0b00111111)
	if ((branch_byte >> 6) & 0b01) == 0 {
		branch_byte, next_address = zmachine.Memory.ReadByteNext(next_address)
		offset = (offset << 8) | word(branch_byte)
	}

	switch offset {
	case 0:
		branch.Behavior = BB_ReturnFalse
	case 1:
		branch.Behavior = BB_ReturnTrue
	default:
		branch.Address = next_address.OffsetBytes(int(offset) - 2)
	}

	return branch, next_address
}

type Operand word

func (operand Operand) asWord() word {
	return word(operand)
}

func (operand Operand) asByte() byte {
	return byte(operand)
}

func (operand Operand) asVarNum() VarNum {
	return VarNum(operand)
}

func (operand Operand) asAddress() memory.Address {
	return memory.Address(operand)
}

func (operand Operand) asInt() int {
	return int(operand)
}

func (operand Operand) asObjectId() ObjectId {
	return ObjectId(operand)
}

func (operand Operand) asPropertyId() PropertyId {
	return PropertyId(operand)
}

func (zmachine ZMachine) readOperand(optype OperandType, address memory.Address) (Operand, memory.Address) {
	switch optype {
	case OT_Large:
		opvalue, next_address := zmachine.Memory.ReadWordNext(address)
		return Operand(opvalue), next_address
	case OT_Small:
		fallthrough
	case OT_Variable:
		opvalue, next_address := zmachine.Memory.ReadByteNext(address)
		return Operand(opvalue), next_address
	default:
		return 0, address
	}
}
