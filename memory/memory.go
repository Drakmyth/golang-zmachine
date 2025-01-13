package memory

import (
	"bytes"
	"os"

	"github.com/Drakmyth/golang-zmachine/assert"
	"github.com/Drakmyth/golang-zmachine/zstring"
)

type word = uint16

func (m Memory) RoutinePackedAddress(address word) Address {
	return packedAddress(address, m.GetVersion(), m.ReadWord(Addr_ROM_W_RoutinesOffset))
}

func (m Memory) StringPackedAddress(address word) Address {
	return packedAddress(address, m.GetVersion(), m.ReadWord(Addr_ROM_W_StringsOffset))
}

func packedAddress(address word, version int, offset uint16) Address {
	assert.Between(1, 9, version, "Unknown Version.")

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

type Memory struct {
	memory      []byte
	initialized bool
}

func (m Memory) GetBytes(address Address, length int) []byte {
	return m.memory[address : int(address)+length]
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

func (m Memory) GetZString(address Address) zstring.ZString {
	var b byte
	next_address := address
	foundEnd := false
	length := 0

	for !foundEnd {
		b, next_address = m.ReadByteNext(next_address)
		_, next_address = m.ReadByteNext(next_address)
		length += 2

		if b>>7 == 1 {
			foundEnd = true
		}
	}

	return m.GetBytes(address, length)
}

func (m *Memory) GetAbbreviation(bank int, index int) zstring.ZString {
	abbr_entry := m.GetAbbreviationsAddress().OffsetWords(int((32*(bank-1) + index)))
	address := m.ReadWord(abbr_entry)
	abbreviation := m.GetZString(Address(address * 2))
	return abbreviation
}

func (m Memory) GetAlphabet() []rune {
	alphabetAddress := Address(m.ReadWord(Addr_ROM_A_AlphabetTable))

	if alphabetAddress == 0 {
		return zstring.GetDefaultAlphabet(m.GetVersion())
	}

	return bytes.Runes(m.GetBytes(Addr_ROM_A_AlphabetTable, 78))
}
