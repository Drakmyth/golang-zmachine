package memory_test

import (
	"fmt"
	"testing"

	"github.com/Drakmyth/golang-zmachine/zmachine"
	"github.com/Drakmyth/golang-zmachine/zmachine/internal/memory"
)

func TestWord_HighByte(t *testing.T) {
	val := memory.Word(0xbeef)
	expected := byte(0xbe)
	got := val.HighByte()

	zmachine.AssertEqual(t, expected, got)
}

func TestWord_LowByte(t *testing.T) {
	val := memory.Word(0xbeef)
	expected := byte(0xef)
	got := val.LowByte()

	zmachine.AssertEqual(t, expected, got)
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

func TestAddress_OffsetBytes(t *testing.T) {
	for _, test := range byteAddressOffsetTests {
		t.Run(fmt.Sprint(test.offset), func(t *testing.T) {
			address := memory.Address(test.base)
			expected := memory.Address(test.expected)
			got := address.OffsetBytes(test.offset)

			zmachine.AssertEqual(t, expected, got)
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

func TestAddress_OffsetWords(t *testing.T) {
	for _, test := range wordAddressOffsetTests {
		t.Run(fmt.Sprint(test.offset), func(t *testing.T) {
			address := memory.Address(test.base)
			expected := memory.Address(test.expected)
			got := address.OffsetWords(test.offset)

			zmachine.AssertEqual(t, expected, got)
		})
	}
}
