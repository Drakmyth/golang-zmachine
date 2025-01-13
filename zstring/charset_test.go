package zstring

import (
	"testing"
	"unicode/utf8"

	"github.com/Drakmyth/golang-zmachine/testassert"
)

func TestCharacterSetShifting(t *testing.T) {
	version := 2
	defaultAlphabet := GetDefaultAlphabet(version)
	ctrlchars := GetDefaultCtrlCharMapping(version)

	type spec struct {
		expected string
		control  func(Charset)
	}

	tests := map[string]spec{
		"Shift": {
			expected: "cCc",
			control: func(charset Charset) {
				charset.Shift()
			},
		},
		"ShiftLock": {
			expected: "cCC",
			control: func(charset Charset) {
				charset.Shift()
				charset.Lock()
			},
		},
		"Backshift": {
			expected: "c0c",
			control: func(charset Charset) {
				charset.Backshift()
			},
		},
		"BackshiftLock": {
			expected: "c00",
			control: func(charset Charset) {
				charset.Backshift()
				charset.Lock()
			},
		},
		"Double Shift": {
			expected: "cCc",
			control: func(charset Charset) {
				charset.Shift()
				charset.Shift()
			},
		},
		"Double Backshift": {
			expected: "c0c",
			control: func(charset Charset) {
				charset.Backshift()
				charset.Backshift()
			},
		},
	}

	zc := ZChar(8)
	for name, s := range tests {
		t.Run(name, func(t *testing.T) {
			charset := NewStaticCharset(defaultAlphabet, ctrlchars)
			actual := utf8.AppendRune(make([]byte, 0, 3), charset.PrintRune(zc))
			s.control(charset)
			actual = utf8.AppendRune(actual, charset.PrintRune(zc))
			actual = utf8.AppendRune(actual, charset.PrintRune(zc))

			testassert.Same(t, s.expected, string(actual))
		})
	}
}

func TestCharacterSet(t *testing.T) {
	version := 1
	defaultAlphabet := GetDefaultAlphabet(version)
	ctrlchars := GetDefaultCtrlCharMapping(version)

	type spec struct {
		chars []rune
		init  func(Charset)
	}

	tests := map[string]spec{
		"a0": {
			chars: defaultAlphabet[0:26],
			init:  func(charset Charset) {},
		},
		"a1": {
			chars: defaultAlphabet[26:52],
			init: func(charset Charset) {
				charset.Shift()
				charset.Lock()
			},
		},
		"a2": {
			chars: defaultAlphabet[52:],
			init: func(charset Charset) {
				charset.Backshift()
				charset.Lock()
			},
		},
	}

	for name, s := range tests {
		t.Run(name, func(t *testing.T) {
			charset := NewStaticCharset(defaultAlphabet, ctrlchars)

			s.init(charset)

			// Z-Characters range from 0 to 1F
			// 0-5 are control characters, 6-1F get converted to ZSCII by the alphabet table
			for zc := ZChar(6); zc <= 0x1F; zc++ {
				actual := charset.PrintRune(zc)
				expected := s.chars[zc-6]
				testassert.Same(t, expected, actual)
			}
		})
	}
}

func TestDefaultControlCharacterMapping(t *testing.T) {
	type TestCharset struct {
		ZChar0 ctrlchar
		ZChar1 ctrlchar
		ZChar2 ctrlchar
		ZChar3 ctrlchar
		ZChar4 ctrlchar
		ZChar5 ctrlchar
	}

	type spec struct {
		version int
		chars   TestCharset
	}

	v3chars := TestCharset{
		ZChar0: CTRL_Space,
		ZChar1: CTRL_Abbreviation,
		ZChar2: CTRL_Abbreviation,
		ZChar3: CTRL_Abbreviation,
		ZChar4: CTRL_Shift,
		ZChar5: CTRL_Backshift,
	}

	tests := map[string]spec{
		"v1": {
			version: 1,
			chars: TestCharset{
				ZChar0: CTRL_Space,
				ZChar1: CTRL_NewLine,
				ZChar2: CTRL_Shift,
				ZChar3: CTRL_Backshift,
				ZChar4: CTRL_ShiftLock,
				ZChar5: CTRL_BackshiftLock,
			},
		},
		"v2": {
			version: 2,
			chars: TestCharset{
				ZChar0: CTRL_Space,
				ZChar1: CTRL_Abbreviation,
				ZChar2: CTRL_Shift,
				ZChar3: CTRL_Backshift,
				ZChar4: CTRL_ShiftLock,
				ZChar5: CTRL_BackshiftLock,
			},
		},
		"v3": {
			version: 3,
			chars:   v3chars,
		},
		"v4": {
			version: 4,
			chars:   v3chars,
		},
		"v5": {
			version: 5,
			chars:   v3chars,
		},
		"v6": {
			version: 6,
			chars:   v3chars,
		},
		"v7": {
			version: 7,
			chars:   v3chars,
		},
		"v8": {
			version: 8,
			chars:   v3chars,
		},
	}

	for name, s := range tests {
		t.Run(name, func(t *testing.T) {
			actual := GetDefaultCtrlCharMapping(s.version)
			testassert.Same(t, s.chars.ZChar0, actual[0])
			testassert.Same(t, s.chars.ZChar1, actual[1])
			testassert.Same(t, s.chars.ZChar2, actual[2])
			testassert.Same(t, s.chars.ZChar3, actual[3])
			testassert.Same(t, s.chars.ZChar4, actual[4])
			testassert.Same(t, s.chars.ZChar5, actual[5])
		})
	}
}

func TestDefaultAlphabet(t *testing.T) {
	v2alphabet := "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ\x00\n0123456789.,!?_#'\"/\\-:()"

	type spec struct {
		version  int
		alphabet string
	}

	tests := map[string]spec{
		"v1": {
			version:  1,
			alphabet: "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ\x000123456789.,!?_#'\"/\\<-:()",
		},
		"v2": {
			version:  2,
			alphabet: v2alphabet,
		},
		"v3": {
			version:  3,
			alphabet: v2alphabet,
		},
		"v4": {
			version:  4,
			alphabet: v2alphabet,
		},
		"v5": {
			version:  5,
			alphabet: v2alphabet,
		},
		"v6": {
			version:  6,
			alphabet: v2alphabet,
		},
		"v7": {
			version:  7,
			alphabet: v2alphabet,
		},
		"v8": {
			version:  8,
			alphabet: v2alphabet,
		},
	}

	for name, s := range tests {
		t.Run(name, func(t *testing.T) {
			actual := GetDefaultAlphabet(s.version)
			testassert.Same(t, len(s.alphabet), len(actual))

			for i, r := range s.alphabet {
				testassert.Same(t, r, actual[i])
			}
		})
	}
}
