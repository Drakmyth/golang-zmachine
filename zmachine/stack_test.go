package zmachine

import (
	"testing"
)

func TestStack_push(t *testing.T) {
	stack := Stack[int]{}

	data := []int{5, 7, 2}
	for _, val := range data {
		stack.push(val)
	}

	for i, val := range stack {
		assertEqual(t, data[len(data)-1-i], val)
	}
}

func TestStack_pop(t *testing.T) {
	data := []int{5, 7, 2}
	stack := Stack[int]{data[0], data[1], data[2]}

	val := stack.pop()

	assertEqual(t, data[0], val)

	for i, val2 := range stack {
		if data[i+1] != val2 {
			t.Fatalf("got: %v, expected %v", stack, data[1:])
		}
	}
}
