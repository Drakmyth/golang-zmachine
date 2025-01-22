package stack

import (
	"errors"
)

type Stack[E any] []E

func (s *Stack[E]) Push(v E) {
	*s = append(*s, v)
}

func (s Stack[E]) Peek() (*E, error) {
	if len(s) == 0 {
		var defaultE *E
		return defaultE, errors.New("Cannot peek from empty stack")
	}

	return &s[len(s)-1], nil
}

func (s *Stack[E]) Pop() (E, error) {
	if len(*s) == 0 {
		var defaultE E
		return defaultE, errors.New("Cannot pop from empty stack")
	}

	v := (*s)[len(*s)-1]
	*s = (*s)[:len(*s)-1]
	return v, nil
}

func (s Stack[E]) Size() int {
	return len(s)
}
