package zmachine

import (
	"github.com/Drakmyth/golang-zmachine/memory"
)

// TODO: In versions 4+ all these fields along with the length of the prop default table get bigger
type Object struct {
	Attributes     uint32
	Parent         byte
	Sibling        byte
	Child          byte
	PropertiesAddr memory.Address
}

type PropertiesTable struct {
	Name       string
	Properties map[byte][]byte
}

func (object Object) hasAttribute(index int) bool {
	return (object.Attributes>>(31-index))&0b1 == 1
}

func (zmachine ZMachine) getObject(index int) Object {
	tree_address := zmachine.Memory.GetObjectsAddress().OffsetWords(31)
	object_address := tree_address.OffsetBytes(9 * (index - 1)) // object zero is not stored in memory, so we shift the index back one

	temp, next_address := zmachine.Memory.ReadWordNext(object_address)
	attributes := uint32(temp) << 16
	temp, next_address = zmachine.Memory.ReadWordNext(next_address)
	attributes |= uint32(temp)
	parent, next_address := zmachine.Memory.ReadByteNext(next_address)
	sibling, next_address := zmachine.Memory.ReadByteNext(next_address)
	child, next_address := zmachine.Memory.ReadByteNext(next_address)
	propsAddr := memory.Address(zmachine.Memory.ReadWord(next_address))
	object := Object{
		Attributes:     attributes,
		Parent:         parent,
		Sibling:        sibling,
		Child:          child,
		PropertiesAddr: propsAddr,
	}

	return object
}

func (zmachine ZMachine) readProperties(address memory.Address) (PropertiesTable, memory.Address) {
	next_address := address.OffsetBytes(1) // The length byte is redundant since it's part of the name string
	var object_name string
	object_name, next_address = zmachine.readZString(next_address)

	// TODO: Version 4+ also supports 2-byte sizes and has a slightly different size-byte structure
	var prop_size byte
	prop_size, next_address = zmachine.Memory.ReadByteNext(next_address)
	table := PropertiesTable{Name: object_name, Properties: make(map[byte][]byte)}
	for prop_size != 0 {
		prop_length := int(prop_size>>5) + 1
		prop_number := prop_size & 0b11111
		var prop_data []byte = make([]byte, prop_length)
		for i := 0; i < prop_length; i++ {
			var next_byte byte
			next_byte, next_address = zmachine.Memory.ReadByteNext(next_address)
			prop_data[i] = next_byte
		}
		table.Properties[prop_number] = prop_data
		prop_size, next_address = zmachine.Memory.ReadByteNext(next_address)
	}

	return table, next_address
}

func (zmachine ZMachine) getPropertyDefault(index int) word {
	address := zmachine.Memory.GetObjectsAddress()
	val := zmachine.Memory.ReadWord(address.OffsetWords(index))
	return val
}
