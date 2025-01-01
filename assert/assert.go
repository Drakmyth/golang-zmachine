package assert

import (
	"slices"
)

type numeric interface {
	~uint | ~uint8 | ~uint16 | ~uint32 | ~uint64 |
		~int | ~int8 | ~int16 | ~int32 | ~int64 |
		~float32 | ~float64 | ~uintptr
}

func Between[E numeric](inclusiveMin E, exclusiveMax E, v E, message string) {
	if v < inclusiveMin || v >= exclusiveMax {
		panic(message)
	}
}

func NoError(v error, message string) {
	if v != nil {
		panic(message)
	}
}

func NotContains[S ~[]E, E comparable](s S, v E, message string) {
	if slices.Contains(s, v) {
		panic(message)
	}
}

func NotEmpty[S ~[]E, E any](s S, message string) {
	if len(s) == 0 {
		panic(message)
	}
}
