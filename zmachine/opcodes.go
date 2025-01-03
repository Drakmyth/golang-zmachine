package zmachine

import (
	"errors"
	"fmt"

	"github.com/Drakmyth/golang-zmachine/memory"
)

type Opcode uint16

var opcodes = map[Opcode]InstructionInfo{
	0x04: {IF_Long, IM_Branch, []OperandType{OT_Small, OT_Small}, dec_chk},
	0x05: {IF_Long, IM_Branch, []OperandType{OT_Small, OT_Small}, inc_chk},
	0x0d: {IF_Long, IM_None, []OperandType{OT_Small, OT_Small}, store},
	0x0f: {IF_Long, IM_Store, []OperandType{OT_Small, OT_Small}, loadw},
	0x10: {IF_Long, IM_Store, []OperandType{OT_Small, OT_Small}, loadb},
	0x2d: {IF_Long, IM_None, []OperandType{OT_Small, OT_Variable}, store},
	0x30: {IF_Long, IM_Store, []OperandType{OT_Small, OT_Variable}, loadb},
	0x49: {IF_Long, IM_Store, []OperandType{OT_Variable, OT_Small}, and},
	0x4a: {IF_Long, IM_Branch, []OperandType{OT_Variable, OT_Small}, test_attr},
	0x4f: {IF_Long, IM_Store, []OperandType{OT_Variable, OT_Small}, loadw},
	0x54: {IF_Long, IM_Store, []OperandType{OT_Variable, OT_Small}, add},
	0x55: {IF_Long, IM_Store, []OperandType{OT_Variable, OT_Small}, sub},
	0x61: {IF_Long, IM_Branch, []OperandType{OT_Variable, OT_Variable}, je},
	0x74: {IF_Long, IM_Store, []OperandType{OT_Variable, OT_Variable}, add},
	0x86: {IF_Short, IM_None, []OperandType{OT_Large}, dec},
	// // 0x87: {IF_Short, IM_None, []OperandType{OT_Large}, print_addr},
	0x8c: {IF_Short, IM_None, []OperandType{OT_Large}, jump},
	0xa0: {IF_Short, IM_Branch, []OperandType{OT_Variable}, jz},
	0xab: {IF_Short, IM_None, []OperandType{OT_Variable}, ret},
	0xb0: {IF_Short, IM_None, []OperandType{}, rtrue},
	0xb2: {IF_Short, IM_None, []OperandType{}, print},
	0xbb: {IF_Short, IM_None, []OperandType{}, new_line},
	0xc9: {IF_Variable, IM_Store, []OperandType{}, and},
	0xe0: {IF_Variable, IM_Store, []OperandType{}, call},
	0xe1: {IF_Variable, IM_None, []OperandType{}, storew},
	// 0xe2: {IF_Variable, IM_None, []OperandType{}, storeb},
	0xe3: {IF_Variable, IM_None, []OperandType{}, put_prop},
	0xe5: {IF_Variable, IM_None, []OperandType{}, print_char},
	0xe6: {IF_Variable, IM_None, []OperandType{}, print_num},
}

func (zmachine ZMachine) readOpcode(address memory.Address) (Opcode, memory.Address) {
	opcode, next_address := zmachine.Memory.ReadByteNext(address)

	if opcode == 0xbe {
		var ext_opcode word
		ext_opcode, next_address = zmachine.Memory.ReadWordNext(address)
		return Opcode(ext_opcode), next_address
	}

	return Opcode(opcode), next_address
}

// func (zmachine ZMachine) getRoutineAddress(address memory.Address) memory.Address {
// 	switch zmachine.Memory.GetVersion() {
// 	case 1, 2, 3:
// 		return address * 2
// 	case 4, 5:
// 		return address * 4
// 	case 6, 7:
// 		return address*4 + zmachine.Header.RoutinesAddr*8
// 	case 8:
// 		return address * 8
// 	}

// 	panic("Unknown version")
// }

