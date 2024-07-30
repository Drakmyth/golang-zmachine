package memory_test

import (
	"slices"
	"testing"

	"github.com/Drakmyth/golang-zmachine/zmachine"
	"github.com/Drakmyth/golang-zmachine/zmachine/internal/memory"
)

func TestStack_Push(t *testing.T) {
	stack := memory.Stack[int]{}

	data := []int{5, 7, 2}
	for _, val := range data {
		stack.Push(val)
		zmachine.AssertEqual(t, val, *stack.Peek())
	}
}

func TestStack_Pop(t *testing.T) {
	data := []int{5, 7, 2}
	stack := memory.Stack[int]{data[0], data[1], data[2]}

	slices.Reverse(data)
	for _, expected := range data {
		val := stack.Pop()
		zmachine.AssertEqual(t, expected, val)
	}
}

func TestStack_Peek(t *testing.T) {
	data := []int{5, 7, 2}
	stack := memory.Stack[int]{data[0], data[1], data[2]}

	zmachine.AssertEqual(t, data[2], *stack.Peek())
	zmachine.AssertEqual(t, 3, len(stack))
}
