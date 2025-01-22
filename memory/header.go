package memory

type Flags1 byte

// Version 1-3
const (
	_ Flags1 = 1 << iota
	Flags1_StatusLineType
	Flags1_SplitAcrossDiscs
	Flags1_Tandy
	Flags1_StatusLineNotAvailable
	Flags1_ScreenSplittingAvailable
	Flags1_VariablePitchFontDefault
	_
)

// Version 4+
const (
	Flags1_ColorsAvailable Flags1 = 1 << iota
	Flags1_PictureDisplayingAvailable
	Flags1_BoldfaceAvailable
	Flags1_ItalicAvailable
	Flags1_FixedSpaceStyleAvailable
	Flags1_SoundEffectsAvailable
	_
	Flags1_TimedKeyboardInputAvailable
)

type Flags2 word

const (
	Flags2_TranscriptingOn Flags2 = 1 << iota
	Flags2_ForceFixedPitchPrinting
	Flags2_RequestScreenRedraw
	Flags2_UsePictures
	Flags2_UseUNDO
	Flags2_UseMouse
	Flags2_UseColors
	Flags2_UseSoundEffects
	Flags2_UseMenus
	_
	Flags2_TranscriptionError
	_
	_
	_
	_
	_
)

const (
	Addr_ROM_B_Version Address = iota
	Addr_IROM_B_Flags1
	Addr_ROM_W_ReleaseNumber
	_
	Addr_ROM_A_HighMem
	_
	Addr_ROM_A_InitialProgramCounter
	_
	Addr_ROM_A_Dictionary
	_
	Addr_ROM_A_ObjectTable
	_
	Addr_ROM_A_Globals
	_
	Addr_ROM_A_StaticMem
	_
	Addr_RAM_W_Flags2
	_
	Addr_ROM_S_SerialCode // ASCII, 6 bytes
	_
	_
	_
	_
	_
	Addr_ROM_A_AbbreviationsTable
	_
	Addr_ROM_W_FileLength
	_
	Addr_ROM_W_Checksum
	_
	Addr_IROM_B_InterpreterNumber
	Addr_IROM_B_InterpreterRevision
	Addr_IROM_B_ScreenHeight
	Addr_IROM_B_ScreenWidth
	Addr_IROM_W_ScreenWidthUnits
	_
	Addr_IROM_W_ScreenHeightUnits
	_
	Addr_IROM_B_FontHeight
	Addr_IROM_B_FontWidth
	Addr_ROM_W_RoutinesOffset
	_
	Addr_ROM_W_StringsOffset
	_
	Addr_IROM_B_BGColor
	Addr_IROM_B_FGColor
	Addr_ROM_A_TermChars
	_
	Addr_IROM_W_Stream3Width
	_
	Address_StandardRev // 0x32 word
	_
	Addr_ROM_A_AlphabetTable
	_
	Addr_ROM_A_HeaderExtension
	_
	Addr_IROM_S_LoginName // ASCII, 8 bytes
	_
	_
	_
	_
	_
	_
	_
)

func (m Memory) GetVersion() int {
	return int(m.ReadByte(Addr_ROM_B_Version))
}

func (m Memory) GetInitialProgramCounter() Address {
	return Address(m.ReadWord(Addr_ROM_A_InitialProgramCounter))
}

func (m Memory) GetGlobalsAddress() Address {
	return Address(m.ReadWord(Addr_ROM_A_Globals))
}

func (m Memory) GetAbbreviationsAddress() Address {
	return Address(m.ReadWord(Addr_ROM_A_AbbreviationsTable))
}

func (m Memory) GetObjectsAddress() Address {
	return Address(m.ReadWord(Addr_ROM_A_ObjectTable))
}

func (m Memory) GetFlag1Bits(bits Flags1) bool {
	return m.ReadByte(Addr_IROM_B_Flags1)&byte(bits) != 0
}

func (m Memory) SetFlag1Bits(bits Flags1) {
	flags1 := m.ReadByte(Addr_IROM_B_Flags1)
	flags1 |= byte(bits)
	m.WriteByte(Addr_IROM_B_Flags1, flags1)
}

func (m Memory) GetFlag2Bits(bits Flags2) bool {
	return m.ReadWord(Addr_RAM_W_Flags2)&word(bits) != 0
}

func (m Memory) SetFlag2Bits(bits Flags2) {
	flags2 := m.ReadWord(Addr_RAM_W_Flags2)
	flags2 |= word(bits)
	m.WriteWord(Addr_RAM_W_Flags2, flags2)
}
