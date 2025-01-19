package zmachine

import (
	"github.com/Drakmyth/golang-zmachine/assert"
	"github.com/Drakmyth/golang-zmachine/memory"
	"github.com/Drakmyth/golang-zmachine/zstring"
)

type ObjectId word
type PropertyId byte

// V1-V3 offset constants
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

// V4+ offset constants
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

type object struct {
	mem     *memory.Memory
	address memory.Address
}

func GetObject(mem *memory.Memory, oid ObjectId) *object {
	objectTableAddr := mem.GetObjectsAddress()

	propCount := 31
	objectSize := 9
	if mem.GetVersion() > 3 {
		propCount = 63
		objectSize = 14
	}

	objectTreeAddr := objectTableAddr.OffsetWords(propCount)

	objectAddr := objectTreeAddr.OffsetBytes(objectSize * int(oid-1)) // Object IDs start at 1

	return &object{
		mem:     mem,
		address: objectAddr,
	}
}

func (o object) HasAttribute(index int) bool {
	maxAttributes := 32
	if o.mem.GetVersion() > 3 {
		maxAttributes = 48
	}
	assert.LessThan(maxAttributes, index, "Invalid attribute index")

	bytesToSkip := index / 8
	newIndex := index % 8

	attributeByteAddr := o.address.OffsetBytes(idx_Attributes).OffsetBytes(bytesToSkip)
	attributeByte := o.mem.ReadByte(attributeByteAddr)
	hasAttribute := (attributeByte >> (7 - newIndex)) & 0b1
	return hasAttribute == 1
}

func (o *object) ClearAttribute(index int) {
	maxAttributes := 32
	if o.mem.GetVersion() > 3 {
		maxAttributes = 48
	}
	assert.LessThan(maxAttributes, index, "Invalid attribute index")

	bytesToSkip := index / 8
	newIndex := index % 8

	attributeByteAddr := o.address.OffsetBytes(idx_Attributes).OffsetBytes(bytesToSkip)
	attributeByte := o.mem.ReadByte(attributeByteAddr)
	mask := ^byte(0b1 << (7 - newIndex))
	attributeByte = attributeByte & mask
	o.mem.WriteByte(attributeByteAddr, attributeByte)
}

func (o *object) SetAttribute(index int) {
	maxAttributes := 32
	if o.mem.GetVersion() > 3 {
		maxAttributes = 48
	}
	assert.LessThan(maxAttributes, index, "Invalid attribute index")

	bytesToSkip := index / 8
	newIndex := index % 8

	attributeByteAddr := o.address.OffsetBytes(idx_Attributes).OffsetBytes(bytesToSkip)
	attributeByte := o.mem.ReadByte(attributeByteAddr)
	attributeByte = attributeByte | (0b1 << (7 - newIndex))
	o.mem.WriteByte(attributeByteAddr, attributeByte)
}

func (o object) Parent() ObjectId {
	if o.mem.GetVersion() <= 3 {
		return ObjectId(o.mem.ReadByte(o.address.OffsetBytes(idxV1_Parent)))
	} else {
		return ObjectId(o.mem.ReadWord(o.address.OffsetBytes(idxV4_Parent)))
	}
}

func (o object) Sibling() ObjectId {
	if o.mem.GetVersion() <= 3 {
		return ObjectId(o.mem.ReadByte(o.address.OffsetBytes(idxV1_Sibling)))
	} else {
		return ObjectId(o.mem.ReadWord(o.address.OffsetBytes(idxV4_Sibling)))
	}
}

func (o object) Child() ObjectId {
	if o.mem.GetVersion() <= 3 {
		return ObjectId(o.mem.ReadByte(o.address.OffsetBytes(idxV1_Child)))
	} else {
		return ObjectId(o.mem.ReadWord(o.address.OffsetBytes(idxV4_Child)))
	}
}

func (o *object) SetParent(parent ObjectId) {
	if o.mem.GetVersion() <= 3 {
		o.mem.WriteByte(o.address.OffsetBytes(idxV1_Parent), byte(parent))
	} else {
		o.mem.WriteWord(o.address.OffsetBytes(idxV4_Parent), word(parent))
	}
}

func (o *object) SetSibling(sibling ObjectId) {
	if o.mem.GetVersion() <= 3 {
		o.mem.WriteByte(o.address.OffsetBytes(idxV1_Sibling), byte(sibling))
	} else {
		o.mem.WriteWord(o.address.OffsetBytes(idxV4_Sibling), word(sibling))
	}
}

func (o *object) SetChild(child ObjectId) {
	if o.mem.GetVersion() <= 3 {
		o.mem.WriteByte(o.address.OffsetBytes(idxV1_Child), byte(child))
	} else {
		o.mem.WriteWord(o.address.OffsetBytes(idxV4_Child), word(child))
	}
}

