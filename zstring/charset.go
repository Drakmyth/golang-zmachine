package zstring

import (
	"errors"
	"slices"
)

type ZChar byte

var defaultAlphabet = []rune{
	'a', 'b', 'c', 'd', 'e', 'f', 'g', 'h', 'i', 'j', 'k', 'l', 'm', 'n', 'o', 'p', 'q', 'r', 's', 't', 'u', 'v', 'w', 'x', 'y', 'z',
	'A', 'B', 'C', 'D', 'E', 'F', 'G', 'H', 'I', 'J', 'K', 'L', 'M', 'N', 'O', 'P', 'Q', 'R', 'S', 'T', 'U', 'V', 'W', 'X', 'Y', 'Z',
	0x00, '\n', '0', '1', '2', '3', '4', '5', '6', '7', '8', '9', '.', ',', '!', '?', '_', '#', '\'', '"', '/', '\\', '-', ':', '(', ')',
}

var v1ThirdRow = []rune{
	0x00, '0', '1', '2', '3', '4', '5', '6', '7', '8', '9', '.', ',', '!', '?', '_', '#', '\'', '"', '/', '\\', '<', '-', ':', '(', ')',
}

type ctrlchar int

const (
	CTRL_Space ctrlchar = iota
	CTRL_NewLine
	CTRL_Abbreviation
	CTRL_Shift
	CTRL_Backshift
	CTRL_ShiftLock
	CTRL_BackshiftLock
)

type Charset interface {
	Shift()
	Backshift()
	Lock()
	PrintRune(zc ZChar) (rune, error)
	GetControlCharacter(zc ZChar) (ctrlchar, error)
}

type charset struct {
	baseCharset    int
	currentCharset int
	ctrlchars      []ctrlchar
}

type staticCharset struct {
	*charset

	alphabet []rune
}

type dynamicCharset struct {
	*charset

	getAlphabet func() []rune
}

func NewStaticCharset(alphabet []rune, ctrlchars []ctrlchar) (Charset, error) {
	if len(alphabet) != 78 {
		return nil, errors.New("Invalid alphabet table length")
	}
	if len(ctrlchars) != 6 {
		return nil, errors.New("Invalid count of control characters")
	}

	return staticCharset{
		charset: &charset{
			baseCharset:    0,
			currentCharset: 0,
			ctrlchars:      ctrlchars,
		},
		alphabet: alphabet,
	}, nil
}

func NewDynamicCharset(alphabet func() []rune, ctrlchars []ctrlchar) (Charset, error) {
	if len(ctrlchars) != 6 {
		return nil, errors.New("Invalid count of control characters")
	}

	return dynamicCharset{
		charset: &charset{
			baseCharset:    0,
			currentCharset: 0,
			ctrlchars:      ctrlchars,
		},
		getAlphabet: alphabet,
	}, nil
}

func (c *charset) Shift() {
	c.currentCharset = (c.baseCharset + 1) % 3
}

func (c *charset) Backshift() {
	c.currentCharset = (c.baseCharset + 2) % 3 // adding length - 1 == subtracting 1
}

func (c *charset) Lock() {
	c.baseCharset = c.currentCharset
}

func (c staticCharset) PrintRune(zc ZChar) (rune, error) {
	return c.printRune(c.alphabet, zc)
}

func (c dynamicCharset) PrintRune(zc ZChar) (rune, error) {
	alphabet := c.getAlphabet()
	if len(alphabet) != 78 {
		return '\x00', errors.New("Invalid alphabet table length")
	}
	return c.printRune(alphabet, zc)
}

func (c *charset) printRune(alphabet []rune, zc ZChar) (rune, error) {
	if zc < 6 {
		return '\x00', errors.New("Control ZCharacter not translatable to rune")
	}

	r := alphabet[int(zc-6)+(c.currentCharset*26)]
	c.currentCharset = c.baseCharset
	return r, nil
}

func (c charset) GetControlCharacter(zc ZChar) (ctrlchar, error) {
	if zc > 5 {
		return CTRL_Space, errors.New("Provided ZCharacter not in control range")
	}

	return c.ctrlchars[zc], nil
}

func GetDefaultAlphabet(version int) []rune {
	alphabet := defaultAlphabet
	if version == 1 {
		alphabet = make([]rune, len(defaultAlphabet))
		copy(alphabet, defaultAlphabet)
		alphabet = slices.Replace(alphabet, 52, len(alphabet), v1ThirdRow...)
	}

	return alphabet
}

func GetDefaultCtrlCharMapping(version int) []ctrlchar {
	/*
	 * Control Characters
	 *   Char | V1        | V2        | V3+
	 *   -----|-----------|-----------|----------
	 *   0    | Space     | Space     | Space
	 *   1    | New-Line  | Abbrev    | Abbrev
	 *   2    | Shift     | Shift     | Abbrev
	 *   3    | Backshift | Backshift | Abbrev
	 *   4    | Lock      | Lock      | Shift
	 *   5    | Backlock  | Backlock  | Backshift
	 */

	switch version {
	case 1, 2:
		mapping := []ctrlchar{CTRL_Space, CTRL_NewLine, CTRL_Shift, CTRL_Backshift, CTRL_ShiftLock, CTRL_BackshiftLock}
		if version == 2 {
			mapping[1] = CTRL_Abbreviation
		}
		return mapping
	default:
		return []ctrlchar{CTRL_Space, CTRL_Abbreviation, CTRL_Abbreviation, CTRL_Abbreviation, CTRL_Shift, CTRL_Backshift}
	}
}
