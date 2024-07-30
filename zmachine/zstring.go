package zmachine

import (
	"errors"
	"strings"

	"github.com/Drakmyth/golang-zmachine/zmachine/internal/memory"
)

type ZStringKeyboard struct {
	ZMachine       ZMachine
	BaseCharset    int
	CurrentCharset int
	Version        uint8
}

var v1_alphabet = [3][32]byte{
	{' ', '\n', 0x00, 0x00, 0x00, 0x00, 'a', 'b', 'c', 'd', 'e', 'f', 'g', 'h', 'i', 'j', 'k', 'l', 'm', 'n', 'o', 'p', 'q', 'r', 's', 't', 'u', 'v', 'w', 'x', 'y', 'z'},
	{' ', '\n', 0x00, 0x00, 0x00, 0x00, 'A', 'B', 'C', 'D', 'E', 'F', 'G', 'H', 'I', 'J', 'K', 'L', 'M', 'N', 'O', 'P', 'Q', 'R', 'S', 'T', 'U', 'V', 'W', 'X', 'Y', 'Z'},
	{' ', '\n', 0x00, 0x00, 0x00, 0x00, 0x00, '0', '1', '2', '3', '4', '5', '6', '7', '8', '9', '.', ',', '!', '?', '_', '#', '\'', '"', '/', '\\', '<', '-', ':', '(', ')'},
}

var v2_alphabet = [3][32]byte{
	v1_alphabet[0],
	v1_alphabet[1],
	{' ', 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, '\n', '0', '1', '2', '3', '4', '5', '6', '7', '8', '9', '.', ',', '!', '?', '_', '#', '\'', '"', '/', '\\', '-', ':', '(', ')'},
}

func (keyboard *ZStringKeyboard) print(zchars []byte) (string, error) {
	builder := strings.Builder{}

	alphabet := v1_alphabet
	if keyboard.Version > 1 {
		alphabet = v2_alphabet
	}

	for i := 0; i < len(zchars); i++ {
		zchar := zchars[i]
		switch zchar {
		case 0x02:
			if keyboard.Version < 3 {
				keyboard.CurrentCharset = (keyboard.CurrentCharset + 1) % 3
				continue
			}
			fallthrough
		case 0x03:
			if keyboard.Version < 3 {
				keyboard.CurrentCharset = ((keyboard.CurrentCharset-1)%3 + 3) % 3
				continue
			}
			fallthrough
		case 0x01:
			if keyboard.Version >= 2 {
				i++
				builder.WriteString(keyboard.ZMachine.getAbbreviation(zchars[i], zchar))
				continue
			}
		case 0x04:
			if keyboard.Version < 3 {
				keyboard.CurrentCharset = (keyboard.CurrentCharset + 1) % 3
				keyboard.BaseCharset = (keyboard.BaseCharset + 1) % 3
			} else {
				keyboard.CurrentCharset = 1
			}
			continue
		case 0x05:
			if keyboard.Version < 3 {
				keyboard.CurrentCharset = ((keyboard.CurrentCharset-1)%3 + 3) % 3
				keyboard.BaseCharset = ((keyboard.BaseCharset-1)%3 + 3) % 3
			} else {
				keyboard.CurrentCharset = 2
			}
			continue
		case 0x06:
			if keyboard.CurrentCharset == 2 {
				return "", errors.New("unimplemented: keyboard multi-byte control character 0x06")
				// continue
			}
		}

		builder.WriteByte(alphabet[keyboard.CurrentCharset][zchar])
		keyboard.CurrentCharset = keyboard.BaseCharset
	}
	return builder.String(), nil
}

func (zmachine ZMachine) readZString(address memory.Address) (string, memory.Address) {
	words := make([]word, 0)
	zstr_word, next_address := zmachine.readWord(address)
	words = append(words, zstr_word)

	for zstr_word>>15 == 0 {
		zstr_word, next_address = zmachine.readWord(next_address)
		words = append(words, zstr_word)
	}

	zchars := make([]byte, 0, len(words)*3)
	for _, word := range words {
		zchars = append(zchars, byte((word>>10)&0b11111))
		zchars = append(zchars, byte((word>>5)&0b11111))
		zchars = append(zchars, byte(word&0b11111))
	}

	keyboard := ZStringKeyboard{ZMachine: zmachine, Version: zmachine.Header.Version}
	str, err := keyboard.print(zchars)
	if err != nil {
		panic(err)
	}
	return str, next_address
}

func (zmachine ZMachine) getAbbreviation(index byte, control byte) string {
	abbr_entry := zmachine.Header.AbbreviationsAddr.OffsetWords(int((32*(control-1) + index)))
	address, _ := zmachine.readWord(abbr_entry)
	abbreviation, _ := zmachine.readZString(memory.Address(address * 2))
	return abbreviation
}
