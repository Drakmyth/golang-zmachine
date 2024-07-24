package zmachine

type InstructionHandler func(*ZMachine, Instruction) bool

type OpcodeInfo struct {
	PerformsStore  bool
	PerformsBranch bool
	HasText        bool
	Handler        InstructionHandler
}

var opcodes = map[uint8]OpcodeInfo{
	0x54: {true, false, false, add},
	0x55: {true, false, false, sub},
	0x61: {false, true, false, je},
	0x74: {true, false, false, add},
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

	for i := 0; i < min(int(num_locals), len(instruction.Operands)); i++ {
		frame.Locals[i] = zmachine.get_operand_value(instruction.Operands[i])
	}
	frame.Counter = next_address
	zmachine.StackFrames = append(zmachine.StackFrames, frame)
	return false // Return false because the previous frame hasn't been updated yet even though there is a new frame
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
