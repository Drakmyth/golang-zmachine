package memory

type Address uint16
type Word uint16

func (word Word) HighByte() byte {
	return byte(word >> 8)
}

func (word Word) LowByte() byte {
	return byte(word)
}

func (address Address) OffsetBytes(amount int) Address {
	return Address(int(address) + amount)
}

func (address Address) OffsetWords(amount int) Address {
	return Address(int(address) + 2*amount)
}
