package zmachine

import (
	"slices"
	"testing"
)

func TestStack_push(t *testing.T) {
	stack := Stack[int]{}

	data := []int{5, 7, 2}
	for _, val := range data {
		stack.push(val)
		AssertEqual(t, val, *stack.peek())
	}
}

func TestStack_pop(t *testing.T) {
	data := []int{5, 7, 2}
	stack := Stack[int]{data[0], data[1], data[2]}

	slices.Reverse(data)
	for _, expected := range data {
		val := stack.pop()
		AssertEqual(t, expected, val)
	}
}

func TestStack_peek(t *testing.T) {
	data := []int{5, 7, 2}
	stack := Stack[int]{data[0], data[1], data[2]}

	AssertEqual(t, data[2], *stack.peek())
	AssertEqual(t, 3, len(stack))
}
