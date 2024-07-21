package zmachine

import (
	"bytes"
	"encoding/binary"
	"os"
)

type ZMachine struct {
	Counter       uint8
	Memory        []uint8
	DynamicMemory []uint8
	StaticMemory  []uint8
	HighMemory    []uint8
	Stack         []uint8
}

const FLAGS1_None uint8 = 0
const (
	_                 uint8 = 1 << iota
	FLAGS1_StatusLine       // 0 = score/turns, 1 = hours:mins
	FLAGS1_SplitStory
	FLAGS1_Tandy
	FLAGS1_StatusUnavailable
	FLAGS1_ScreenSplit
	FLAGS1_VariableFont
	_
)

const (
	FLAGS1_Colors uint8 = 1 << iota
	FLAGS1_Pictures
	FLAGS1_Boldface
	FLAGS1_Italics
	FLAGS1_Monospace
	FLAGS1_Sounds
	_
	FLAGS1_Timed
)

const FLAGS2_None uint16 = 0
const (
	FLAGS2_Transcript uint16 = 1 << iota
	FLAGS2_ForceMono
	FLAGS2_Redraw
	FLAGS2_Pictures
	FLAGS2_Undo // FLAGS2_TLHSounds
	FLAGS2_Mouse
	FLAGS2_Color
	FLAGS2_Sounds
	FLAGS2_Menus
	_
	FLAGS2_PrintError
	_
	_
	_
	_
	_
)

type Header struct {
	Version               uint8
	Flags1                uint8
	ReleaseNumber         uint16
	HighMemoryAddr        uint16
	InitialProgramCounter uint16
	DictionaryAddr        uint16
	ObjectsAddr           uint16
	GlobalsAddr           uint16
	StaticMemory          uint16
	Flags2                uint16
	Serial                [6]byte
	AbbreviationsAddr     uint16
	FileLength            uint16
	Checksum              uint16
	Interpreter           Interpreter
	Screen                Screen
	Font                  Font
	RoutinesAddr          uint16
	StaticStringsAddr     uint16
	BackgroundColor       uint8
	ForegroundColor       uint8
	TermCharsAddr         uint16
	Stream3Width          uint16
	StandardRev           uint16
	AlphabetAddr          uint16
	HeaderExtensionAddr   uint16
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

func (zmachine ZMachine) getHeader() *Header {
	header := Header{}
	err := binary.Read(bytes.NewBuffer(zmachine.Memory[0:64]), binary.BigEndian, &header)
	if err != nil {
		panic(err)
	}

	return &header
}

func (zmachine *ZMachine) init(memory []uint8) {
	zmachine.Memory = memory

	header := zmachine.getHeader()
	zmachine.DynamicMemory = memory[:header.StaticMemory]
	zmachine.StaticMemory = memory[header.StaticMemory:min(len(memory), 0xffff)]
	zmachine.HighMemory = memory[header.HighMemoryAddr:]
}

func Load(story_path string) error {
	zmachine := ZMachine{}

	memory, err := os.ReadFile(story_path)
	if err != nil {
		panic(err)
	}

	zmachine.init(memory)

	print(zmachine.getHeader().Version)
	return nil
}
