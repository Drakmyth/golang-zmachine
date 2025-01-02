package assert

import (
	"errors"
	"testing"

	"github.com/Drakmyth/golang-zmachine/testassert"
)

func TestBetween(t *testing.T) {
	type spec struct {
		min    int
		max    int
		value  int
		assert testassert.PanicAssertion
	}

	tests := map[string]spec{
		"below min":     {min: 10, max: 12, value: 9, assert: testassert.Panics},
		"inclusive min": {min: 10, max: 12, value: 10, assert: testassert.NoPanic},
		"exclusive max": {min: 10, max: 12, value: 12, assert: testassert.Panics},
		"above max":     {min: 10, max: 12, value: 13, assert: testassert.Panics},
	}

	for name, s := range tests {
		t.Run(name, func(t *testing.T) {
			s.assert(t, func() {
				Between(s.min, s.max, s.value, "Error message goes here")
			})
		})
	}
}

func TestNoError(t *testing.T) {
	type spec struct {
		err    error
		assert testassert.PanicAssertion
	}

	tests := map[string]spec{
		"error":    {err: errors.New("Error"), assert: testassert.Panics},
		"no error": {err: nil, assert: testassert.NoPanic},
	}

	for name, s := range tests {
		t.Run(name, func(t *testing.T) {
			s.assert(t, func() {
				NoError(s.err, "Error message goes here")
			})
		})
	}
}

func TestNotContains(t *testing.T) {
	type spec[E any] struct {
		data   []E
		value  E
		assert testassert.PanicAssertion
	}

	tests := map[string]spec[int]{
		"contains":     {data: []int{1, 2, 3, 4, 5}, value: 3, assert: testassert.Panics},
		"not contains": {data: []int{1, 2, 3, 4, 5}, value: 10, assert: testassert.NoPanic},
	}

	for name, s := range tests {
		t.Run(name, func(t *testing.T) {
			s.assert(t, func() {
				NotContains(s.data, s.value, "Error message goes here")
			})
		})
	}
}

func TestNotEmpty(t *testing.T) {
	type spec[E any] struct {
		data   []E
		assert testassert.PanicAssertion
	}

	tests := map[string]spec[int]{
		"empty":     {data: []int{}, assert: testassert.Panics},
		"not empty": {data: []int{1, 2, 3, 4, 5}, assert: testassert.NoPanic},
	}

	for name, s := range tests {
		t.Run(name, func(t *testing.T) {
			s.assert(t, func() {
				NotEmpty(s.data, "Error message goes here")
			})
		})
	}
}
