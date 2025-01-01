package stack

import "github.com/Drakmyth/golang-zmachine/assert"

type Stack[E any] []E

func (s *Stack[E]) Push(v E) {
	*s = append(*s, v)
}

func (s Stack[E]) Peek() *E {
	assert.NotEmpty(s, "Cannot peek from empty stack")
	return &s[len(s)-1]
}

func (s *Stack[E]) Pop() E {
	assert.NotEmpty(*s, "Cannot pop from empty stack")
	v := (*s)[len(*s)-1]
	*s = (*s)[:len(*s)-1]
	return v
}

func (s Stack[E]) Size() int {
	return len(s)
}
