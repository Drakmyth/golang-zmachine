package zmachine

type InstructionHandler func(*ZMachine, Instruction, Address)

type OpcodeInfo struct {
	PerformsStore  bool
	PerformsBranch bool
	HasText        bool
	Handler        InstructionHandler
}

var opcodes = map[uint8]OpcodeInfo{
	0xe0: {true, false, false, call}, // call, TODO: In V4 Store should equal false
}

func call(zmachine *ZMachine, instruction Instruction, return_address Address) {
	routineAddr := zmachine.get_routine_address((Address)(instruction.Operands[0].Value))
	zmachine.Stack = append(zmachine.Stack, (uint16)(return_address))
	num_locals, next_address := zmachine.read_byte(routineAddr)
	for range num_locals {
		var local uint16
		local, next_address = zmachine.read_word(next_address)
		zmachine.Stack = append(zmachine.Stack, local)
	}
	zmachine.Counter = next_address
}
