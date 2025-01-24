package screen

import (
	"fmt"
	"os"

	"github.com/Drakmyth/golang-zmachine/assert"
	"github.com/gdamore/tcell/v2"
)

type Screen struct {
	screen           tcell.Screen
	Events           chan tcell.Event
	QuitEvents       chan struct{}
	cursorX, cursorY int
	Wordwrap         bool
}

func NewScreen() *Screen {
	s, err := tcell.NewScreen()
	assert.NoError(err, "Error initializing screen")

	s.Init()
	s.Clear()
	s.Show()

	_, height := s.Size()

	quit := make(chan struct{})
	events := make(chan tcell.Event)

	go s.ChannelEvents(events, quit)

	return &Screen{
		screen:     s,
		Events:     events,
		QuitEvents: quit,
		cursorX:    0,
		cursorY:    height - 1,
		Wordwrap:   true,
	}
}

func (s *Screen) End() {
	s.screen.Fini()
}

// TODO: Use this to handle input/resize events
func (s *Screen) HandleEvent(ev tcell.Event) {
	switch ev := ev.(type) {
	case *tcell.EventResize:
		_, height := s.screen.Size()
		s.cursorY = height - 1
		s.screen.Sync()
	case *tcell.EventKey:
		if ev.Key() == tcell.KeyEscape || ev.Key() == tcell.KeyCtrlC {
			s.screen.Fini()
			os.Exit(0)
		}
	}

	s.screen.Show()
}

func (s *Screen) PrintText(text string) {
	fmt.Print(text) // This will print to any output log, but not to the screen
	width, _ := s.screen.Size()

	for _, r := range text {
		if r == '\n' || (s.Wordwrap && s.cursorX >= width) {
			s.ScrollUp()
			s.cursorX = 0
			if r == '\n' {
				continue
			}
		}
		s.screen.SetContent(s.cursorX, s.cursorY, r, []rune{}, tcell.StyleDefault)
		s.cursorX++
	}
}

func (s *Screen) ScrollUp() {
	width, height := s.screen.Size()
	for y := 0; y < height; y++ {
		for x := 0; x < width; x++ {
			r, cr, style, _ := s.screen.GetContent(x, y)
			s.screen.SetContent(x, y-1, r, cr, style)
			s.screen.SetContent(x, y, ' ', nil, tcell.StyleDefault)
		}
	}
	s.screen.Show()
}
