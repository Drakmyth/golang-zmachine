package zmachine

type Header struct {
	Version               uint8
	Flags1                uint8
	ReleaseNumber         uint16
	HighMemoryAddr        Address
	InitialProgramCounter Address
	DictionaryAddr        Address
	ObjectsAddr           Address
	GlobalsAddr           Address
	StaticMemoryAddr      Address
	Flags2                uint16
	Serial                [6]byte
	AbbreviationsAddr     Address
	FileLength            uint16
	Checksum              uint16
	Interpreter           Interpreter
	Screen                Screen
	Font                  Font
	RoutinesAddr          Address
	StaticStringsAddr     Address
	BackgroundColor       uint8
	ForegroundColor       uint8
	TermCharsAddr         Address
	Stream3Width          uint16
	StandardRev           uint16
	AlphabetAddr          Address
	HeaderExtensionAddr   Address
}

type Interpreter struct {
	Number   uint8
	Revision uint8
}

type Screen struct {
	Height      uint8
	Width       uint8
	WidthUnits  uint8
	HeightUnits uint8
}

type Font struct {
	Height      uint8
	Width       uint8
	HeightUnits uint8
	WidthUnits  uint8
}

// const FLAGS1_None uint8 = 0
// const (
// 	_                 uint8 = 1 << iota
// 	FLAGS1_StatusLine       // 0 = score/turns, 1 = hours:mins
// 	FLAGS1_SplitStory
// 	FLAGS1_Tandy
// 	FLAGS1_StatusUnavailable
// 	FLAGS1_ScreenSplit
// 	FLAGS1_VariableFont
// 	_
// )

// const (
// 	FLAGS1_Colors uint8 = 1 << iota
// 	FLAGS1_Pictures
// 	FLAGS1_Boldface
// 	FLAGS1_Italics
// 	FLAGS1_Monospace
// 	FLAGS1_Sounds
// 	_
// 	FLAGS1_Timed
// )

// const FLAGS2_None uint16 = 0
// const (
// 	FLAGS2_Transcript uint16 = 1 << iota
// 	FLAGS2_ForceMono
// 	FLAGS2_Redraw
// 	FLAGS2_Pictures
// 	FLAGS2_Undo // FLAGS2_TLHSounds
// 	FLAGS2_Mouse
// 	FLAGS2_Color
// 	FLAGS2_Sounds
// 	FLAGS2_Menus
// 	_
// 	FLAGS2_PrintError
// 	_
// 	_
// 	_
// 	_
// 	_
// )
