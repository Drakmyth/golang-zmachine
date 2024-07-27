package zmachine

type Address uint16
type word uint16

func (word word) highByte() byte {
	return byte(word >> 8)
}

func (word word) lowByte() byte {
	return byte(word)
}

func (address Address) offsetBytes(amount int) Address {
	return Address(int(address) + amount)
}

func (address Address) offsetWords(amount int) Address {
	return Address(int(address) + 2*amount)
}

func (zmachine ZMachine) readByte(address Address) (byte, Address) {
	return zmachine.Memory[address], address.offsetBytes(1)
}

func (zmachine *ZMachine) writeByte(value byte, address Address) {
	zmachine.Memory[address] = value
}

func (zmachine ZMachine) readWord(address Address) (word, Address) {
	high := word(zmachine.Memory[address])
	low := word(zmachine.Memory[address.offsetBytes(1)])
	return (high << 8) | low, address.offsetWords(1)
}

func (zmachine *ZMachine) writeWord(value word, address Address) {
	high := byte(value >> 8)
	low := byte(value)
	zmachine.Memory[address] = high
	zmachine.Memory[address.offsetBytes(1)] = low
}
