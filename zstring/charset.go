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

type CtrlChar uint8

const (
	CTRL_Space CtrlChar = iota
	CTRL_NewLine
	CTRL_Abbreviation
	CTRL_Shift
	CTRL_Backshift
	CTRL_ShiftLock
	CTRL_BackshiftLock
)

type Charset struct {
	alphabet       []rune
	ctrlchars      []CtrlChar
	baseCharset    uint8
	currentCharset uint8
}

func NewCharset(alphabet []rune, ctrlchars []CtrlChar) Charset {
	assert.Length(alphabet, 78, "Invalid alphabet table length")
	assert.Length(ctrlchars, 6, "Invalid count of control characters")

	return Charset{
		alphabet,
		ctrlchars,
		0,
		0,
	}
}

func (c *Charset) Shift() {
	c.currentCharset = (c.currentCharset + 1) % 3
}

func (c *Charset) Backshift() {
	c.currentCharset = (c.currentCharset + 2) % 3 // adding length - 1 == subtracting 1
}

func (c *Charset) Lock() {
	c.baseCharset = c.currentCharset
}

func (c *Charset) Reset() {
	c.currentCharset = c.baseCharset
}

func (c Charset) GetRune(zc ZChar) rune {
	assert.GreaterThan(5, zc, "Control ZCharacter not translatable to rune")
	return c.alphabet[uint8(zc-6)+(c.currentCharset*26)]
}

func (c Charset) GetControlCharacter(zc ZChar) CtrlChar {
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

func GetDefaultCtrlCharMapping(version int) []CtrlChar {
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
		mapping := []CtrlChar{CTRL_Space, CTRL_NewLine, CTRL_Shift, CTRL_Backshift, CTRL_ShiftLock, CTRL_BackshiftLock}
		if version == 2 {
			mapping[1] = CTRL_Abbreviation
		}
		return mapping
	default:
		return []CtrlChar{CTRL_Space, CTRL_Abbreviation, CTRL_Abbreviation, CTRL_Abbreviation, CTRL_Shift, CTRL_Backshift}
	}
}
