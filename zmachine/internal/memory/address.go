package memory

type Address uint16

func (address Address) OffsetBytes(amount int) Address {
	return Address(int(address) + amount)
}

func (address Address) OffsetWords(amount int) Address {
	return Address(int(address) + 2*amount)
}
