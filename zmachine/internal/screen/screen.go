package screen

import "fmt"

// See https://en.wikipedia.org/wiki/ANSI_escape_code for escape code definitions

const ANSI_EraseInDisplay_EntireScreen = "\033[2J"
const ANSI_EraseInDisplay_EntireScreenAndScrollback = "\033[3J" // At least in PowerShell this seems to only clear scrollback and not the screen
const ANSI_CursorPosition_BottomLeft = "\033[999;1H"

func Clear() {
	clearScreen()
	eraseScrollback()
	resetCursor()
}

func clearScreen() {
	fmt.Print(ANSI_EraseInDisplay_EntireScreen)
}

func eraseScrollback() {
	fmt.Print(ANSI_EraseInDisplay_EntireScreenAndScrollback)
}

func resetCursor() {
	fmt.Print(ANSI_CursorPosition_BottomLeft)
}