func (o *object) ShortName() zstring.ZString {
	propertyTableOffset := idxV1_PropertiesAddr
	if o.mem.GetVersion() > 3 {
		propertyTableOffset = idxV4_PropertiesAddr
	}

	propertyTablePointer := o.address.OffsetBytes(propertyTableOffset)
	propertyTableAddr := memory.Address(o.mem.ReadWord(propertyTablePointer))
	return o.mem.GetZString(propertyTableAddr.OffsetBytes(1))
}

func (o object) Property(pid PropertyId) []byte {
	o.assertValidPropertyId(pid)
	data, found := o.findProperty(pid)

	if found {
		return data
	}

	return getPropertyDefault(o.mem, pid)
}

func (o *object) SetProperty(pid PropertyId, data []byte) {
	o.assertValidPropertyId(pid)
	existingData, found := o.findProperty(pid)
	assert.True(found, "Cannot set property that does not exist")

	for i := 0; i < min(len(data), len(existingData)); i++ {
		// TODO: Is this right if the arrays are different lengths?
		existingData[i] = data[i]
	}
}

func (o object) GetNextPropertyId(pid PropertyId) PropertyId {
	o.assertValidPropertyId(pid)
	propId, _, nextAddress := o.getFirstProperty()

	if pid == 0 {
		return propId
	}

	for propId != pid && propId != 0 {
		propId, _, nextAddress = parseProperty(o.mem, nextAddress)
	}
	assert.NotSame(propId, 0, "Can't get next property of non-existant property")

	propId, _, _ = parseProperty(o.mem, nextAddress)
	return propId
}

func (o object) GetPropertyDataAddress(pid PropertyId) memory.Address {
	o.assertValidPropertyId(pid)
	propId, data, nextAddress := o.getFirstProperty()

	for propId != pid && propId != 0 {
		propId, data, nextAddress = parseProperty(o.mem, nextAddress)
	}
	if propId == 0 {
		return 0
	}

	return nextAddress.OffsetBytes(-len(data))
}

func (o object) findProperty(pid PropertyId) ([]byte, bool) {
	propId, data, nextAddress := o.getFirstProperty()

	for propId != pid && propId != 0 {
		propId, data, nextAddress = parseProperty(o.mem, nextAddress)
	}

	if propId == pid {
		return data, true
	}

	return []byte{}, false
}

func (o object) getFirstProperty() (PropertyId, []byte, memory.Address) {
	propertyTableOffset := idxV1_PropertiesAddr
	if o.mem.GetVersion() > 3 {
		propertyTableOffset = idxV4_PropertiesAddr
	}

	propertyTablePointer := o.address.OffsetBytes(propertyTableOffset)
	propertyTableAddr := memory.Address(o.mem.ReadWord(propertyTablePointer))
	headerLength, headerDataAddr := o.mem.ReadByteNext(propertyTableAddr)
	propertyAddr := headerDataAddr.OffsetWords(int(headerLength))

	return parseProperty(o.mem, propertyAddr)
}

func parseProperty(mem *memory.Memory, address memory.Address) (PropertyId, []byte, memory.Address) {
	if mem.GetVersion() <= 3 {
		return parsePropertyV1(mem, address)
	}
	return parsePropertyV4(mem, address)
}

func parsePropertyV1(mem *memory.Memory, address memory.Address) (PropertyId, []byte, memory.Address) {
	sizeByte, nextAddress := mem.ReadByteNext(address)

	if sizeByte == 0 {
		return 0, []byte{}, nextAddress
	}

	propId := PropertyId(sizeByte & 0b11111) // bottom 5 bits
	length := int((sizeByte >> 5) + 1)
	data, nextAddress := mem.GetBytesNext(nextAddress, length)
	return propId, data, nextAddress
}

func parsePropertyV4(mem *memory.Memory, address memory.Address) (PropertyId, []byte, memory.Address) {
	sizeByte, nextAddress := mem.ReadByteNext(address)

	if sizeByte == 0 {
		return 0, []byte{}, nextAddress
	}

	propId := PropertyId(sizeByte & 0b111111) // bottom 6 bits
	var length int

	if (sizeByte >> 7) == 0 {
		length = int((sizeByte>>6)&0b1) + 1
	} else {
		sizeByte, nextAddress = mem.ReadByteNext(nextAddress)
		length = int(sizeByte & 0b111111) // bottom 6 bits
		if length == 0 {
			length = 64
		}
	}

	data, nextAddress := mem.GetBytesNext(nextAddress, length)
	return propId, data, nextAddress
}

func getPropertyDefault(mem *memory.Memory, pid PropertyId) []byte {
	defaultsTableAddr := mem.GetObjectsAddress()
	propDefaultAddr := defaultsTableAddr.OffsetWords(int(pid) - 1) // Property IDs start at 1
	return mem.GetBytes(propDefaultAddr, 2)
}

func (o object) assertValidPropertyId(pid PropertyId) {
	maxPropertyId := 31
	if o.mem.GetVersion() > 3 {
		maxPropertyId = 63
	}

	assert.LessThanEqual(maxPropertyId, int(pid), "PropertyId too large for version")
}
