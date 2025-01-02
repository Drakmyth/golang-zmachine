package zmachine

import "testing"

func AssertEqual[E comparable](t *testing.T, expected E, got E) {
	if expected != got {
		t.Fatalf("got: %v, expected %v", got, expected)
	}
}

func assertTrue(t *testing.T, got bool) {
	AssertEqual(t, true, got)
}

func assertFalse(t *testing.T, got bool) {
	AssertEqual(t, false, got)
}
