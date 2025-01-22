package zstring

type ZString []byte

func (z ZString) LenZChars() int {
	return len(z) * 3
}

func (z ZString) LenBytes() int {
	return len(z)
}
