package zstring

import (
	"fmt"
	"slices"

	"github.com/Drakmyth/golang-zmachine/assert"
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
	PrintRune(zc ZChar) rune
	GetControlCharacter(zc ZChar) ctrlchar
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

func NewStaticCharset(alphabet []rune, ctrlchars []ctrlchar) Charset {
	assert.Length(alphabet, 78, "Invalid alphabet table length")
	assert.Length(ctrlchars, 6, "Invalid count of control characters")

	return staticCharset{
		charset: &charset{
			baseCharset:    0,
			currentCharset: 0,
			ctrlchars:      ctrlchars,
		},
		alphabet: alphabet,
	}
}

func NewDynamicCharset(alphabet func() []rune, ctrlchars []ctrlchar) Charset {
	assert.Length(ctrlchars, 6, "Invalid count of control characters")

	return dynamicCharset{
		charset: &charset{
			baseCharset:    0,
			currentCharset: 0,
			ctrlchars:      ctrlchars,
		},
		getAlphabet: alphabet,
	}
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

func (c staticCharset) PrintRune(zc ZChar) rune {
	return c.printRune(c.alphabet, zc)
}

func (c dynamicCharset) PrintRune(zc ZChar) rune {
	return c.printRune(c.getAlphabet(), zc)
}

func (c *charset) printRune(alphabet []rune, zc ZChar) rune {
	assert.Length(alphabet, 78, "Invalid alphabet table length")
	assert.GreaterThan(5, zc, "Control ZCharacter not translatable to rune")
	r := alphabet[int(zc-6)+(c.currentCharset*26)]
	c.currentCharset = c.baseCharset
	return r
}

func (c charset) GetControlCharacter(zc ZChar) ctrlchar {
	assert.Between(0, 6, zc, "Provided ZCharacter not in control range")
	return c.ctrlchars[zc]
}

func GetDefaultAlphabet(version int) []rune {
	alphabet := defaultAlphabet
	if version == 1 {
		alphabet = make([]rune, len(defaultAlphabet))
		copied := copy(alphabet, defaultAlphabet)
		assert.Same(copied, len(defaultAlphabet), fmt.Sprintf("Failed copying default alphabet: copied %d, expected %d", copied, len(defaultAlphabet)))
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
