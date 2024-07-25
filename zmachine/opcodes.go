package zmachine

type InstructionHandler func(*ZMachine, Instruction) bool

type InstructionForm uint8

const (
	IF_Short    InstructionForm = 0
	IF_Long     InstructionForm = 1
	IF_Variable InstructionForm = 2
	IF_Extended InstructionForm = 3
)

type OperandType uint8

const (
	OT_Large    OperandType = 0
	OT_Small    OperandType = 1
	OT_Variable OperandType = 2
	OT_Omitted  OperandType = 3
)

type InstructionMeta uint8

const (
	IM_None   = 0
	IM_Store  = 1
	IM_Branch = 2
	IM_Text   = 4
)

type Instruction2 struct {
	Form         InstructionForm
	Meta         InstructionMeta
	OperandTypes []OperandType
	Handler      InstructionHandler
}

var opcodes2 = map[uint8]Instruction2{
	0x04: {IF_Long, IM_Branch, []OperandType{OT_Small, OT_Small}, dec_chk},
	0x4f: {IF_Long, IM_Store, []OperandType{OT_Variable, OT_Small}, loadw},
	0x54: {IF_Long, IM_Store, []OperandType{OT_Variable, OT_Small}, add},
	0x55: {IF_Long, IM_Store, []OperandType{OT_Variable, OT_Small}, sub},
	0x61: {IF_Long, IM_Branch, []OperandType{OT_Variable, OT_Variable}, je},
	0x74: {IF_Long, IM_Store, []OperandType{OT_Variable, OT_Variable}, add},
	0x86: {IF_Short, IM_None, []OperandType{OT_Large}, dec},
	// 0x87: {IF_Short, IM_None, []OperandType{OT_Large}, print_addr},
	0x8c: {IF_Short, IM_None, []OperandType{OT_Large}, jump},
	0xa0: {IF_Short, IM_Branch, []OperandType{OT_Variable}, jz},
	0xe0: {IF_Variable, IM_Store, []OperandType{}, call}, // TODO: In V4 Store should equal false
	0xe1: {IF_Variable, IM_None, []OperandType{}, storew},
	0xe2: {IF_Variable, IM_None, []OperandType{}, storeb},
}

var opcodes = map[uint8]OpcodeInfo{
	0x04: {false, true, false, dec_chk},
	0x4f: {true, false, false, loadw},
	0x54: {true, false, false, add},
	0x55: {true, false, false, sub},
	0x61: {false, true, false, je},
	0x74: {true, false, false, add},
	0x86: {false, false, false, dec},
	// 0x87: {false, false, false, print_addr},
	0x8c: {false, false, false, jump},
	0xa0: {false, true, false, jz},
	0xe0: {true, false, false, call}, // TODO: In V4 Store should equal false
	0xe1: {false, false, false, storew},
	0xe2: {false, false, false, storeb},
}

func add(zmachine *ZMachine, instruction Instruction) bool {
	a := zmachine.get_operand_value(instruction.Operands[0])
	b := zmachine.get_operand_value(instruction.Operands[1])
	zmachine.write_variable(a+b, instruction.Store)
	return false
}

func call(zmachine *ZMachine, instruction Instruction) bool {
	// TODO: This probably isn't implemented completely correctly. There isn't really a way to return from a call right now...
	// TODO: Need to store return value via `zmachine.write_variable(value, instruction.Store)`
	packed_address := Address(zmachine.get_operand_value(instruction.Operands[0]))
	if packed_address == 0 {
		panic("unimplemented: call address 0")
	}

	routineAddr := zmachine.get_routine_address(packed_address)
	num_locals, next_address := zmachine.read_byte(routineAddr)
	frame := StackFrame{}
	for range num_locals {
		var local uint16
		if zmachine.Header.Version < 5 {
			local, next_address = zmachine.read_word(next_address)
		} else {
			local = 0
		}
		frame.Locals = append(frame.Locals, local)
	}

	for i := 0; i < min(int(num_locals), len(instruction.Operands)-1); i++ {
		frame.Locals[i] = zmachine.get_operand_value(instruction.Operands[i+1])
	}
	frame.Counter = next_address
	zmachine.StackFrames = append(zmachine.StackFrames, frame)
	return false // Return false because the previous frame hasn't been updated yet even though there is a new frame
}

func dec(zmachine *ZMachine, instruction Instruction) bool {
	variable := uint8(zmachine.get_operand_value(instruction.Operands[0]))

	variable_value := zmachine.read_variable(variable)
	variable_value--
	zmachine.write_variable(variable_value, variable)

	return false
}

