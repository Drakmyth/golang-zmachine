package assert

import (
	"errors"
	"testing"
)

func TestBetween(t *testing.T) {
	type spec struct {
		min         int
		max         int
		value       int
		shouldPanic bool
	}

	tests := map[string]spec{
		"below min":     {min: 10, max: 12, value: 9, shouldPanic: true},
		"inclusive min": {min: 10, max: 12, value: 10, shouldPanic: false},
		"exclusive max": {min: 10, max: 12, value: 12, shouldPanic: true},
		"above max":     {min: 10, max: 12, value: 13, shouldPanic: true},
	}

	for name, s := range tests {
		t.Run(name, func(t *testing.T) {
			defer catchPanic(t, s.shouldPanic)

			Between(s.min, s.max, s.value, "Error message goes here")

			assertPanic(t, s.shouldPanic)
		})
	}
}

func TestNoError(t *testing.T) {
	type spec struct {
		err         error
		shouldPanic bool
	}

	tests := map[string]spec{
		"error":    {err: errors.New("Error"), shouldPanic: true},
		"no error": {err: nil, shouldPanic: false},
	}

	for name, s := range tests {
		t.Run(name, func(t *testing.T) {
			defer catchPanic(t, s.shouldPanic)

			NoError(s.err, "Error message goes here")

			assertPanic(t, s.shouldPanic)
		})
	}
}

func TestNotContains(t *testing.T) {
	type spec[E any] struct {
		data        []E
		value       E
		shouldPanic bool
	}

	tests := map[string]spec[int]{
		"contains":     {data: []int{1, 2, 3, 4, 5}, value: 3, shouldPanic: true},
		"not contains": {data: []int{1, 2, 3, 4, 5}, value: 10, shouldPanic: false},
	}

	for name, s := range tests {
		t.Run(name, func(t *testing.T) {
			defer catchPanic(t, s.shouldPanic)

			NotContains(s.data, s.value, "Error message goes here")

			assertPanic(t, s.shouldPanic)
		})
	}
}

func TestNotEmpty(t *testing.T) {
	type spec[E any] struct {
		data        []E
		shouldPanic bool
	}

	tests := map[string]spec[int]{
		"empty":     {data: []int{}, shouldPanic: true},
		"not empty": {data: []int{1, 2, 3, 4, 5}, shouldPanic: false},
	}

	for name, s := range tests {
		t.Run(name, func(t *testing.T) {
			defer catchPanic(t, s.shouldPanic)

			NotEmpty(s.data, "Error message goes here")

			assertPanic(t, s.shouldPanic)
		})
	}
}

func catchPanic(t *testing.T, shouldPanic bool) {
	err := recover()
	if err != nil && !shouldPanic {
		t.Errorf("Unexpected panick")
	}
}

func assertPanic(t *testing.T, shouldPanic bool) {
	if shouldPanic {
		t.Errorf("Expected panick but didn't")
	}
}