func (zmachine ZMachine) performBranch(branch Branch, condition bool) bool {
	if branch.Condition == BC_OnTrue && condition ||
		branch.Condition == BC_OnFalse && !condition {
		switch branch.Behavior {
		case BB_Normal:
			zmachine.Stack.Peek().Counter = branch.Address
			return true
		case BB_ReturnFalse:
			zmachine.endCurrentFrame(0)
		case BB_ReturnTrue:
			zmachine.endCurrentFrame(1)
		}
	}

	return false
}

func add(zmachine *ZMachine, instruction Instruction) (bool, error) {
	a := instruction.Operands[0].asWord()
	b := instruction.Operands[1].asWord()

	instruction.StoreVariable.Write(a + b)
	return false, nil
}

func and(zmachine *ZMachine, instruction Instruction) (bool, error) {
	a := instruction.Operands[0].asWord()
	b := instruction.Operands[1].asWord()

	instruction.StoreVariable.Write(a & b)
	return false, nil
}

func call(zmachine *ZMachine, instruction Instruction) (bool, error) {
	packed_address := instruction.Operands[0].asWord()
	if packed_address == 0 {
		return false, errors.New("unimplemented: call address 0")
	}

	routineAddr := zmachine.Memory.RoutinePackedAddress(packed_address)
	num_locals, next_address := zmachine.Memory.ReadByteNext(routineAddr)
	frame := zmachine.NewFrame()
	frame.Locals = make([]word, 0, num_locals)
	for range num_locals {
		var local word
		if zmachine.Memory.GetVersion() < 5 {
			local, next_address = zmachine.Memory.ReadWordNext(next_address)
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

	zmachine.Stack.Push(frame)
	return false, nil // Return false because the previous frame hasn't been updated yet even though there is a new frame
}

func dec(zmachine *ZMachine, instruction Instruction) (bool, error) {
	variable := zmachine.getVariable(instruction.Operands[0].asVarNum())

	// TODO: Fix stack handling, needs to read/write in place instead of modifying stack
	// Is this actually a problem? It will pop it off, but then push it right back on.
	// The address will change potentially, but does that matter?
	value := variable.Read()
	value--
	variable.Write(value)

	return false, nil
}

func dec_chk(zmachine *ZMachine, instruction Instruction) (bool, error) {
	variable := zmachine.getVariable(instruction.Operands[0].asVarNum())
	condition := instruction.Operands[1].asWord()

	// TODO: Fix stack handling, needs to read/write in place instead of modifying stack
	// Is this actually a problem? It will pop it off, but then push it right back on.
	// The address will change potentially, but does that matter?
	value := variable.Read()
	value--
	variable.Write(value)

	return zmachine.performBranch(instruction.Branch, value < condition), nil
}

func inc_chk(zmachine *ZMachine, instruction Instruction) (bool, error) {
	variable := zmachine.getVariable(instruction.Operands[0].asVarNum())
	condition := instruction.Operands[1].asWord()

	// TODO: Fix stack handling, needs to read/write in place instead of modifying stack
	// Is this actually a problem? It will pop it off, but then push it right back on.
	// The address will change potentially, but does that matter?
	value := variable.Read()
	value++
	variable.Write(value)

	return zmachine.performBranch(instruction.Branch, value > condition), nil
}

func je(zmachine *ZMachine, instruction Instruction) (bool, error) {
	a := instruction.Operands[0].asWord()
	b := instruction.Operands[1].asWord()

	return zmachine.performBranch(instruction.Branch, a == b), nil
}

func jump(zmachine *ZMachine, instruction Instruction) (bool, error) {
	offset := instruction.Operands[0].asInt()

	frame := zmachine.Stack.Peek()
	frame.Counter = instruction.NextAddress.OffsetBytes(offset - 2)
	return true, nil
}

func jz(zmachine *ZMachine, instruction Instruction) (bool, error) {
	a := instruction.Operands[0].asWord()

	return zmachine.performBranch(instruction.Branch, a == 0), nil
}

func loadb(zmachine *ZMachine, instruction Instruction) (bool, error) {
	array := instruction.Operands[0].asAddress()
	index := instruction.Operands[1].asInt()

	value := zmachine.Memory.ReadByte(array.OffsetBytes(index))
	instruction.StoreVariable.Write(word(value))

	return false, nil
}

func loadw(zmachine *ZMachine, instruction Instruction) (bool, error) {
	array := instruction.Operands[0].asAddress()
	word_index := instruction.Operands[1].asInt()

	address := array.OffsetWords(word_index)
	value := zmachine.Memory.ReadWord(address)

	instruction.StoreVariable.Write(value)
	return false, nil
}

func new_line(zmachine *ZMachine, instruction Instruction) (bool, error) {
	fmt.Println()
	return false, nil
}

func print(zmachine *ZMachine, instruction Instruction) (bool, error) {
	str, next_address := zmachine.readZString(instruction.NextAddress)
	fmt.Print(str)
	if zmachine.Debug {
		fmt.Println()
	}

	branch := Branch{
		Address:   next_address,
		Behavior:  BB_Normal,
		Condition: BC_OnTrue,
	}

	return zmachine.performBranch(branch, true), nil
}

// func print_addr(zmachine *ZMachine, instruction Instruction) (bool, error) {
// 	address := instruction.Operands[0].asAddress()
// 	zmachine.read_zstring(address)
// 	return false, nil
// }

func print_char(zmachine *ZMachine, instruction Instruction) (bool, error) {
	a := instruction.Operands[0].asByte()

	// TODO: This should convert to ZSCII rather than ASCII/Unicode
	fmt.Print(string(a))
	if zmachine.Debug {
		fmt.Println()
	}

	return false, nil
}

func print_num(zmachine *ZMachine, instruction Instruction) (bool, error) {
	a := instruction.Operands[0].asInt()
	fmt.Printf("%v", a)
	if zmachine.Debug {
		fmt.Println()
	}

	return false, nil
}

func put_prop(zmachine *ZMachine, instruction Instruction) (bool, error) {
	object_index := instruction.Operands[0].asInt()
	property_index := instruction.Operands[1].asByte()
	value := instruction.Operands[2].asWord()

	object := zmachine.getObject(object_index)
	properties, _ := zmachine.readProperties(object.PropertiesAddr)

	property_data_length := len(properties.Properties[property_index])
	switch property_data_length {
	case 1:
		properties.Properties[property_index][0] = byte(value)
	case 2:
		properties.Properties[property_index][0] = byte(value >> 8)
		properties.Properties[property_index][1] = byte(value)
	default:
		return false, fmt.Errorf("unsupported put_prop data length: %d", property_data_length)
	}

	return false, nil
}

func ret(zmachine *ZMachine, instruction Instruction) (bool, error) {
	value := instruction.Operands[0].asWord()

	zmachine.endCurrentFrame(value)
	return true, nil
}

func rtrue(zmachine *ZMachine, instruction Instruction) (bool, error) {
	zmachine.endCurrentFrame(1)
	return true, nil
}

func storew(zmachine *ZMachine, instruction Instruction) (bool, error) {
	array := instruction.Operands[0].asAddress()
	word_index := instruction.Operands[1].asInt()
	value := instruction.Operands[2].asWord()

	address := array.OffsetWords(word_index)
	zmachine.Memory.WriteWord(address, value)
	return false, nil
}

func store(zmachine *ZMachine, instruction Instruction) (bool, error) {
	variable := zmachine.getVariable(instruction.Operands[0].asVarNum())
	value := instruction.Operands[0].asWord()

	variable.Write(value)
	return false, nil
}

// func storeb(zmachine *ZMachine, instruction Instruction) (bool, error) {
// array := instruction.Operands[0].asAddress()
// byte_index := instruction.Operands[1].asInt()
// value := instruction.Operands[2].asByte()

// 	address := array.OffsetBytes(byte_index)
// 	zmachine.writeByte(value, address)
// 	return false, nil
// }

func sub(zmachine *ZMachine, instruction Instruction) (bool, error) {
	a := instruction.Operands[0].asWord()
	b := instruction.Operands[1].asWord()
	instruction.StoreVariable.Write(a - b)
	return false, nil
}

func test_attr(zmachine *ZMachine, instruction Instruction) (bool, error) {
	object_index := instruction.Operands[0].asInt()
	attribute_index := instruction.Operands[1].asInt()

	return zmachine.performBranch(instruction.Branch, zmachine.getObject(object_index).hasAttribute(attribute_index)), nil
}
