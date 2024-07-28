package zmachine

import (
	"bytes"
	"encoding/binary"
)

// TODO: In versions 4+ all these fields along with the length of the prop default table get bigger
type Object struct {
	Attributes     uint32
	Parent         byte
	Sibling        byte
	Child          byte
	PropertiesAddr Address
}

type PropertiesTable struct {
	Name       string
	Properties map[byte][]byte
}

func (object Object) hasAttribute(index int) bool {
	return (object.Attributes>>(31-index))&0b1 == 1
}

func (zmachine ZMachine) getObject(index int) Object {
	tree_address := zmachine.Header.ObjectsAddr.offsetWords(31)
	object_address := tree_address.offsetBytes(9 * (index - 1)) // object zero is not stored in memory, so we shift the index back one
	object_memory := bytes.NewBuffer(zmachine.Memory[object_address : object_address+9])
	object := Object{}
	binary.Read(object_memory, binary.BigEndian, &object)
	return object
}

func (zmachine ZMachine) readProperties(address Address) (PropertiesTable, Address) {
	_, next_address := zmachine.readByte(address) // The length byte is redundant since it's part of the name string
	var object_name string
	object_name, next_address = zmachine.readZString(next_address)

	// TODO: Version 4+ also supports 2-byte sizes and has a slightly different size-byte structure
	var prop_size byte
	prop_size, next_address = zmachine.readByte(next_address)
	table := PropertiesTable{Name: object_name, Properties: make(map[byte][]byte)}
	for prop_size != 0 {
		prop_length := int(prop_size>>5) + 1
		prop_number := prop_size & 0b11111
		prop_data := zmachine.Memory[next_address:next_address.offsetBytes(prop_length)]
		table.Properties[prop_number] = prop_data
		prop_size, next_address = zmachine.readByte(next_address.offsetBytes(prop_length))
	}

	return table, next_address
}

func (zmachine ZMachine) getPropertyDefault(index int) word {
	address := zmachine.Header.ObjectsAddr
	val, _ := zmachine.readWord(address.offsetWords(index))
	return val
}
