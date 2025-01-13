package memory

import (
	"encoding/binary"
	"fmt"

	"github.com/Drakmyth/golang-zmachine/assert"
	"github.com/Drakmyth/golang-zmachine/zstring"
)

type ObjectId word
type PropertyId byte

const (
	idx_Attributes int = iota // uint32
	_
	_
	_
	idxV1_Parent         // byte
	idxV1_Sibling        // byte
	idxV1_Child          // byte
	idxV1_PropertiesAddr // word
	_
)

const (
	_ int = iota // uint48, same start as idx_Attributes
	_
	_
	_
	_
	_
	idxV4_Parent // word
	_
	idxV4_Sibling // word
	_
	idxV4_Child // word
	_
	idxV4_PropertiesAddr // word
	_
)

type Object struct {
	id             ObjectId
	data           []byte
	version        int
	shortName      zstring.ZString
	propertiesData []byte
}

func (o Object) HasAttribute(index int) bool {
	maxAttributes := 32
	if o.version > 3 {
		maxAttributes = 48
	}
	assert.LessThan(maxAttributes, index, "Invalid attribute index")

	bytesToSkip := index / 8
	newIndex := index % 8

	attributeByte := o.data[idx_Attributes+bytesToSkip]
	return (attributeByte>>newIndex)&0b1 == 1
}

func (o Object) SetAttribute(index int) {
	maxAttributes := 32
	if o.version > 3 {
		maxAttributes = 48
	}
	assert.LessThan(maxAttributes, index, "Invalid attribute index")

	bytesToSkip := index / 8
	newIndex := index % 8

	attributeByte := o.data[idx_Attributes+bytesToSkip]
	attributeByte = attributeByte | (0b1 << (7 - newIndex))
	o.data[idx_Attributes+bytesToSkip] = attributeByte
}

func (o Object) GetParent() ObjectId {
	if o.version <= 3 {
		return ObjectId(o.data[idxV1_Parent])
	} else {
		return ObjectId(binary.BigEndian.Uint16(o.data[idxV4_Parent : idxV4_Parent+1]))
	}
}

func (o Object) GetSibling() ObjectId {
	if o.version <= 3 {
		return ObjectId(o.data[idxV1_Sibling])
	} else {
		return ObjectId(binary.BigEndian.Uint16(o.data[idxV4_Sibling : idxV4_Sibling+1]))
	}
}

func (o Object) GetChild() ObjectId {
	if o.version <= 3 {
		return ObjectId(o.data[idxV1_Child])
	} else {
		return ObjectId(binary.BigEndian.Uint16(o.data[idxV4_Child : idxV4_Child+1]))
	}
}

func (o *Object) SetParent(parent ObjectId) {
	if o.version <= 3 {
		o.data[idxV1_Parent] = byte(parent)
	} else {
		binary.BigEndian.PutUint16(o.data[idxV4_Parent:idxV4_Parent+1], word(parent))
	}
}

func (o *Object) SetSibling(sibling ObjectId) {
	if o.version <= 3 {
		o.data[idxV1_Sibling] = byte(sibling)
	} else {
		binary.BigEndian.PutUint16(o.data[idxV4_Parent:idxV4_Sibling+1], word(sibling))
	}
}

func (o *Object) SetChild(child ObjectId) {
	if o.version <= 3 {
		o.data[idxV1_Child] = byte(child)
	} else {
		binary.BigEndian.PutUint16(o.data[idxV4_Child:idxV4_Child+1], word(child))
	}
}

func (o *Object) SetProperty(property PropertyId, value word) {
	next_address := 0
	header_length := o.propertiesData[next_address]
	next_address++ // Need to advance the address pointer manually as we read

	next_address += int(header_length) * 2 // Skip the property table header string
	property_size := o.propertiesData[next_address]
	next_address++

	// TODO: V4+ can have 1 or 2 byte sizes
	for property_size != 0 {
		prop_number := PropertyId(property_size & 0b11111)
		prop_length := (property_size >> 5) + 1
		if prop_length == 0 {
			prop_length = 64
		}

		if prop_number == property {
			switch prop_length {
			case 1:
				o.propertiesData[next_address] = byte(value)
			case 2:
				o.propertiesData[next_address] = byte(value >> 8)
				o.propertiesData[next_address+1] = byte(value)
			default:
				panic(fmt.Sprintf("unsupported put_prop data length: %d", prop_length))
			}
			return
		} else if prop_number > property {
			next_address += int(prop_length)
			property_size = o.propertiesData[next_address]
			next_address++
			continue
		} else {
			break
		}
	}

	panic(fmt.Sprintf("Property %d not found on Object %d", property, o.id))
}

