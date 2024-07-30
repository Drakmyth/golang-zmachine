package memory

type Stack[T any] []T

func (stack *Stack[T]) Push(value T) {
	*stack = append(*stack, value)
}

func (stack *Stack[T]) Pop() T {
	value := (*stack)[len(*stack)-1]
	*stack = (*stack)[:len(*stack)-1]
	return value
}

func (stack Stack[T]) Peek() *T {
	return &stack[len(stack)-1]
}
