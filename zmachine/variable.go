package zmachine

import (
	"fmt"

	"github.com/Drakmyth/golang-zmachine/zmachine/internal/memory"
)

type VarNum uint8

const MinVarNum VarNum = 0x00
const MaxVarNum VarNum = 0xff
const StackVarNum VarNum = 0x00
const MinLocalVarNum VarNum = 0x01
const MaxLocalVarNum VarNum = 0x0f
const MinGlobalVarNum VarNum = 0x10
const MaxGlobalVarNum VarNum = 0xff

func (varnum VarNum) String() string {
	if varnum == 0 {
		return "sp"
	} else if varnum.isLocal() {
		return fmt.Sprintf("local%d", varnum.asLocal())
	} else {
		return fmt.Sprintf("g%d", varnum.asGlobal())
	}
}

func (varnum VarNum) isLocal() bool {
	return MinLocalVarNum <= varnum && varnum <= MaxLocalVarNum
}

func (varnum VarNum) isGlobal() bool {
	return MinGlobalVarNum <= varnum && varnum <= MaxGlobalVarNum
}

func (varnum VarNum) asLocal() int {
	return int(varnum - MinLocalVarNum)
}

func (varnum VarNum) asGlobal() int {
	return int(varnum - MinGlobalVarNum)
}

func (zmachine ZMachine) getGlobal(index int) memory.Word {
	global, _ := zmachine.readWord(zmachine.Header.GlobalsAddr.OffsetWords(index))
	return global
}

func (zmachine *ZMachine) setGlobal(value memory.Word, index int) {
	zmachine.writeWord(value, zmachine.Header.GlobalsAddr.OffsetWords(index))
}

func (zmachine ZMachine) readVariable(index VarNum) memory.Word {
	if index == 0 {
		return zmachine.Stack.Peek().Stack.Pop()
	} else if index.isLocal() {
		return zmachine.Stack.Peek().Locals[index.asLocal()]
	} else {
		return zmachine.getGlobal(index.asGlobal())
	}
}

func (zmachine *ZMachine) writeVariable(value memory.Word, index VarNum) {
	if index == 0 {
		zmachine.Stack.Peek().Stack.Push(value)
	} else if index.isLocal() {
		zmachine.Stack.Peek().Locals[index.asLocal()] = value
	} else {
		zmachine.setGlobal(value, index.asGlobal())
	}
}

func (zmachine ZMachine) readVarNum(address memory.Address) (VarNum, memory.Address) {
	varnum_byte, next_address := zmachine.readByte(address)
	return VarNum(varnum_byte), next_address
}
