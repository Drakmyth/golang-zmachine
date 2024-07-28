package screen

import "fmt"

func Clear() {
	clearScreen()
	eraseScrollback()
	resetCursor()
}

func clearScreen() {
	fmt.Print("\033[2J")
}

func eraseScrollback() {
	fmt.Print("\033[3J")
}

func resetCursor() {
	// fmt.Print("\033[H")
	fmt.Print("\033[999;1H")
}
