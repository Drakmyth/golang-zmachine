package zmachine

type Stack[T any] []T

func (stack *Stack[T]) push(value T) {
	*stack = append([]T{value}, *stack...)
}

func (stack *Stack[T]) pop() T {
	value := (*stack)[0]
	*stack = (*stack)[1:]
	return value
}
