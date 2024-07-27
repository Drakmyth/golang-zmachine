package zmachine

import "testing"

func assertEqual(t *testing.T, expected any, got any) {
	if expected != got {
		t.Fatalf("got: %x, expected %x", got, expected)
	}
}

func assertTrue(t *testing.T, got any) {
	assertEqual(t, true, got)
}

func assertFalse(t *testing.T, got any) {
	assertEqual(t, false, got)
}