func dec_chk(zmachine *ZMachine, instruction Instruction) bool {
	variable := uint8(zmachine.get_operand_value(instruction.Operands[0]))
	value := zmachine.get_operand_value(instruction.Operands[1])

	variable_value := zmachine.read_variable(variable)
	variable_value--
	zmachine.write_variable(variable_value, variable)

	if instruction.BranchBehavior == BRANCHBEHAVIOR_None {
		panic("branch with no behavior")
	}

	if instruction.BranchBehavior == BRANCHBEHAVIOR_BranchOnTrue && variable_value < value ||
		instruction.BranchBehavior == BRANCHBEHAVIOR_BranchOnFalse && variable_value >= value {
		switch instruction.BranchAddress {
		case 0:
			panic("unimplemented: branch offset 0")
		case 1:
			panic("unimplemented: branch offset 1")
		default:
			zmachine.CurrentFrame().Counter = instruction.BranchAddress
			return true
		}
	}

	return false
}

func je(zmachine *ZMachine, instruction Instruction) bool {
	a := zmachine.get_operand_value(instruction.Operands[0])
	b := zmachine.get_operand_value(instruction.Operands[1])

	if instruction.BranchBehavior == BRANCHBEHAVIOR_None {
		panic("branch with no behavior")
	}

	if instruction.BranchBehavior == BRANCHBEHAVIOR_BranchOnTrue && a == b ||
		instruction.BranchBehavior == BRANCHBEHAVIOR_BranchOnFalse && a != b {
		switch instruction.BranchAddress {
		case 0:
			panic("unimplemented: branch offset 0")
		case 1:
			panic("unimplemented: branch offset 1")
		default:
			zmachine.CurrentFrame().Counter = instruction.BranchAddress
			return true
		}
	}

	return false
}

func jump(zmachine *ZMachine, instruction Instruction) bool {
	offset := zmachine.get_operand_value(instruction.Operands[0])
	frame := zmachine.CurrentFrame()
	frame.Counter = Address(uint16(frame.Counter) + offset)
	return true
}

func jz(zmachine *ZMachine, instruction Instruction) bool {
	a := zmachine.get_operand_value(instruction.Operands[0])

	if instruction.BranchBehavior == BRANCHBEHAVIOR_None {
		panic("branch with no behavior")
	}

	if instruction.BranchBehavior == BRANCHBEHAVIOR_BranchOnTrue && a == 0 ||
		instruction.BranchBehavior == BRANCHBEHAVIOR_BranchOnFalse && a != 0 {
		switch instruction.BranchAddress {
		case 0:
			panic("unimplemented: branch offset 0")
		case 1:
			panic("unimplemented: branch offset 1")
		default:
			zmachine.CurrentFrame().Counter = instruction.BranchAddress
			return true
		}
	}

	return false
}

func loadw(zmachine *ZMachine, instruction Instruction) bool {
	array := zmachine.get_operand_value(instruction.Operands[0])
	word_index := zmachine.get_operand_value(instruction.Operands[1])
	value, _ := zmachine.read_word(Address(array + 2*word_index))
	zmachine.write_variable(value, instruction.Store)
	return false
}

func print_addr(zmachine *ZMachine, instruction Instruction) bool {
	address := Address(zmachine.get_operand_value(instruction.Operands[0]))
	zmachine.read_zstring(address)
	return false
}

func storew(zmachine *ZMachine, instruction Instruction) bool {
	array := zmachine.get_operand_value(instruction.Operands[0])
	word_index := zmachine.get_operand_value(instruction.Operands[1])
	value := zmachine.get_operand_value(instruction.Operands[2])

	address := Address(array + 2*word_index)
	zmachine.write_word(value, address)
	return false
}

func storeb(zmachine *ZMachine, instruction Instruction) bool {
	array := zmachine.get_operand_value(instruction.Operands[0])
	byte_index := zmachine.get_operand_value(instruction.Operands[1])
	value := uint8(zmachine.get_operand_value(instruction.Operands[2]))

	address := Address(array + byte_index)
	zmachine.write_byte(value, address)
	return false
}

func sub(zmachine *ZMachine, instruction Instruction) bool {
	a := zmachine.get_operand_value(instruction.Operands[0])
	b := zmachine.get_operand_value(instruction.Operands[1])
	zmachine.write_variable(a-b, instruction.Store)
	return false
}
