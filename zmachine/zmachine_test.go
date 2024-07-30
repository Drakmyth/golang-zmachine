package zmachine

import (
	"testing"

	"github.com/Drakmyth/golang-zmachine/zmachine/internal/memory"
)

func TestZMachine_readByte(t *testing.T) {
	zmachine := ZMachine{}
	zmachine.init(make([]byte, 100))

	address := memory.Address(0x2f)
	expected := byte(0xbe)
	expected_next_address := address.OffsetBytes(1)

	zmachine.Memory[address] = expected
	got, next_address := zmachine.readByte(address)

	AssertEqual(t, expected, got)
	AssertEqual(t, expected_next_address, next_address)
}

func TestZMachine_writeByte(t *testing.T) {
	zmachine := ZMachine{}
	zmachine.init(make([]byte, 100))

	address := memory.Address(0x2f)
	initial := byte(0xfe)
	expected := byte(0xbe)

	zmachine.Memory[address] = initial
	zmachine.writeByte(expected, address)

	got, _ := zmachine.readByte(address)
	AssertEqual(t, expected, got)
}

func TestZMachine_readWord(t *testing.T) {
	zmachine := ZMachine{}
	zmachine.init(make([]byte, 100))

	address := memory.Address(0x2f)
	expected := memory.Word(0xbeef)
	expected_next_address := address.OffsetWords(1)

	zmachine.Memory[address] = expected.HighByte()
	zmachine.Memory[address.OffsetBytes(1)] = expected.LowByte()
	got, next_address := zmachine.readWord(address)

	AssertEqual(t, expected, got)
	AssertEqual(t, expected_next_address, next_address)
}

func TestZMachine_writeWord(t *testing.T) {
	zmachine := ZMachine{}
	zmachine.init(make([]byte, 100))

	address := memory.Address(0x2f)
	initial := memory.Word(0xfeed)
	expected := memory.Word(0xbeef)

	zmachine.Memory[address] = initial.HighByte()
	zmachine.Memory[address.OffsetBytes(1)] = initial.LowByte()
	zmachine.writeWord(expected, address)

	got, _ := zmachine.readWord(address)
	AssertEqual(t, expected, got)
}
