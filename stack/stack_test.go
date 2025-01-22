package stack

import (
	"testing"

	"github.com/Drakmyth/golang-zmachine/testassert"
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

			testassert.Same(t, 0, stack.Size()) // "Stack not empty on init"

			for i, v := range s.input {
				stack.Push(v)
				testassert.Same(t, i+1, stack.Size()) // "Incorrect stack size after push"
				actual, err := stack.Peek()
				testassert.NoError(t, err)
				testassert.Same(t, v, *actual) // "Incorrect top element peeked"
			}

			testassert.Same(t, len(s.input), stack.Size()) // "Incorrect stack size before pop"
			for j := range s.input {
				i := len(s.input) - j - 1
				v := s.input[i]
				v2, err := stack.Pop()
				testassert.NoError(t, err)
				testassert.Same(t, v, v2)           // "Incorrect element popped"
				testassert.Same(t, i, stack.Size()) // "Incorrect stack size after pop"
			}
		})
	}
}
