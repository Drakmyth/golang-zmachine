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

func (varnum VarNum) isStack() bool {
	return varnum == 0
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

type Variable struct {
	zmachine *ZMachine
	Number   VarNum
}

func (variable Variable) isStack() bool {
	return variable.Number.isStack()
}

func (variable Variable) isLocal() bool {
	return variable.Number.isLocal()
}

func (variable Variable) isGlobal() bool {
	return variable.Number.isGlobal()
}

func (variable Variable) Read() memory.Word {
	zmachine := variable.zmachine

	if variable.isStack() {
		return zmachine.Stack.Peek().Stack.Pop()
	} else if variable.isLocal() {
		return zmachine.Stack.Peek().Locals[variable.Number.asLocal()]
	} else {
		global, _ := zmachine.readWord(zmachine.Header.GlobalsAddr.OffsetWords(variable.Number.asGlobal()))
		return global
	}
}

func (variable *Variable) Write(value memory.Word) {
	zmachine := variable.zmachine

	if variable.isStack() {
		zmachine.Stack.Peek().Stack.Push(value)
	} else if variable.isLocal() {
		zmachine.Stack.Peek().Locals[variable.Number.asLocal()] = value
	} else {
		zmachine.writeWord(value, zmachine.Header.GlobalsAddr.OffsetWords(variable.Number.asGlobal()))
	}
}

func (zmachine *ZMachine) getVariable(index VarNum) Variable {
	return Variable{zmachine, index}
}

func (zmachine *ZMachine) readVariable(address memory.Address) (Variable, memory.Address) {
	varnum_byte, next_address := zmachine.readByte(address)
	return Variable{zmachine, VarNum(varnum_byte)}, next_address
}
