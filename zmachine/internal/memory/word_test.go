package memory_test

import (
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
