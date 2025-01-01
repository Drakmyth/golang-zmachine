package memory

import "testing"

func TestMemory_ReadWriteByte(t *testing.T) {
	m := NewMemory("./memtest.z3", func(m *Memory) {})

	var target byte = 0x5A
	targetAddr := Address(target)
	assertSame(t, target, m.ReadByte(targetAddr), "Read unexpected value")

	var value byte = 0xFF
	m.WriteByte(targetAddr, value)
	assertSame(t, value, m.ReadByte(targetAddr), "Read unexpected value")
}

func TestMemory_ReadByteNextAddress(t *testing.T) {
	m := NewMemory("./memtest.z3", func(m *Memory) {})

	address := Address(0x5A)
	_, next_address := m.ReadByteNext(address)
	assertSame(t, address.OffsetBytes(1), next_address, "Unexpected Address mismatch")
}

func TestMemory_ReadWriteWord(t *testing.T) {
	m := NewMemory("./memtest.z3", func(m *Memory) {})

	targetAddr := Address(0x5A)
	assertSame(t, 0x5A5B, m.ReadWord(targetAddr), "Read unexpected value")

	var value word = 0xFEFF
	m.WriteWord(targetAddr, value)
	assertSame(t, value, m.ReadWord(targetAddr), "Read unexpected value")
}

func TestMemory_ReadWordNextAddress(t *testing.T) {
	m := NewMemory("./memtest.z3", func(m *Memory) {})

	address := Address(0x5A)
	_, next_address := m.ReadWordNext(address)
	assertSame(t, address.OffsetWords(1), next_address, "Unexpected Address mismatch")
}

func assertSame[E comparable](t *testing.T, expected E, actual E, message string) {
	if expected != actual {
		t.Error(message)
	}
}
