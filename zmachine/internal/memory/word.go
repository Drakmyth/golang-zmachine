package memory

type Word uint16

func (word Word) HighByte() byte {
	return byte(word >> 8)
}

func (word Word) LowByte() byte {
	return byte(word)
}
