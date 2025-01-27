package memory

import (
	"os"
	"slices"

	"github.com/Drakmyth/golang-zmachine/assert"
	"github.com/Drakmyth/golang-zmachine/zstring"
)

type word = uint16

type Memory struct {
	path        string
	version     int
	memory      []byte
	initialized bool
}

func NewMemoryFromFile(path string, initializer func(*Memory)) (*Memory, error) {
	bytes, err := os.ReadFile(path)

	if err != nil {
		return nil, err
	}

	m := Memory{
		path:        path,
		version:     int(bytes[0]),
		memory:      bytes,
		initialized: false,
	}

	initializer(&m)
	m.initialized = true

	return &m, nil
}

// TODO: This might be needed to support checksum verification. See `verify` opcode.
// func (m Memory) OriginalFileState() (*Memory, error) {
// 	return NewMemoryFromFile(m.path, func(memory *Memory) {})
// }

func (m Memory) GetBytes(address Address, length int) []byte {
	assert.True(m.initialized, "Cannot call Memory#GetBytes during memory initialization!")
	return m.memory[address:address.OffsetBytes(length)]
}

func (m Memory) GetBytesNext(address Address, length int) ([]byte, Address) {
	return m.GetBytes(address, length), address.OffsetBytes(length)
}

func (m *Memory) SetBytes(address Address, data []byte) {
	assert.True(m.initialized, "Cannot call Memory#SetBytes during memory initialization!")
	m.memory = slices.Replace(m.memory, int(address), int(address)+len(data), data...)
}

func (m *Memory) SetBytesNext(address Address, data []byte) Address {
	m.SetBytes(address, data)
	return address.OffsetBytes(len(data))
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
	// TODO: Error when writing to ROM, or IROM when initialized is false
	m.memory[address] = data
	return address.OffsetBytes(1)
}

func (m *Memory) WriteWord(address Address, data word) Address {
	// TODO: Error when writing to ROM, or IROM when initialized is false
	next_address := m.WriteByte(address, byte(data>>8))
	next_address = m.WriteByte(next_address, byte(data))
	return next_address
}

func (m Memory) RoutinePackedAddress(address word) Address {
	return m.packedAddress(address, m.ReadWord(Addr_ROM_W_RoutinesOffset))
}

func (m Memory) StringPackedAddress(address word) Address {
	return m.packedAddress(address, m.ReadWord(Addr_ROM_W_StringsOffset))
}

// NOTE: packedAddresses won't be calculated correctly for relative memory, only absolute memory
func (m Memory) packedAddress(address word, offset word) Address {
	switch m.version {
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
