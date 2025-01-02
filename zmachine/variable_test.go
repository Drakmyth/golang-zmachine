package zmachine

import (
	"fmt"
	"testing"

	"github.com/Drakmyth/golang-zmachine/memory"
	"github.com/Drakmyth/golang-zmachine/testassert"
)

func TestVarNum_isLocal_BelowMinimum(t *testing.T) {
	for i := MinVarNum; i < MinLocalVarNum; i++ {
		t.Run(fmt.Sprint(i), func(t *testing.T) {
			testassert.False(t, i.isLocal())
		})
	}
}

func TestVarNum_isLocal_AboveMaximum(t *testing.T) {
	for i := MaxVarNum; i > MaxLocalVarNum; i-- {
		t.Run(fmt.Sprint(i), func(t *testing.T) {
			testassert.False(t, i.isLocal())
		})
	}
}

func TestVarNum_isLocal_InRange(t *testing.T) {
	for i := MinLocalVarNum; i <= MaxLocalVarNum; i++ {
		t.Run(fmt.Sprint(i), func(t *testing.T) {
			testassert.True(t, i.isLocal())
		})
	}
}

func TestVarNum_isGlobal_BelowMinimum(t *testing.T) {
	for i := MinVarNum; i < MinGlobalVarNum; i++ {
		t.Run(fmt.Sprint(i), func(t *testing.T) {
			testassert.False(t, i.isGlobal())
		})
	}
}

func TestVarNum_isGlobal_AboveMaximum(t *testing.T) {
	for i := MaxVarNum; i > MaxGlobalVarNum; i-- {
		t.Run(fmt.Sprint(i), func(t *testing.T) {
			testassert.False(t, i.isGlobal())
		})
	}
}

func TestVarNum_isGlobal_InRange(t *testing.T) {
	// Count down because MaxGlobalVarNum + 1 overflows...
	for i := MaxGlobalVarNum; i >= MinGlobalVarNum; i-- {
		t.Run(fmt.Sprint(i), func(t *testing.T) {
			testassert.True(t, i.isGlobal())
		})
	}
}

func TestVariable_Read_Global(t *testing.T) {
	zmachine, _ := Load("./blank.z3")
	zmachine.Memory.WriteWord(memory.Addr_ROM_A_Globals, 0x01f4)

	global_num := VarNum(0x12)
	address := zmachine.Memory.GetGlobalsAddress().OffsetWords(global_num.asGlobal())
	expected := word(0xbeef)

	zmachine.Memory.WriteWord(address, expected)
	actual := zmachine.getVariable(global_num).Read()

	testassert.Same(t, expected, actual)
}

func TestVariable_Write_Global(t *testing.T) {
	zmachine, _ := Load("./blank.z3")
	zmachine.Memory.WriteWord(memory.Addr_ROM_A_Globals, 0x01f4)

	global_num := VarNum(0x12)
	address := zmachine.Memory.GetGlobalsAddress().OffsetWords(global_num.asGlobal())
	initial := word(0xfeed)
	expected := word(0xbeef)

	zmachine.Memory.WriteWord(address, initial)

	variable := zmachine.getVariable(global_num)
	variable.Write(expected)
	actual := variable.Read()

	testassert.Same(t, expected, actual)
}
