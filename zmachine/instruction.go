package zmachine

import (
	"fmt"
	"reflect"
	"runtime"
	"strings"
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
	Address       Address
	NextAddress   Address
	Opcode        Opcode
	Operands      []Operand
	StoreVariable VarNum
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

func (zmachine ZMachine) readBranch(address Address) (Branch, Address) {
	branch := Branch{}
	branch_byte, next_address := zmachine.readByte(address)
	branch.Condition = BranchCondition(branch_byte >> 7)

	var offset word
	offset = word(branch_byte & 0b00111111)
	if ((branch_byte >> 6) & 0b01) == 0 {
		branch_byte, next_address = zmachine.readByte(next_address)
		offset = (offset << 8) | word(branch_byte)
	}

	switch offset {
	case 0:
		branch.Behavior = BB_ReturnFalse
	case 1:
		branch.Behavior = BB_ReturnTrue
	default:
		branch.Address = next_address.offsetBytes(int(offset) - 2)
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

func (operand Operand) asAddress() Address {
	return Address(operand)
}

func (operand Operand) asInt() int {
	return int(operand)
}

func (zmachine ZMachine) readOperand(optype OperandType, address Address) (Operand, Address) {
	switch optype {
	case OT_Large:
		opvalue, next_address := zmachine.readWord(address)
		return Operand(opvalue), next_address
	case OT_Small:
		fallthrough
	case OT_Variable:
		opvalue, next_address := zmachine.readByte(address)
		return Operand(opvalue), next_address
	default:
		return 0, address
	}
}
