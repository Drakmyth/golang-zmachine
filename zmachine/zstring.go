package zmachine

// func (zmachine ZMachine) read_zstring(address Address) {
// 	words := make([]uint16, 0)
// 	zstr_word, next_address := zmachine.read_word(address)
// 	words = append(words, zstr_word)

// 	for zstr_word>>15 == 0 {
// 		zstr_word, next_address = zmachine.read_word(next_address)
// 		words = append(words, zstr_word)
// 	}

// 	zchars := make([]uint8, 0, len(words)*3)
// 	for _, word := range words {
// 		zchars = append(zchars, uint8((word>>10)&0b11111))
// 		zchars = append(zchars, uint8((word>>5)&0b11111))
// 		zchars = append(zchars, uint8(word&0b11111))
// 	}

// 	print("temp")
// }
