package zmachine

import (
	"fmt"
	"testing"
)

func TestWord_highByte(t *testing.T) {
	val := word(0xbeef)
	expected := byte(0xbe)
	got := val.highByte()

	assertEqual(t, expected, got)
}

func TestWord_lowByte(t *testing.T) {
	val := word(0xbeef)
	expected := byte(0xef)
	got := val.lowByte()

	assertEqual(t, expected, got)
}

type offsetTest struct {
	base, offset, expected int
}

var byteAddressOffsetTests = []offsetTest{
	{0x00ff, 1, 0x0100},
	{0x00ff, 2, 0x0101},
	{0x00ff, -1, 0x00fe},
	{0x00ff, -2, 0x00fd},
	{0x00ff, 0, 0x00ff},
}

func TestAddress_offsetBytes(t *testing.T) {
	for _, test := range byteAddressOffsetTests {
		t.Run(fmt.Sprint(test.offset), func(t *testing.T) {
			address := Address(test.base)
			expected := Address(test.expected)
			got := address.offsetBytes(test.offset)

			assertEqual(t, expected, got)
		})
	}
}

var wordAddressOffsetTests = []offsetTest{
	{0x00ff, 1, 0x0101},
	{0x00ff, 2, 0x0103},
	{0x00ff, -1, 0x00fd},
	{0x00ff, -2, 0x00fb},
	{0x00ff, 0, 0x00ff},
}

func TestAddress_offsetWords(t *testing.T) {
	for _, test := range wordAddressOffsetTests {
		t.Run(fmt.Sprint(test.offset), func(t *testing.T) {
			address := Address(test.base)
			expected := Address(test.expected)
			got := address.offsetWords(test.offset)

			assertEqual(t, expected, got)
		})
	}
}

func TestZMachine_readByte(t *testing.T) {
	zmachine := ZMachine{}
	zmachine.init(make([]byte, 100))

	address := Address(0x2f)
	expected := byte(0xbe)
	expected_next_address := address.offsetBytes(1)

	zmachine.Memory[address] = expected
	got, next_address := zmachine.readByte(address)

	assertEqual(t, expected, got)
	assertEqual(t, expected_next_address, next_address)
}

func TestZMachine_writeByte(t *testing.T) {
	zmachine := ZMachine{}
	zmachine.init(make([]byte, 100))

	address := Address(0x2f)
	initial := byte(0xfe)
	expected := byte(0xbe)

	zmachine.Memory[address] = initial
	zmachine.writeByte(expected, address)

	got, _ := zmachine.readByte(address)
	assertEqual(t, expected, got)
}

func TestZMachine_readWord(t *testing.T) {
	zmachine := ZMachine{}
	zmachine.init(make([]byte, 100))

	address := Address(0x2f)
	expected := word(0xbeef)
	expected_next_address := address.offsetWords(1)

	zmachine.Memory[address] = expected.highByte()
	zmachine.Memory[address.offsetBytes(1)] = expected.lowByte()
	got, next_address := zmachine.readWord(address)

	assertEqual(t, expected, got)
	assertEqual(t, expected_next_address, next_address)
}

func TestZMachine_writeWord(t *testing.T) {
	zmachine := ZMachine{}
	zmachine.init(make([]byte, 100))

	address := Address(0x2f)
	initial := word(0xfeed)
	expected := word(0xbeef)

	zmachine.Memory[address] = initial.highByte()
	zmachine.Memory[address.offsetBytes(1)] = initial.lowByte()
	zmachine.writeWord(expected, address)

	got, _ := zmachine.readWord(address)
	assertEqual(t, expected, got)
}
