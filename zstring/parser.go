package zstring

import (
	"errors"
	"fmt"
	"strings"
)

type word = uint16

/*
 * Input Truncation
 *   V1: Truncate to 6 characters
 *   V4: Truncate to 9 characters
 *
 * Unicode translation table added in V5

 * Version 6:
 *   ZSCII code 9 ("tab") is defined for output
 *   ZSCII code 11 ("sentence space") is defined for output
 *   Complex unicode formatting is not supported
 *   Menu clicks are available
 * Version !6:
 *   Complex unicode formatting is optionally supported in window 0 (but not 1)
 */

type GetAbbreviationHandler func(bank int, index int) ZString

type parser struct {
	charset                 Charset
	pendingAbbreviationBank int
	UseAbbreviations        bool
	getAbbreviation         GetAbbreviationHandler
}

func NewParser(charset Charset, abbrevHandler GetAbbreviationHandler) parser {
	return parser{
		charset:                 charset,
		pendingAbbreviationBank: 0,
		UseAbbreviations:        true,
		getAbbreviation:         abbrevHandler,
	}
}

func (p parser) Parse(data ZString) (string, error) {
	zchars, err := parseZCharacters(data)
	if err != nil {
		return "", err
	}
	builder := strings.Builder{}

	for i := 0; i < len(zchars); i++ {
		zc := zchars[i]
		if p.UseAbbreviations && p.pendingAbbreviationBank > 0 {
			p.processAbbreviation(p.pendingAbbreviationBank, int(zc), &builder)
			continue
		}

		if zc < 6 {
			p.processControlCharacter(zc, &builder)
			continue
		}

		switch zc {
		case 6:
			// if p.charset.currentCharset == 2 {
			// TODO: Multi-byte character
			// 	continue
			// }
			fallthrough
		default:
			builder.WriteRune(p.charset.PrintRune(zc))
		}
	}

	return builder.String(), nil
}

func (p *parser) processControlCharacter(zc ZChar, builder *strings.Builder) error {
	ctrl := p.charset.GetControlCharacter(zc)

	switch ctrl {
	case CTRL_Abbreviation:
		if !p.UseAbbreviations {
			return nil
		}

		p.pendingAbbreviationBank = int(zc)
	case CTRL_Backshift:
		p.charset.Backshift()
	case CTRL_BackshiftLock:
		p.charset.Backshift()
		p.charset.Lock()
	case CTRL_NewLine:
		_, err := builder.WriteRune('\n')
		return err
	case CTRL_Shift:
		p.charset.Shift()
	case CTRL_ShiftLock:
		p.charset.Shift()
		p.charset.Lock()
	case CTRL_Space:
		_, err := builder.WriteRune(' ')
		return err
	default:
		panic(fmt.Sprintf("unexpected zstring.CtrlChar: %#v", ctrl))
	}

	return nil
}

func parseZCharacters(data ZString) ([]ZChar, error) {
	if len(data)%2 != 0 {
		return []ZChar{}, errors.New("ZString must contain even number of bytes")
	}

	const MASK_ZChar = 0b11111

	zchars := make([]ZChar, 0, len(data)*3/2)

	lastWord := false
	for i := 0; i < len(data); i += 2 {
		zword := word(data[i])<<8 | word(data[i+1])
		zchar1 := ZChar((zword >> 10) & MASK_ZChar)
		zchar2 := ZChar((zword >> 5) & MASK_ZChar)
		zchar3 := ZChar(zword & MASK_ZChar)
		lastWord = zword>>15 != 0

		zchars = append(zchars, zchar1, zchar2, zchar3)
		if lastWord {
			break
		}
	}

	var err error = nil
	if !lastWord {
		err = errors.New("Malformed ZString encountered, missing end flag")
	}

	return zchars, err
}

func (p *parser) processAbbreviation(bank int, index int, builder *strings.Builder) error {
	abbreviation := p.getAbbreviation(bank, index)
	oldValue := p.UseAbbreviations
	p.UseAbbreviations = false
	str, err := p.Parse(abbreviation)
	if err != nil {
		return err
	}
	builder.WriteString(str)
	p.UseAbbreviations = oldValue
	p.pendingAbbreviationBank = 0
	return nil
}
