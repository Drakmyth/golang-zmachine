package memory

import (
	"os"

	"github.com/Drakmyth/golang-zmachine/assert"
)

type word = uint16
type Address word

// Should only be used in the Abbreviations table
func WordAddress(address word) Address {
	return Address(2 * address)
}

func (m Memory) RoutinePackedAddress(address word) Address {
	return packedAddress(address, m.GetVersion(), m.ReadWord(Addr_ROM_W_RoutinesOffset))
}

func (m Memory) StringPackedAddress(address word) Address {
	return packedAddress(address, m.GetVersion(), m.ReadWord(Addr_ROM_W_StringsOffset))
}

func packedAddress(address word, version byte, offset uint16) Address {
	assert.Between(1, 9, version, "Unknown Version.")
	assert.NotContains([]byte{6, 7}, version, "PackedAddress not implemented in Version 6/7. Call OffsetPackedAddress instead.")

	switch version {
	case 1, 2, 3:
		return Address(2 * address)
	case 4, 5:
		return Address(4 * address)
	case 6, 7:
		return Address(4*address + 8*offset)
	case 8:
		return Address(8 * address)
	}
	panic("Unknown version")
}

func (address Address) OffsetBytes(amount int) Address {
	return Address(int(address) + amount)
}

func (address Address) OffsetWords(amount int) Address {
	return Address(int(address) + 2*amount)
}

type Memory struct {
	memory      []byte
	initialized bool
}

// NOTE: The warning here is due to the native Go stdmethods checker being over-eager on
// ensuring interfaces are being implemented correctly. As of Go 1.23.4 there is no way
// to configure or override this behavior and the "standard" signature doesn't meet my
// needs, but I'm neither willing to change the name of the method to something less
// representative of its behavior nor to disable the stdmethods check entirely.
func (m Memory) ReadByte(address Address) byte {
	return m.memory[address]
}

func (m Memory) ReadByteNext(address Address) (byte, Address) {
	return m.ReadByte(address), address.OffsetBytes(1)
}

func (m Memory) ReadWord(address Address) word {
	high, next_address := m.ReadByteNext(address)
	low := m.ReadByte(next_address)
	return (word(high) << 8) | word(low)
}

func (m Memory) ReadWordNext(address Address) (word, Address) {
	return m.ReadWord(address), address.OffsetWords(1)
}

// NOTE: The warning here is due to the native Go stdmethods checker being over-eager on
// ensuring interfaces are being implemented correctly. As of Go 1.23.4 there is no way
// to configure or override this behavior and the "standard" signature doesn't meet my
// needs, but I'm neither willing to change the name of the method to something less
// representative of its behavior nor to disable the stdmethods check entirely.
func (m *Memory) WriteByte(address Address, data byte) Address {
	// TODO: Error when writing to IROM when initialized is false or ROM
	m.memory[address] = data
	return address.OffsetBytes(1)
}

func (m *Memory) WriteWord(address Address, data word) Address {
	// TODO: Error when writing to IROM when initialized is false or ROM
	m.WriteByte(address, byte(data>>8))
	m.WriteByte(address.OffsetBytes(1), byte(data))
	return address.OffsetWords(1)
}

func NewMemory(path string, handler func(*Memory)) *Memory {
	bytes, err := os.ReadFile(path)
	assert.NoError(err, "Error loading file.")
	m := Memory{
		memory:      bytes,
		initialized: false,
	}
	handler(&m)
	m.initialized = true
	return &m
}
