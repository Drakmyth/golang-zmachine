package memory

type Address word

func (address Address) OffsetBytes(amount int) Address {
	return Address(int(address) + amount)
}

func (address Address) OffsetWords(amount int) Address {
	return Address(int(address) + 2*amount)
}
