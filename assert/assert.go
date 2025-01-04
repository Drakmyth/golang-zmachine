package assert

import (
	"cmp"
	"fmt"
	"slices"
)

func Between[E cmp.Ordered](inclusiveMin E, exclusiveMax E, v E, message string, msgargs ...any) {
	if v < inclusiveMin || v >= exclusiveMax {

		panic(fmt.Sprintf(message, msgargs...))
	}
}

func NoError(v error, message string, msgargs ...any) {
	if v != nil {
		panic(fmt.Sprintf(message, msgargs...))
	}
}

func NotContains[S ~[]E, E comparable](s S, v E, message string, msgargs ...any) {
	if slices.Contains(s, v) {
		panic(fmt.Sprintf(message, msgargs...))
	}
}

func NotEmpty[S ~[]E, E any](s S, message string, msgargs ...any) {
	if len(s) == 0 {
		panic(fmt.Sprintf(message, msgargs...))
	}
}

func True(v bool, message string, msgargs ...any) {
	if !v {
		panic(fmt.Sprintf(message, msgargs...))
	}
}

func NotSame[E comparable](v1 E, v2 E, message string, msgargs ...any) {
	if v2 == v1 {
		panic(fmt.Sprintf(message, msgargs...))
	}
}
