package zmachine

import "errors"

type Opcode uint16

var opcodes = map[Opcode]InstructionInfo{
	// 0x04: {IF_Long, IM_Branch, []OperandType{OT_Small, OT_Small}, dec_chk},
	// 0x4f: {IF_Long, IM_Store, []OperandType{OT_Variable, OT_Small}, loadw},
	0x54: {IF_Long, IM_Store, []OperandType{OT_Variable, OT_Small}, add},
	0x55: {IF_Long, IM_Store, []OperandType{OT_Variable, OT_Small}, sub},
	0x61: {IF_Long, IM_Branch, []OperandType{OT_Variable, OT_Variable}, je},
	0x74: {IF_Long, IM_Store, []OperandType{OT_Variable, OT_Variable}, add},
	// 0x86: {IF_Short, IM_None, []OperandType{OT_Large}, dec},
	// // 0x87: {IF_Short, IM_None, []OperandType{OT_Large}, print_addr},
	// 0x8c: {IF_Short, IM_None, []OperandType{OT_Large}, jump},
	0xa0: {IF_Short, IM_Branch, []OperandType{OT_Variable}, jz},
	0xe0: {IF_Variable, IM_Store, []OperandType{}, call},
	0xe1: {IF_Variable, IM_None, []OperandType{}, storew},
	// 0xe2: {IF_Variable, IM_None, []OperandType{}, storeb},
}

func (zmachine ZMachine) readOpcode(address Address) (Opcode, Address) {
	opcode, next_address := zmachine.readByte(address)

	if opcode == 0xbe {
		var ext_opcode word
		ext_opcode, next_address = zmachine.readWord(address)
		return Opcode(ext_opcode), next_address
	}

	return Opcode(opcode), next_address
}

func add(zmachine *ZMachine, instruction Instruction) (bool, error) {
	a := instruction.Operands[0].asWord()
	b := instruction.Operands[1].asWord()

	zmachine.writeVariable(a+b, instruction.StoreVariable)
	return false, nil
}

func call(zmachine *ZMachine, instruction Instruction) (bool, error) {
	packed_address := instruction.Operands[0].asAddress()
	if packed_address == 0 {
		return false, errors.New("unimplemented: call address 0")
	}

	routineAddr := zmachine.getRoutineAddress(packed_address)
	num_locals, next_address := zmachine.readByte(routineAddr)
	frame := StackFrame{}
	frame.Locals = make([]word, 0, num_locals)
	for range num_locals {
		var local word
		if zmachine.Header.Version < 5 {
			local, next_address = zmachine.readWord(next_address)
		} else {
			local = 0
		}
		frame.Locals = append(frame.Locals, local)
	}

	for i := 0; i < min(int(num_locals), len(instruction.Operands)-1); i++ {
		frame.Locals[i] = instruction.Operands[i+1].asWord()
	}

	frame.Counter = next_address

	frame.DiscardReturn = !instruction.StoresResult()
	if instruction.StoresResult() {
		frame.ReturnVariable = instruction.StoreVariable
	}

	zmachine.StackFrames.push(frame)
	return false, nil // Return false because the previous frame hasn't been updated yet even though there is a new frame
}

// func dec(zmachine *ZMachine, instruction Instruction) (bool, error) {
// 	variable := uint8(zmachine.get_operand_value(instruction, 0))

// 	variable_value := zmachine.read_variable(variable)
// 	variable_value--
// 	zmachine.write_variable(variable_value, variable)

// 	return false, nil
// }

// func dec_chk(zmachine *ZMachine, instruction Instruction) (bool, error) {
// 	variable := uint8(zmachine.get_operand_value(instruction, 0))
// 	value := zmachine.get_operand_value(instruction, 1)

// 	variable_value := zmachine.read_variable(variable)
// 	variable_value--
// 	zmachine.write_variable(variable_value, variable)

// 	if instruction.Branch.Condition == BC_OnTrue && variable_value < value ||
// 		instruction.Branch.Condition == BC_OnFalse && variable_value >= value {
// 		switch instruction.Branch.Behavior {
// 		case BB_Normal:
// 			zmachine.CurrentFrame().Counter = instruction.Branch.Address
// 			return true, nil
// 		case BB_ReturnFalse:
// 			zmachine.end_current_frame(0)
// 		case BB_ReturnTrue:
// 			zmachine.end_current_frame(1)
// 		}
// 	}

// 	return false, nil
// }

func je(zmachine *ZMachine, instruction Instruction) (bool, error) {
	a := instruction.Operands[0].asWord()
	b := instruction.Operands[1].asWord()

	if instruction.Branch.Condition == BC_OnTrue && a == b ||
		instruction.Branch.Condition == BC_OnFalse && a != b {
		switch instruction.Branch.Behavior {
		case BB_Normal:
			zmachine.StackFrames[0].Counter = instruction.Branch.Address
			return true, nil
		case BB_ReturnFalse:
			zmachine.endCurrentFrame(0)
		case BB_ReturnTrue:
			zmachine.endCurrentFrame(1)
		}
	}

	return false, nil
}

// func jump(zmachine *ZMachine, instruction Instruction) (bool, error) {
// 	offset := zmachine.get_operand_value(instruction, 0)
// 	frame := zmachine.CurrentFrame()
// 	frame.Counter = Address(uint16(frame.Counter) + offset)
// 	return true, nil
// }

func jz(zmachine *ZMachine, instruction Instruction) (bool, error) {
	a := instruction.Operands[0].asWord()

	if instruction.Branch.Condition == BC_OnTrue && a == 0 ||
		instruction.Branch.Condition == BC_OnFalse && a != 0 {
		switch instruction.Branch.Behavior {
		case BB_Normal:
			zmachine.StackFrames[0].Counter = instruction.Branch.Address
			return true, nil
		case BB_ReturnFalse:
			zmachine.endCurrentFrame(0)
		case BB_ReturnTrue:
			zmachine.endCurrentFrame(1)
		}
	}

	return false, nil
}

// func loadw(zmachine *ZMachine, instruction Instruction) (bool, error) {
// 	array := zmachine.get_operand_value(instruction, 0)
// 	word_index := zmachine.get_operand_value(instruction, 1)
// 	value, _ := zmachine.read_word(Address(array + 2*word_index))
// 	zmachine.write_variable(value, instruction.StoreVariable)
// 	return false, nil
// }

// // func print_addr(zmachine *ZMachine, instruction Instruction) bool {
// // 	address := Address(zmachine.get_operand_value(instruction.OperandValues[0]))
// // 	zmachine.read_zstring(address)
// // 	return false
// // }

func storew(zmachine *ZMachine, instruction Instruction) (bool, error) {
	array := instruction.Operands[0].asAddress()
	word_index := instruction.Operands[1].asInt()
	value := instruction.Operands[2].asWord()

	address := array.offsetWords(word_index)
	zmachine.writeWord(value, address)
	return false, nil
}

// func storeb(zmachine *ZMachine, instruction Instruction) (bool, error) {
// 	array := zmachine.get_operand_value(instruction, 0)
// 	byte_index := zmachine.get_operand_value(instruction, 1)
// 	value := uint8(zmachine.get_operand_value(instruction, 2))

// 	address := Address(array + byte_index)
// 	zmachine.write_byte(value, address)
// 	return false, nil
// }

func sub(zmachine *ZMachine, instruction Instruction) (bool, error) {
	a := instruction.Operands[0].asWord()
	b := instruction.Operands[1].asWord()
	zmachine.writeVariable(a-b, instruction.StoreVariable)
	return false, nil
}
