package zmachine

type InstructionHandler func(*ZMachine, Instruction)

type OpcodeInfo struct {
	PerformsStore  bool
	PerformsBranch bool
	HasText        bool
	Handler        InstructionHandler
}

var opcodes = map[uint8]OpcodeInfo{
	0x54: {true, false, false, add},
	0xe0: {true, false, false, call}, // TODO: In V4 Store should equal false
	0xe1: {false, false, false, storew},
	0xe2: {false, false, false, storeb},
}

func add(zmachine *ZMachine, instruction Instruction) {
	a := instruction.Operands[0].Value
	b := instruction.Operands[1].Value
	zmachine.write_variable(a+b, instruction.Store)
}

func call(zmachine *ZMachine, instruction Instruction) {
	// TODO: This probably isn't implemented completely correctly. There isn't really a way to return from a call right now...
	// TODO: Need to store return value via `zmachine.write_variable(value, instruction.Store)`
	routineAddr := zmachine.get_routine_address((Address)(instruction.Operands[0].Value))
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
	frame.Counter = next_address
	zmachine.StackFrames = append(zmachine.StackFrames, frame)
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
