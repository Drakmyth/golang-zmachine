package assert

import (
	"cmp"
	"slices"
)

func Between[E cmp.Ordered](inclusiveMin E, exclusiveMax E, v E, message string) {
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
