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

func GreaterThan[E cmp.Ordered](exclusiveMin E, v E, message string) {
	if v <= exclusiveMin {
		panic(message)
	}
}

func LessThan[E cmp.Ordered](exclusiveMax E, v E, message string) {
	if v >= exclusiveMax {
		panic(message)
	}
}

func LessThanEqual[E cmp.Ordered](exclusiveMax E, v E, message string) {
	if v > exclusiveMax {
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

func NotSame[E comparable](v1 E, v2 E, message string) {
	if v1 == v2 {
		panic(message)
	}
}

func Same[E comparable](v1 E, v2 E, message string) {
	if v1 != v2 {
		panic(message)
	}
}

func True(v bool, message string) {
	if !v {
		panic(message)
	}
}

func Length[S ~[]E, E any](s S, length int, message string) {
	if len(s) != length {
		panic(message)
	}
}
