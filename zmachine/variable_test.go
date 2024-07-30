package zmachine

import (
	"fmt"
	"testing"

	"github.com/Drakmyth/golang-zmachine/zmachine/internal/memory"
)

func TestVarNum_isLocal_BelowMinimum(t *testing.T) {
	for i := MinVarNum; i < MinLocalVarNum; i++ {
		t.Run(fmt.Sprint(i), func(t *testing.T) {
			assertFalse(t, i.isLocal())
		})
	}
}

func TestVarNum_isLocal_AboveMaximum(t *testing.T) {
	for i := MaxVarNum; i > MaxLocalVarNum; i-- {
		t.Run(fmt.Sprint(i), func(t *testing.T) {
			assertFalse(t, i.isLocal())
		})
	}
}

func TestVarNum_isLocal_InRange(t *testing.T) {
	for i := MinLocalVarNum; i <= MaxLocalVarNum; i++ {
		t.Run(fmt.Sprint(i), func(t *testing.T) {
			assertTrue(t, i.isLocal())
		})
	}
}

func TestVarNum_isGlobal_BelowMinimum(t *testing.T) {
	for i := MinVarNum; i < MinGlobalVarNum; i++ {
		t.Run(fmt.Sprint(i), func(t *testing.T) {
			assertFalse(t, i.isGlobal())
		})
	}
}

func TestVarNum_isGlobal_AboveMaximum(t *testing.T) {
	for i := MaxVarNum; i > MaxGlobalVarNum; i-- {
		t.Run(fmt.Sprint(i), func(t *testing.T) {
			assertFalse(t, i.isGlobal())
		})
	}
}

func TestVarNum_isGlobal_InRange(t *testing.T) {
	// Count down because MaxGlobalVarNum + 1 overflows...
	for i := MaxGlobalVarNum; i >= MinGlobalVarNum; i-- {
		t.Run(fmt.Sprint(i), func(t *testing.T) {
			assertTrue(t, i.isGlobal())
		})
	}
}

func TestVariable_Read_Global(t *testing.T) {
	zmachine := ZMachine{}
	zmachine.init(make([]byte, 1000))
	zmachine.Header.GlobalsAddr = memory.Address(0x01f4)

	global_num := VarNum(0x12)
	address := zmachine.Header.GlobalsAddr.OffsetWords(global_num.asGlobal())
	expected := memory.Word(0xbeef)

	zmachine.Memory[address] = expected.HighByte()
	zmachine.Memory[address.OffsetBytes(1)] = expected.LowByte()
	got := zmachine.getVariable(global_num).Read()

	AssertEqual(t, expected, got)
}

func TestVariable_Write_Global(t *testing.T) {
	zmachine := ZMachine{}
	zmachine.init(make([]byte, 1000))
	zmachine.Header.GlobalsAddr = memory.Address(0x01f4)

	global_num := VarNum(0x12)
	address := zmachine.Header.GlobalsAddr.OffsetWords(global_num.asGlobal())
	initial := memory.Word(0xfeed)
	expected := memory.Word(0xbeef)

	zmachine.Memory[address] = initial.HighByte()
	zmachine.Memory[address.OffsetBytes(1)] = initial.LowByte()

	variable := zmachine.getVariable(global_num)
	variable.Write(expected)
	got := variable.Read()

	AssertEqual(t, expected, got)
}
