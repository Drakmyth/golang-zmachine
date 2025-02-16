package zmachine

import (
	"fmt"
	"math/rand/v2"
	"slices"
	"strings"
	"time"

	"github.com/Drakmyth/golang-zmachine/assert"
	"github.com/Drakmyth/golang-zmachine/memory"
	"github.com/Drakmyth/golang-zmachine/zstring"
)

type Opcode word

var opcodes = map[Opcode]InstructionInfo{
	0x01: {IF_Long, IM_Branch, []OperandType{OT_Small, OT_Small}, je},
	0x02: {IF_Long, IM_Branch, []OperandType{OT_Small, OT_Small}, jl},
	0x03: {IF_Long, IM_Branch, []OperandType{OT_Small, OT_Small}, jg},
	0x04: {IF_Long, IM_Branch, []OperandType{OT_Small, OT_Small}, dec_chk},
	0x05: {IF_Long, IM_Branch, []OperandType{OT_Small, OT_Small}, inc_chk},
	0x06: {IF_Long, IM_Branch, []OperandType{OT_Small, OT_Small}, jin},
	0x07: {IF_Long, IM_Branch, []OperandType{OT_Small, OT_Small}, test},
	0x08: {IF_Long, IM_Store, []OperandType{OT_Small, OT_Small}, or},
	0x09: {IF_Long, IM_Store, []OperandType{OT_Small, OT_Small}, and},
	0x0a: {IF_Long, IM_Branch, []OperandType{OT_Small, OT_Small}, test_attr},
	0x0b: {IF_Long, IM_None, []OperandType{OT_Small, OT_Small}, set_attr},
	0x0c: {IF_Long, IM_None, []OperandType{OT_Small, OT_Small}, clear_attr},
	0x0d: {IF_Long, IM_None, []OperandType{OT_Small, OT_Small}, store},
	0x0e: {IF_Long, IM_None, []OperandType{OT_Small, OT_Small}, insert_obj},
	0x0f: {IF_Long, IM_Store, []OperandType{OT_Small, OT_Small}, loadw},
	0x10: {IF_Long, IM_Store, []OperandType{OT_Small, OT_Small}, loadb},
	0x11: {IF_Long, IM_Store, []OperandType{OT_Small, OT_Small}, get_prop},
	0x12: {IF_Long, IM_Store, []OperandType{OT_Small, OT_Small}, get_prop_addr},
	0x13: {IF_Long, IM_Store, []OperandType{OT_Small, OT_Small}, get_next_prop},
	0x14: {IF_Long, IM_Store, []OperandType{OT_Small, OT_Small}, add},
	0x15: {IF_Long, IM_Store, []OperandType{OT_Small, OT_Small}, sub},
	0x16: {IF_Long, IM_Store, []OperandType{OT_Small, OT_Small}, mul},
	0x17: {IF_Long, IM_Store, []OperandType{OT_Small, OT_Small}, div},
	0x18: {IF_Long, IM_Store, []OperandType{OT_Small, OT_Small}, mod},
	0x21: {IF_Long, IM_Branch, []OperandType{OT_Small, OT_Variable}, je},
	0x22: {IF_Long, IM_Branch, []OperandType{OT_Small, OT_Variable}, jl},
	0x23: {IF_Long, IM_Branch, []OperandType{OT_Small, OT_Variable}, jg},
	0x24: {IF_Long, IM_Branch, []OperandType{OT_Small, OT_Variable}, dec_chk},
	0x25: {IF_Long, IM_Branch, []OperandType{OT_Small, OT_Variable}, inc_chk},
	0x26: {IF_Long, IM_Branch, []OperandType{OT_Small, OT_Variable}, jin},
	0x27: {IF_Long, IM_Branch, []OperandType{OT_Small, OT_Variable}, test},
	0x28: {IF_Long, IM_Store, []OperandType{OT_Small, OT_Variable}, or},
	0x29: {IF_Long, IM_Store, []OperandType{OT_Small, OT_Variable}, and},
	0x2a: {IF_Long, IM_Branch, []OperandType{OT_Small, OT_Variable}, test_attr},
	0x2b: {IF_Long, IM_None, []OperandType{OT_Small, OT_Variable}, set_attr},
	0x2c: {IF_Long, IM_None, []OperandType{OT_Small, OT_Variable}, clear_attr},
	0x2d: {IF_Long, IM_None, []OperandType{OT_Small, OT_Variable}, store},
	0x2e: {IF_Long, IM_None, []OperandType{OT_Small, OT_Variable}, insert_obj},
	0x2f: {IF_Long, IM_Store, []OperandType{OT_Small, OT_Variable}, loadw},
	0x30: {IF_Long, IM_Store, []OperandType{OT_Small, OT_Variable}, loadb},
	0x31: {IF_Long, IM_Store, []OperandType{OT_Small, OT_Variable}, get_prop},
	0x32: {IF_Long, IM_Store, []OperandType{OT_Small, OT_Variable}, get_prop_addr},
	0x33: {IF_Long, IM_Store, []OperandType{OT_Small, OT_Variable}, get_next_prop},
	0x34: {IF_Long, IM_Store, []OperandType{OT_Small, OT_Variable}, add},
	0x35: {IF_Long, IM_Store, []OperandType{OT_Small, OT_Variable}, sub},
	0x36: {IF_Long, IM_Store, []OperandType{OT_Small, OT_Variable}, mul},
	0x37: {IF_Long, IM_Store, []OperandType{OT_Small, OT_Variable}, div},
	0x38: {IF_Long, IM_Store, []OperandType{OT_Small, OT_Variable}, mod},
	0x41: {IF_Long, IM_Branch, []OperandType{OT_Variable, OT_Small}, je},
	0x42: {IF_Long, IM_Branch, []OperandType{OT_Variable, OT_Small}, jl},
	0x43: {IF_Long, IM_Branch, []OperandType{OT_Variable, OT_Small}, jg},
	0x44: {IF_Long, IM_Branch, []OperandType{OT_Variable, OT_Small}, dec_chk},
	0x45: {IF_Long, IM_Branch, []OperandType{OT_Variable, OT_Small}, inc_chk},
	0x46: {IF_Long, IM_Branch, []OperandType{OT_Variable, OT_Small}, jin},
	0x47: {IF_Long, IM_Branch, []OperandType{OT_Variable, OT_Small}, test},
	0x48: {IF_Long, IM_Store, []OperandType{OT_Variable, OT_Small}, or},
	0x49: {IF_Long, IM_Store, []OperandType{OT_Variable, OT_Small}, and},
	0x4a: {IF_Long, IM_Branch, []OperandType{OT_Variable, OT_Small}, test_attr},
	0x4b: {IF_Long, IM_None, []OperandType{OT_Variable, OT_Small}, set_attr},
	0x4c: {IF_Long, IM_None, []OperandType{OT_Variable, OT_Small}, clear_attr},
	0x4d: {IF_Long, IM_None, []OperandType{OT_Variable, OT_Small}, store},
	0x4e: {IF_Long, IM_None, []OperandType{OT_Variable, OT_Small}, insert_obj},
	0x4f: {IF_Long, IM_Store, []OperandType{OT_Variable, OT_Small}, loadw},
	0x50: {IF_Long, IM_Store, []OperandType{OT_Variable, OT_Small}, loadb},
	0x51: {IF_Long, IM_Store, []OperandType{OT_Variable, OT_Small}, get_prop},
	0x52: {IF_Long, IM_Store, []OperandType{OT_Variable, OT_Small}, get_prop_addr},
	0x53: {IF_Long, IM_Store, []OperandType{OT_Variable, OT_Small}, get_next_prop},
	0x54: {IF_Long, IM_Store, []OperandType{OT_Variable, OT_Small}, add},
	0x55: {IF_Long, IM_Store, []OperandType{OT_Variable, OT_Small}, sub},
	0x56: {IF_Long, IM_Store, []OperandType{OT_Variable, OT_Small}, mul},
	0x57: {IF_Long, IM_Store, []OperandType{OT_Variable, OT_Small}, div},
	0x58: {IF_Long, IM_Store, []OperandType{OT_Variable, OT_Small}, mod},
	0x61: {IF_Long, IM_Branch, []OperandType{OT_Variable, OT_Variable}, je},
	0x62: {IF_Long, IM_Branch, []OperandType{OT_Variable, OT_Variable}, jl},
	0x63: {IF_Long, IM_Branch, []OperandType{OT_Variable, OT_Variable}, jg},
	0x64: {IF_Long, IM_Branch, []OperandType{OT_Variable, OT_Variable}, dec_chk},
	0x65: {IF_Long, IM_Branch, []OperandType{OT_Variable, OT_Variable}, inc_chk},
	0x66: {IF_Long, IM_Branch, []OperandType{OT_Variable, OT_Variable}, jin},
	0x67: {IF_Long, IM_Branch, []OperandType{OT_Variable, OT_Variable}, test},
	0x68: {IF_Long, IM_Store, []OperandType{OT_Variable, OT_Variable}, or},
	0x69: {IF_Long, IM_Store, []OperandType{OT_Variable, OT_Variable}, and},
	0x6a: {IF_Long, IM_Branch, []OperandType{OT_Variable, OT_Variable}, test_attr},
	0x6b: {IF_Long, IM_None, []OperandType{OT_Variable, OT_Variable}, set_attr},
	0x6c: {IF_Long, IM_None, []OperandType{OT_Variable, OT_Variable}, clear_attr},
	0x6d: {IF_Long, IM_None, []OperandType{OT_Variable, OT_Variable}, store},
	0x6e: {IF_Long, IM_None, []OperandType{OT_Variable, OT_Variable}, insert_obj},
	0x6f: {IF_Long, IM_Store, []OperandType{OT_Variable, OT_Variable}, loadw},
	0x70: {IF_Long, IM_Store, []OperandType{OT_Variable, OT_Variable}, loadb},
	0x71: {IF_Long, IM_Store, []OperandType{OT_Variable, OT_Variable}, get_prop},
	0x72: {IF_Long, IM_Store, []OperandType{OT_Variable, OT_Variable}, get_prop_addr},
	0x73: {IF_Long, IM_Store, []OperandType{OT_Variable, OT_Variable}, get_next_prop},
	0x74: {IF_Long, IM_Store, []OperandType{OT_Variable, OT_Variable}, add},
	0x75: {IF_Long, IM_Store, []OperandType{OT_Variable, OT_Variable}, sub},
	0x76: {IF_Long, IM_Store, []OperandType{OT_Variable, OT_Variable}, mul},
	0x77: {IF_Long, IM_Store, []OperandType{OT_Variable, OT_Variable}, div},
	0x78: {IF_Long, IM_Store, []OperandType{OT_Variable, OT_Variable}, mod},
	0x80: {IF_Short, IM_Branch, []OperandType{OT_Large}, jz},
	0x81: {IF_Short, IM_Branch | IM_Store, []OperandType{OT_Large}, get_sibling},
	0x82: {IF_Short, IM_Branch | IM_Store, []OperandType{OT_Large}, get_child},
	0x83: {IF_Short, IM_Store, []OperandType{OT_Large}, get_parent},
	0x84: {IF_Short, IM_Store, []OperandType{OT_Large}, get_prop_len},
	0x85: {IF_Short, IM_None, []OperandType{OT_Large}, inc},
	0x86: {IF_Short, IM_None, []OperandType{OT_Large}, dec},
	0x87: {IF_Short, IM_None, []OperandType{OT_Large}, print_addr},
	0x89: {IF_Short, IM_None, []OperandType{OT_Large}, remove_obj},
	0x8a: {IF_Short, IM_None, []OperandType{OT_Large}, print_obj},
	0x8b: {IF_Short, IM_None, []OperandType{OT_Large}, ret},
	0x8c: {IF_Short, IM_None, []OperandType{OT_Large}, jump},
	0x8d: {IF_Short, IM_None, []OperandType{OT_Large}, print_paddr},
	0x8e: {IF_Short, IM_Store, []OperandType{OT_Large}, load},
	0x8f: {IF_Short, IM_Store, []OperandType{OT_Large}, not},
	0x90: {IF_Short, IM_Branch, []OperandType{OT_Small}, jz},
	0x91: {IF_Short, IM_Branch | IM_Store, []OperandType{OT_Small}, get_sibling},
	0x92: {IF_Short, IM_Branch | IM_Store, []OperandType{OT_Small}, get_child},
	0x93: {IF_Short, IM_Store, []OperandType{OT_Small}, get_parent},
	0x94: {IF_Short, IM_Store, []OperandType{OT_Small}, get_prop_len},
	0x95: {IF_Short, IM_None, []OperandType{OT_Small}, inc},
	0x96: {IF_Short, IM_None, []OperandType{OT_Small}, dec},
	0x97: {IF_Short, IM_None, []OperandType{OT_Small}, print_addr},
	0x99: {IF_Short, IM_None, []OperandType{OT_Small}, remove_obj},
	0x9a: {IF_Short, IM_None, []OperandType{OT_Small}, print_obj},
	0x9b: {IF_Short, IM_None, []OperandType{OT_Small}, ret},
	0x9c: {IF_Short, IM_None, []OperandType{OT_Small}, jump},
	0x9d: {IF_Short, IM_None, []OperandType{OT_Small}, print_paddr},
	0x9e: {IF_Short, IM_Store, []OperandType{OT_Small}, load},
	0x9f: {IF_Short, IM_Store, []OperandType{OT_Small}, not}, // This opcode changed to `call_1n` in V5
	0xa0: {IF_Short, IM_Branch, []OperandType{OT_Variable}, jz},
	0xa1: {IF_Short, IM_Branch | IM_Store, []OperandType{OT_Variable}, get_sibling},
	0xa2: {IF_Short, IM_Branch | IM_Store, []OperandType{OT_Variable}, get_child},
	0xa3: {IF_Short, IM_Store, []OperandType{OT_Variable}, get_parent},
	0xa4: {IF_Short, IM_Store, []OperandType{OT_Variable}, get_prop_len},
	0xa5: {IF_Short, IM_None, []OperandType{OT_Variable}, inc},
	0xa6: {IF_Short, IM_None, []OperandType{OT_Variable}, dec},
	0xa7: {IF_Short, IM_None, []OperandType{OT_Variable}, print_addr},
	0xa9: {IF_Short, IM_None, []OperandType{OT_Variable}, remove_obj},
	0xaa: {IF_Short, IM_None, []OperandType{OT_Variable}, print_obj},
	0xab: {IF_Short, IM_None, []OperandType{OT_Variable}, ret},
	0xac: {IF_Short, IM_None, []OperandType{OT_Variable}, jump},
	0xad: {IF_Short, IM_None, []OperandType{OT_Variable}, print_paddr},
	0xae: {IF_Short, IM_Store, []OperandType{OT_Variable}, load},
	0xaf: {IF_Short, IM_Store, []OperandType{OT_Variable}, not}, // This opcode changed to `call_1n` in V5
	0xb0: {IF_Short, IM_None, []OperandType{}, rtrue},
	0xb1: {IF_Short, IM_None, []OperandType{}, rfalse},
	0xb2: {IF_Short, IM_None, []OperandType{}, print},
	0xb3: {IF_Short, IM_None, []OperandType{}, print_ret},
	0xb8: {IF_Short, IM_None, []OperandType{}, ret_popped},
	0xb9: {IF_Short, IM_None, []OperandType{}, pop}, // This opcode changed to `catch` in V5
	0xba: {IF_Short, IM_None, []OperandType{}, quit},
	0xbb: {IF_Short, IM_None, []OperandType{}, new_line},
	0xbd: {IF_Short, IM_Branch, []OperandType{}, verify},
	0xc1: {IF_Variable, IM_Branch, []OperandType{}, je},
	0xc2: {IF_Variable, IM_Branch, []OperandType{}, jl},
	0xc3: {IF_Variable, IM_Branch, []OperandType{}, jg},
	0xc4: {IF_Variable, IM_Branch, []OperandType{}, dec_chk},
	0xc5: {IF_Variable, IM_Branch, []OperandType{}, inc_chk},
	0xc6: {IF_Variable, IM_Branch, []OperandType{}, jin},
	0xc7: {IF_Variable, IM_Branch, []OperandType{}, test},
	0xc8: {IF_Variable, IM_Store, []OperandType{}, or},
	0xc9: {IF_Variable, IM_Store, []OperandType{}, and},
	0xca: {IF_Variable, IM_Branch, []OperandType{}, test_attr},
	0xcb: {IF_Variable, IM_None, []OperandType{}, set_attr},
	0xcc: {IF_Variable, IM_None, []OperandType{}, clear_attr},
	0xcd: {IF_Variable, IM_None, []OperandType{}, store},
	0xce: {IF_Variable, IM_None, []OperandType{}, insert_obj},
	0xcf: {IF_Variable, IM_Store, []OperandType{}, loadw},
	0xd0: {IF_Variable, IM_Store, []OperandType{}, loadb},
	0xd1: {IF_Variable, IM_Store, []OperandType{}, get_prop},
	0xd2: {IF_Variable, IM_Store, []OperandType{}, get_prop_addr},
	0xd3: {IF_Variable, IM_Store, []OperandType{}, get_next_prop},
	0xd4: {IF_Variable, IM_Store, []OperandType{}, add},
	0xd5: {IF_Variable, IM_Store, []OperandType{}, sub},
	0xd6: {IF_Variable, IM_Store, []OperandType{}, mul},
	0xd7: {IF_Variable, IM_Store, []OperandType{}, div},
	0xd8: {IF_Variable, IM_Store, []OperandType{}, mod},
	0xe0: {IF_Variable, IM_Store, []OperandType{}, call},
	0xe1: {IF_Variable, IM_None, []OperandType{}, storew},
	0xe2: {IF_Variable, IM_None, []OperandType{}, storeb},
	0xe3: {IF_Variable, IM_None, []OperandType{}, put_prop},
	0xe4: {IF_Variable, IM_None, []OperandType{}, read}, // In V5, this uses IM_STORE
	0xe5: {IF_Variable, IM_None, []OperandType{}, print_char},
	0xe6: {IF_Variable, IM_None, []OperandType{}, print_num},
	0xe7: {IF_Variable, IM_Store, []OperandType{}, random},
	0xe8: {IF_Variable, IM_None, []OperandType{}, push},
	0xe9: {IF_Variable, IM_None, []OperandType{}, pull}, // There's an extra argument here in V6
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

func (zmachine *ZMachine) performBranch(branch Branch, condition bool) bool {
	if branch.Condition == BC_OnTrue && condition ||
		branch.Condition == BC_OnFalse && !condition {
		switch branch.Behavior {
		case BB_Normal:
			frame, err := zmachine.Stack.Peek()
			assert.NoError(err, "Error peeking frame stack")
			frame.Counter = branch.Address
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
	a := int16(instruction.Operands[0].asWord())
	b := int16(instruction.Operands[1].asWord())

	instruction.StoreVariable.Write(uint16(a + b))
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
		zmachine.endCurrentFrame(0)
		return false, nil
	}

	routineAddr := zmachine.Memory.RoutinePackedAddress(packed_address)
	num_locals, next_address := zmachine.Memory.ReadByteNext(routineAddr)

	frame := Frame{ReturnVariable: zmachine.getVariable(0)}
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

func clear_attr(zmachine *ZMachine, instruction Instruction) (bool, error) {
	object := instruction.Operands[0].asObjectId()
	attribute := instruction.Operands[1].asInt()

	GetObject(zmachine.Memory, object).ClearAttribute(attribute)

	return false, nil
}

func dec(zmachine *ZMachine, instruction Instruction) (bool, error) {
	variable := zmachine.getVariable(instruction.Operands[0].asVarNum())

	value := int16(variable.ReadInPlace())
	value--
	variable.WriteInPlace(uint16(value))

	return false, nil
}

func dec_chk(zmachine *ZMachine, instruction Instruction) (bool, error) {
	variable := zmachine.getVariable(instruction.Operands[0].asVarNum())
	condition := int16(instruction.Operands[1].asWord())

	value := int16(variable.ReadInPlace())
	value--
	variable.WriteInPlace(uint16(value))

	return zmachine.performBranch(instruction.Branch, value < condition), nil
}

func div(zmachine *ZMachine, instruction Instruction) (bool, error) {
	a := int16(instruction.Operands[0].asWord())
	b := int16(instruction.Operands[1].asWord())

	assert.NotSame(b, 0, "Cannot divide by zero")

	instruction.StoreVariable.Write(uint16(a / b))
	return false, nil
}

func get_child(zmachine *ZMachine, instruction Instruction) (bool, error) {
	object := GetObject(zmachine.Memory, instruction.Operands[0].asObjectId())
	child := object.Child()

	instruction.StoreVariable.Write(word(child))
	return zmachine.performBranch(instruction.Branch, child != 0), nil
}

func get_next_prop(zmachine *ZMachine, instruction Instruction) (bool, error) {
	objectId := instruction.Operands[0].asObjectId()
	propertyId := instruction.Operands[1].asPropertyId()

	object := GetObject(zmachine.Memory, objectId)
	nextPropId := object.GetNextPropertyId(propertyId)

	instruction.StoreVariable.Write(word(nextPropId))
	return false, nil
}

func get_parent(zmachine *ZMachine, instruction Instruction) (bool, error) {
	object := GetObject(zmachine.Memory, instruction.Operands[0].asObjectId())
	parent := object.Parent()

	instruction.StoreVariable.Write(word(parent))
	return false, nil
}

func get_prop(zmachine *ZMachine, instruction Instruction) (bool, error) {
	objectId := instruction.Operands[0].asObjectId()
	object := GetObject(zmachine.Memory, objectId)
	propertyId := instruction.Operands[1].asPropertyId()

	data := object.Property(propertyId)
	dataAsWord := word(data[0]) << 8
	dataAsWord |= word(data[1])
	instruction.StoreVariable.Write(dataAsWord)
	return false, nil
}

func get_prop_addr(zmachine *ZMachine, instruction Instruction) (bool, error) {
	object := GetObject(zmachine.Memory, instruction.Operands[0].asObjectId())
	propertyId := instruction.Operands[1].asPropertyId()

	instruction.StoreVariable.Write(word(object.GetPropertyDataAddress(propertyId)))
	return false, nil
}

func get_prop_len(zmachine *ZMachine, instruction Instruction) (bool, error) {
	propDataAddr := instruction.Operands[0].asAddress()
	sizeByte := zmachine.Memory.ReadByte(propDataAddr.OffsetBytes(-1))
	length := 0

	if zmachine.Memory.GetVersion() <= 3 {
		length = int(sizeByte>>5) + 1
	} else if (sizeByte >> 7) == 0 {
		length = int((sizeByte>>6)&0b1) + 1
	} else {
		length = int(sizeByte & 0b111111)
	}

	instruction.StoreVariable.Write(word(length))
	return false, nil
}

func get_sibling(zmachine *ZMachine, instruction Instruction) (bool, error) {
	object := GetObject(zmachine.Memory, instruction.Operands[0].asObjectId())
	sibling := object.Sibling()

	instruction.StoreVariable.Write(word(sibling))
	return zmachine.performBranch(instruction.Branch, sibling != 0), nil
}

func inc(zmachine *ZMachine, instruction Instruction) (bool, error) {
	variable := zmachine.getVariable(instruction.Operands[0].asVarNum())

	value := int16(variable.ReadInPlace())
	value++
	variable.WriteInPlace(uint16(value))

	return false, nil
}

func inc_chk(zmachine *ZMachine, instruction Instruction) (bool, error) {
	variable := zmachine.getVariable(instruction.Operands[0].asVarNum())
	condition := int16(instruction.Operands[1].asWord())

	value := int16(variable.ReadInPlace())
	value++
	variable.WriteInPlace(uint16(value))

	return zmachine.performBranch(instruction.Branch, value > condition), nil
}

func insert_obj(zmachine *ZMachine, instruction Instruction) (bool, error) {
	o := instruction.Operands[0].asObjectId()
	d := instruction.Operands[1].asObjectId()

	object := GetObject(zmachine.Memory, o)
	destination := GetObject(zmachine.Memory, d)

	object.SetParent(d)
	object.SetSibling(destination.Child())
	destination.SetChild(o)

	return false, nil
}

func je(zmachine *ZMachine, instruction Instruction) (bool, error) {
	a := instruction.Operands[0].asWord()
	others := make([]word, 0, len(instruction.Operands)-1)
	for i := 1; i < len(instruction.Operands); i++ {
		others = append(others, instruction.Operands[i].asWord())
	}

	return zmachine.performBranch(instruction.Branch, slices.Contains(others, a)), nil
}

func jg(zmachine *ZMachine, instruction Instruction) (bool, error) {
	a := int16(instruction.Operands[0].asWord())
	b := int16(instruction.Operands[1].asWord())

	return zmachine.performBranch(instruction.Branch, a > b), nil
}

func jin(zmachine *ZMachine, instruction Instruction) (bool, error) {
	a := GetObject(zmachine.Memory, instruction.Operands[0].asObjectId())
	b := instruction.Operands[1].asObjectId()

	return zmachine.performBranch(instruction.Branch, a.Parent() == b), nil
}

func jl(zmachine *ZMachine, instruction Instruction) (bool, error) {
	a := int16(instruction.Operands[0].asWord())
	b := int16(instruction.Operands[1].asWord())

	return zmachine.performBranch(instruction.Branch, a < b), nil
}

func jump(zmachine *ZMachine, instruction Instruction) (bool, error) {
	offset := instruction.Operands[0].asInt()

	frame, err := zmachine.Stack.Peek()
	assert.NoError(err, "Error peeking frame stack")
	frame.Counter = instruction.NextAddress.OffsetBytes(offset - 2)
	return true, nil
}

func jz(zmachine *ZMachine, instruction Instruction) (bool, error) {
	a := instruction.Operands[0].asWord()

	return zmachine.performBranch(instruction.Branch, a == 0), nil
}

func load(zmachine *ZMachine, instruction Instruction) (bool, error) {
	variable := zmachine.getVariable(instruction.Operands[0].asVarNum())

	value := variable.ReadInPlace()
	instruction.StoreVariable.Write(value)

	return false, nil
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

func mul(zmachine *ZMachine, instruction Instruction) (bool, error) {
	a := int16(instruction.Operands[0].asWord())
	b := int16(instruction.Operands[1].asWord())

	instruction.StoreVariable.Write(uint16(a * b))
	return false, nil
}

func mod(zmachine *ZMachine, instruction Instruction) (bool, error) {
	a := int16(instruction.Operands[0].asWord())
	b := int16(instruction.Operands[1].asWord())

	assert.NotSame(b, 0, "Cannot mod by 0")

	instruction.StoreVariable.Write(uint16(a % b))
	return false, nil
}

func new_line(zmachine *ZMachine, instruction Instruction) (bool, error) {
	zmachine.Screen.PrintText("\n")
	return false, nil
}

func not(zmachine *ZMachine, instruction Instruction) (bool, error) {
	a := instruction.Operands[0].asWord()

	instruction.StoreVariable.Write(^a)
	return false, nil
}

func or(zmachine *ZMachine, instruction Instruction) (bool, error) {
	a := instruction.Operands[0].asWord()
	b := instruction.Operands[1].asWord()

	instruction.StoreVariable.Write(a | b)
	return false, nil
}

func pop(zmachine *ZMachine, instruction Instruction) (bool, error) {
	frame, err := zmachine.Stack.Peek()
	assert.NoError(err, "Error popping frame stack")
	_, err = frame.Stack.Pop()
	assert.NoError(err, "Error popping local stack")

	return false, nil
}

func print(zmachine *ZMachine, instruction Instruction) (bool, error) {
	zstr := zmachine.Memory.GetZString(instruction.NextAddress)
	next_address := instruction.NextAddress.OffsetBytes(zstr.LenBytes())

	parser := zstring.NewParser(zmachine.Charset, zmachine.Memory.GetAbbreviation)
	str, err := parser.Parse(zstr)
	assert.NoError(err, "Error parsing print ZString")

	zmachine.Screen.PrintText(str)
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

func print_addr(zmachine *ZMachine, instruction Instruction) (bool, error) {
	address := instruction.Operands[0].asAddress()
	zstr := zmachine.Memory.GetZString(address)

	parser := zstring.NewParser(zmachine.Charset, zmachine.Memory.GetAbbreviation)
	str, err := parser.Parse(zstr)
	assert.NoError(err, "Error parsing print ZString")

	zmachine.Screen.PrintText(str)
	if zmachine.Debug {
		fmt.Println()
	}

	return false, nil
}

func print_char(zmachine *ZMachine, instruction Instruction) (bool, error) {
	a := instruction.Operands[0].asByte()

	// TODO: This should convert to ZSCII rather than ASCII/Unicode
	zmachine.Screen.PrintText(string(a))
	if zmachine.Debug {
		fmt.Println()
	}

	return false, nil
}

func print_num(zmachine *ZMachine, instruction Instruction) (bool, error) {
	a := int16(instruction.Operands[0].asInt())

	zmachine.Screen.PrintText(fmt.Sprintf("%v", a))
	if zmachine.Debug {
		fmt.Println()
	}

	return false, nil
}

func print_obj(zmachine *ZMachine, instruction Instruction) (bool, error) {
	o := GetObject(zmachine.Memory, instruction.Operands[0].asObjectId())
	parser := zstring.NewParser(zmachine.Charset, zmachine.Memory.GetAbbreviation)

	zstr := o.ShortName()
	str, err := parser.Parse(zstr)
	assert.NoError(err, "Error parsing object short name")
	zmachine.Screen.PrintText(fmt.Sprintf("%v", str))
	if zmachine.Debug {
		fmt.Println()
	}

	return false, nil
}

func print_paddr(zmachine *ZMachine, instruction Instruction) (bool, error) {
	address := zmachine.Memory.StringPackedAddress(instruction.Operands[0].asWord())
	parser := zstring.NewParser(zmachine.Charset, zmachine.Memory.GetAbbreviation)
	zstr := zmachine.Memory.GetZString(address)
	str, err := parser.Parse(zstr)
	assert.NoError(err, "Error parsing paddr ZString")
	zmachine.Screen.PrintText(str)
	if zmachine.Debug {
		fmt.Println()
	}

	return false, nil
}

func print_ret(zmachine *ZMachine, instruction Instruction) (bool, error) {
	zstr := zmachine.Memory.GetZString(instruction.NextAddress)

	parser := zstring.NewParser(zmachine.Charset, zmachine.Memory.GetAbbreviation)
	str, err := parser.Parse(zstr)
	assert.NoError(err, "Error parsing print ZString")

	zmachine.Screen.PrintText(str)
	zmachine.Screen.PrintText("\n")

	zmachine.endCurrentFrame(1)
	return true, nil
}

func pull(zmachine *ZMachine, instruction Instruction) (bool, error) {
	variable := zmachine.getVariable(instruction.Operands[0].asVarNum())

	frame, err := zmachine.Stack.Peek()
	assert.NoError(err, "Error peeking frame stack")
	value, err := frame.Stack.Pop()
	assert.NoError(err, "Error popping local stack")
	variable.WriteInPlace(value)

	return false, nil
}

func push(zmachine *ZMachine, instruction Instruction) (bool, error) {
	value := instruction.Operands[0].asWord()

	frame, err := zmachine.Stack.Peek()
	assert.NoError(err, "Error peeking frame stack")
	frame.Stack.Push(value)

	return false, nil
}

func put_prop(zmachine *ZMachine, instruction Instruction) (bool, error) {
	object_index := instruction.Operands[0].asObjectId()
	property_index := instruction.Operands[1].asPropertyId()
	value := instruction.Operands[2].asBytes()

	object := GetObject(zmachine.Memory, object_index)
	object.SetProperty(property_index, value)

	return false, nil
}

func quit(zmachine *ZMachine, instruction Instruction) (bool, error) {
	zmachine.Shutdown(0)
	return false, nil
}

func random(zmachine *ZMachine, instruction Instruction) (bool, error) {
	r := int16(instruction.Operands[0].asWord())

	if r > 0 {
		value := zmachine.Random.IntN(int(r)) + 1
		instruction.StoreVariable.Write(word(value))
	} else {
		seed := uint64(r)
		if r == 0 {
			seed = uint64(time.Now().UnixMilli())
		}
		zmachine.Random = rand.New(rand.NewPCG(seed, seed))
		instruction.StoreVariable.Write(0)
	}

	return false, nil
}

func read(zmachine *ZMachine, instruction Instruction) (bool, error) {
	text := instruction.Operands[0].asAddress()
	parse := instruction.Operands[1].asAddress()
	// TODO: In V4, there are 2 additional parameters here

	// TODO: Redisplay Status Line
	str := zmachine.Screen.Read()
	zmachine.Screen.PrintText(str)

	maxTextLength, nextAddress := zmachine.Memory.ReadByteNext(text)
	maxTextLength++ // Initial value is maximum length - 1, so we increment
	str = strings.ToLower(str[:min(len(str), int(maxTextLength))])
	str += "\x00"
	zmachine.Memory.SetBytes(nextAddress, []byte(str))

	zmachine.Screen.PrintText("\n")

	// TODO: Perform lexical analysis
	_, nextAddress = zmachine.Memory.ReadByteNext(parse)
	// maxLexicalWords, nextAddress := zmachine.Memory.ReadByteNext(parse)

	return false, nil
}

func remove_obj(zmachine *ZMachine, instruction Instruction) (bool, error) {
	oid := instruction.Operands[0].asObjectId()
	object := GetObject(zmachine.Memory, oid)

	parentId := object.Parent()
	if parentId == 0 {
		return false, nil
	}

	parent := GetObject(zmachine.Memory, object.Parent())
	object.SetParent(0)

	parentChild := parent.Child()
	if parentChild == oid {
		parent.SetChild(object.Sibling())
		return false, nil
	}

	if parentChild != 0 {
		siblingId := parentChild
		for siblingId != 0 {
			sibling := GetObject(zmachine.Memory, siblingId)
			siblingSibling := sibling.Sibling()
			if siblingSibling == oid {
				sibling.SetSibling(object.Sibling())
				break
			}
			siblingId = siblingSibling
		}
	}

	return false, nil
}

func ret(zmachine *ZMachine, instruction Instruction) (bool, error) {
	value := instruction.Operands[0].asWord()

	zmachine.endCurrentFrame(value)
	return true, nil
}

func ret_popped(zmachine *ZMachine, instruction Instruction) (bool, error) {
	frame, err := zmachine.Stack.Peek()
	assert.NoError(err, "Error peeking frame stack")
	value, err := frame.Stack.Pop()
	assert.NoError(err, "Error popping local stack")

	zmachine.endCurrentFrame(value)
	return true, nil
}

func rfalse(zmachine *ZMachine, instruction Instruction) (bool, error) {
	zmachine.endCurrentFrame(0)
	return true, nil
}

func rtrue(zmachine *ZMachine, instruction Instruction) (bool, error) {
	zmachine.endCurrentFrame(1)
	return true, nil
}

func set_attr(zmachine *ZMachine, instruction Instruction) (bool, error) {
	object := instruction.Operands[0].asObjectId()
	attribute := instruction.Operands[1].asInt()

	GetObject(zmachine.Memory, object).SetAttribute(attribute)

	return false, nil
}

func store(zmachine *ZMachine, instruction Instruction) (bool, error) {
	variable := zmachine.getVariable(instruction.Operands[0].asVarNum())
	value := instruction.Operands[1].asWord()

	variable.WriteInPlace(value)
	return false, nil
}

func storeb(zmachine *ZMachine, instruction Instruction) (bool, error) {
	array := instruction.Operands[0].asAddress()
	byte_index := instruction.Operands[1].asInt()
	value := instruction.Operands[2].asByte()

	address := array.OffsetBytes(byte_index)
	zmachine.Memory.WriteByte(address, value)
	return false, nil
}

func storew(zmachine *ZMachine, instruction Instruction) (bool, error) {
	array := instruction.Operands[0].asAddress()
	word_index := instruction.Operands[1].asInt()
	value := instruction.Operands[2].asWord()

	address := array.OffsetWords(word_index)
	zmachine.Memory.WriteWord(address, value)
	return false, nil
}

func sub(zmachine *ZMachine, instruction Instruction) (bool, error) {
	a := int16(instruction.Operands[0].asWord())
	b := int16(instruction.Operands[1].asWord())
	instruction.StoreVariable.Write(uint16(a - b))
	return false, nil
}

func test(zmachine *ZMachine, instruction Instruction) (bool, error) {
	bitmask := instruction.Operands[0].asWord()
	flags := instruction.Operands[1].asWord()

	return zmachine.performBranch(instruction.Branch, bitmask&flags == flags), nil
}

func test_attr(zmachine *ZMachine, instruction Instruction) (bool, error) {
	object_index := instruction.Operands[0].asObjectId()
	attribute_index := instruction.Operands[1].asInt()

	object := GetObject(zmachine.Memory, object_index)

	return zmachine.performBranch(instruction.Branch, object.HasAttribute(attribute_index)), nil
}

func verify(zmachine *ZMachine, instruction Instruction) (bool, error) {
	// TODO: Implement checksum verification. The below logic should be similar to what is needed
	// but doesn't actually pass CZECH right now. Also see memory.OriginalFileState.

	// fileLength := int(zmachine.Memory.ReadWord(memory.Addr_ROM_W_FileLength))
	// checksum := int(zmachine.Memory.ReadWord(memory.Addr_ROM_W_Checksum))

	// sum := 0
	// filemem, err := zmachine.Memory.OriginalFileState()
	// assert.NoError(err, "Error validating checksum")
	// data := filemem.GetBytes(memory.Address(0x40), fileLength-0x40)
	// for _, b := range data {
	// 	sum += int(b)
	// }

	// sum %= 0x10000

	// return zmachine.performBranch(instruction.Branch, sum == checksum), nil
	return zmachine.performBranch(instruction.Branch, true), nil
}