func decodePropertySizeByte(size byte) (byte, PropertyId) {
	length := (size >> 5) + 1
	property := size & 0b11111
	return length, PropertyId(property)
}

func getPropertyDefaultsTableLength(version int) int {
	if version <= 3 {
		return 31
	}

	return 63
}

func getObjectSize(version int) int {
	if version <= 3 {
		return 9
	}

	return 14
}

func getMaxObjectCount(version int) int {
	if version <= 3 {
		return 255
	}

	return 65535
}

func (m Memory) GetObject(objectId ObjectId) Object {
	objectNumber := objectId - 1 // Object 0 is not stored in memory, so we shift the index back one

	objectTableAddress := m.GetObjectsAddress()
	version := m.GetVersion()
	propertyDefaultsTableLength := getPropertyDefaultsTableLength(version)
	objectTreeAddress := objectTableAddress.OffsetWords(propertyDefaultsTableLength)
	objectSize := getObjectSize(version)
	// maxObjectCount := getMaxObjectCount(version)

	objectAddress := objectTreeAddress.OffsetBytes(objectSize * int(objectNumber))
	propertyTableAddressOffset := idxV1_PropertiesAddr
	if version > 3 {
		propertyTableAddressOffset = idxV4_PropertiesAddr
	}
	propertyTableAddress := Address(m.ReadWord(objectAddress.OffsetBytes(propertyTableAddressOffset)))
	shortName, propertiesData := m.getPropertyTableData(propertyTableAddress)

	object := Object{
		id:             objectId,
		data:           m.memory[objectAddress:objectAddress.OffsetBytes(objectSize)],
		version:        version,
		shortName:      shortName,
		propertiesData: propertiesData,
	}

	return object
}

func (m *Memory) getPropertyTableData(propertyTableAddress Address) (zstring.ZString, []byte) {
	headerLength, next_address := m.ReadByteNext(propertyTableAddress)
	headerLengthByteCount := int(headerLength) * 2
	addrAfterShortName := next_address.OffsetBytes(headerLengthByteCount)
	shortName := zstring.ZString(m.memory[next_address:addrAfterShortName])
	next_address = addrAfterShortName
	property_size, next_address := m.ReadByteNext(next_address)

	for property_size != 0 {
		prop_length := (property_size >> 5) + 1
		if prop_length == 0 {
			prop_length = 64
		}
		next_address = next_address.OffsetBytes(int(prop_length))
		property_size, next_address = m.ReadByteNext(next_address)
	}

	return shortName, m.GetBytes(propertyTableAddress, int(next_address)-int(propertyTableAddress))
}

// func (zmachine ZMachine) readProperties(address memory.Address) (PropertiesTable, memory.Address) {
// 	next_address := address.OffsetBytes(1) // The length byte is redundant since it's part of the name string
// 	var object_name string
// 	object_name, next_address = zmachine.readZString(next_address)

// 	// TODO: Version 4+ also supports 2-byte sizes and has a slightly different size-byte structure
// 	var prop_size byte
// 	prop_size, next_address = zmachine.Memory.ReadByteNext(next_address)
// 	table := PropertiesTable{Name: object_name, Properties: make(map[byte][]byte)}
// 	for prop_size != 0 {
// 		prop_length := int(prop_size>>5) + 1
// 		prop_number := prop_size & 0b11111
// 		var prop_data []byte = make([]byte, prop_length)
// 		for i := 0; i < prop_length; i++ {
// 			var next_byte byte
// 			next_byte, next_address = zmachine.Memory.ReadByteNext(next_address)
// 			prop_data[i] = next_byte
// 		}
// 		table.Properties[prop_number] = prop_data
// 		prop_size, next_address = zmachine.Memory.ReadByteNext(next_address)
// 	}

// 	return table, next_address
// }

// func (zmachine ZMachine) getPropertyDefault(index int) word {
// 	address := zmachine.Memory.GetObjectsAddress()
// 	val := zmachine.Memory.ReadWord(address.OffsetWords(index))
// 	return val
// }
