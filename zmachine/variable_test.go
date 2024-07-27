package zmachine

import (
	"fmt"
	"testing"
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

func TestZMachine_getGlobal(t *testing.T) {
	zmachine := ZMachine{}
	zmachine.init(make([]byte, 1000))
	zmachine.Header.GlobalsAddr = Address(0x01f4)

	global_num := 2
	address := zmachine.Header.GlobalsAddr.offsetWords(global_num)
	expected := word(0xbeef)

	zmachine.Memory[address] = expected.highByte()
	zmachine.Memory[address.offsetBytes(1)] = expected.lowByte()
	got := zmachine.getGlobal(global_num)

	assertEqual(t, expected, got)
}

func TestZMachine_setGlobal(t *testing.T) {
	zmachine := ZMachine{}
	zmachine.init(make([]byte, 1000))
	zmachine.Header.GlobalsAddr = Address(0x01f4)

	global_num := 2
	address := zmachine.Header.GlobalsAddr.offsetWords(global_num)
	initial := word(0xfeed)
	expected := word(0xbeef)

	zmachine.Memory[address] = initial.highByte()
	zmachine.Memory[address.offsetBytes(1)] = initial.lowByte()

	zmachine.setGlobal(expected, global_num)
	got := zmachine.getGlobal(global_num)

	assertEqual(t, expected, got)
}
