package zmachine

type InstructionHandler func(*ZMachine, Instruction)

type OpcodeInfo struct {
	PerformsStore  bool
	PerformsBranch bool
	HasText        bool
	Handler        InstructionHandler
}

var opcodes = map[uint8]OpcodeInfo{
	0xe0: {true, false, false, call}, // TODO: In V4 Store should equal false
	0xe1: {false, false, false, storew},
	0xe2: {false, false, false, storeb},
}

func call(zmachine *ZMachine, instruction Instruction) {
	// TODO: This probably isn't implemented completely correctly. There isn't really a way to return from a call right now...
	routineAddr := zmachine.get_routine_address((Address)(instruction.Operands[0].Value))
	num_locals, next_address := zmachine.read_byte(routineAddr)
	for range num_locals {
		var local uint16
		local, next_address = zmachine.read_word(next_address)
		zmachine.Stack = append(zmachine.Stack, local)
	}
	zmachine.Counter = next_address
}

func storew(zmachine *ZMachine, instruction Instruction) {
	array := instruction.Operands[0].Value
	word_index := instruction.Operands[1].Value
	value := instruction.Operands[2].Value

	address := Address(array + 2*word_index)
	zmachine.write_word(value, address)
}

func storeb(zmachine *ZMachine, instruction Instruction) {
	array := instruction.Operands[0].Value
	byte_index := instruction.Operands[1].Value
	value := uint8(instruction.Operands[2].Value)

	address := Address(array + byte_index)
	zmachine.write_byte(value, address)
}
