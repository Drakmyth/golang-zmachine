package zstring

type ZString []byte

func (z ZString) LenChars() int {
	return len(z) * 3
}

func (z ZString) LenBytes() int {
	return len(z)
}
