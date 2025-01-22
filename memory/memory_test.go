package memory

import (
	"testing"

	"github.com/Drakmyth/golang-zmachine/testassert"
)

func TestMemory_ReadWriteByte(t *testing.T) {
	m, err := NewMemoryFromFile("./memtest.z3", func(m *Memory) {})
	testassert.NoError(t, err)

	var target byte = 0x5A
	targetAddr := Address(target)
	testassert.Same(t, target, m.ReadByte(targetAddr))

	var value byte = 0xFF
	m.WriteByte(targetAddr, value)
	testassert.Same(t, value, m.ReadByte(targetAddr))
}

func TestMemory_ReadByteNextAddress(t *testing.T) {
	m, err := NewMemoryFromFile("./memtest.z3", func(m *Memory) {})
	testassert.NoError(t, err)

	address := Address(0x5A)
	_, next_address := m.ReadByteNext(address)
	testassert.Same(t, address.OffsetBytes(1), next_address)
}

func TestMemory_ReadWriteWord(t *testing.T) {
	m, err := NewMemoryFromFile("./memtest.z3", func(m *Memory) {})
	testassert.NoError(t, err)

	targetAddr := Address(0x5A)
	testassert.Same(t, 0x5A5B, m.ReadWord(targetAddr))

	var value word = 0xFEFF
	m.WriteWord(targetAddr, value)
	testassert.Same(t, value, m.ReadWord(targetAddr))
}

func TestMemory_ReadWordNextAddress(t *testing.T) {
	m, err := NewMemoryFromFile("./memtest.z3", func(m *Memory) {})
	testassert.NoError(t, err)

	address := Address(0x5A)
	_, next_address := m.ReadWordNext(address)
	testassert.Same(t, address.OffsetWords(1), next_address)
}
