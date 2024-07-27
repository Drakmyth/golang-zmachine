package zmachine

type Stack[T any] []T

func (stack *Stack[T]) push(value T) {
	*stack = append(*stack, value)
}

func (stack *Stack[T]) pop() T {
	value := (*stack)[len(*stack)-1]
	*stack = (*stack)[:len(*stack)-1]
	return value
}

func (stack Stack[T]) peek() *T {
	return &stack[len(stack)-1]
}
