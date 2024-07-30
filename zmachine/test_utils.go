package zmachine

import "testing"

func AssertEqual(t *testing.T, expected any, got any) {
	if expected != got {
		t.Fatalf("got: %x, expected %x", got, expected)
	}
}

func assertTrue(t *testing.T, got any) {
	AssertEqual(t, true, got)
}

func assertFalse(t *testing.T, got any) {
	AssertEqual(t, false, got)
}
