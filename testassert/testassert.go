package testassert

import "testing"

func Same[E comparable](t *testing.T, expected E, actual E) {
	t.Helper()
	if expected != actual {
		t.Errorf("Expected %v, Received %v", expected, actual)
	}
}

func True(t *testing.T, actual bool) {
	t.Helper()
	Same(t, true, actual)
}

func False(t *testing.T, actual bool) {
	t.Helper()
	Same(t, false, actual)
}

func Panics(t *testing.T, f func()) {
	t.Helper()
	defer catchPanic(t, true)
	f()
	assertPanic(t, true)
}

type PanicAssertion func(t *testing.T, f func())

func NoPanic(t *testing.T, f func()) {
	t.Helper()
	defer catchPanic(t, false)
	f()
	assertPanic(t, false)
}

func catchPanic(t *testing.T, shouldPanic bool) {
	err := recover()
	if err != nil && !shouldPanic {
		t.Error("Function paniced unexpectedly")
	}
}

func assertPanic(t *testing.T, shouldPanic bool) {
	if shouldPanic {
		t.Error("Expected function to panic, but it didn't")
	}
}
