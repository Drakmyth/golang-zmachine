package zmachine

import (
	"fmt"

	"github.com/Drakmyth/golang-zmachine/assert"
	"github.com/Drakmyth/golang-zmachine/memory"
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

func (variable Variable) Read() word {
	zmachine := variable.zmachine

	if variable.isStack() {
		frame, err := zmachine.Stack.Peek()
		assert.NoError(err, "Error peeking frame stack")
		value, err := frame.Stack.Pop()
		assert.NoError(err, "Error popping local stack")
		return value
	} else if variable.isLocal() {
		frame, err := zmachine.Stack.Peek()
		assert.NoError(err, "Error peeking frame stack")
		return frame.Locals[variable.Number.asLocal()]
	} else {
		global := zmachine.Memory.ReadWord(zmachine.Memory.GetGlobalsAddress().OffsetWords(variable.Number.asGlobal()))
		return global
	}
}

func (variable *Variable) Write(value word) {
	zmachine := variable.zmachine

	if variable.isStack() {
		frame, err := zmachine.Stack.Peek()
		assert.NoError(err, "Error peeking frame stack")
		frame.Stack.Push(value)
	} else if variable.isLocal() {
		frame, err := zmachine.Stack.Peek()
		assert.NoError(err, "Error peeking frame stack")
		frame.Locals[variable.Number.asLocal()] = value
	} else {
		zmachine.Memory.WriteWord(zmachine.Memory.GetGlobalsAddress().OffsetWords(variable.Number.asGlobal()), value)
	}
}

func (zmachine *ZMachine) getVariable(index VarNum) Variable {
	return Variable{zmachine, index}
}

func (zmachine *ZMachine) readVariable(address memory.Address) (Variable, memory.Address) {
	varnum_byte, next_address := zmachine.Memory.ReadByteNext(address)
	return Variable{zmachine, VarNum(varnum_byte)}, next_address
}
