package stack

import (
	"testing"
)

func TestStack(t *testing.T) {
	type spec[E any] struct {
		input []E
	}

	tests := map[string]spec[int]{
		"push peek pop size": {input: []int{1, 2, 3, 4, 5}},
	}

	for name, s := range tests {
		t.Run(name, func(t *testing.T) {
			stack := Stack[int]{}
			assertSame(t, 0, stack.Size(), "Stack not empty on init")

			for i, v := range s.input {
				stack.Push(v)
				assertSame(t, i+1, stack.Size(), "Incorrect stack size after push")
				assertSame(t, v, *stack.Peek(), "Incorrect top element peeked")
			}

			assertSame(t, len(s.input), stack.Size(), "Incorrect stack size before pop")
			for j := range s.input {
				i := len(s.input) - j - 1
				v := s.input[i]
				v2 := stack.Pop()
				assertSame(t, v, v2, "Incorrect element popped")
				assertSame(t, i, stack.Size(), "Incorrect stack size after pop")
			}
		})
	}
}

func assertSame[E comparable](t *testing.T, expected E, actual E, message string) {
	if expected != actual {
		t.Error(message)
	}
}
