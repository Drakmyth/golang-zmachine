package screen

import (
	"os"
	"strings"

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

func (s *Screen) Read() string {
	stopReading := false
	buffer := strings.Builder{}

	for !stopReading {
		ev := <-s.Events
		switch eventType := ev.(type) {
		// TODO: Not sure if this is the right place to handle EventResize
		// case *tcell.EventResize:
		// 	_, height := s.screen.Size()
		// 	s.cursorY = height - 1
		// 	s.screen.Sync()
		case *tcell.EventKey:
			switch eventType.Key() {
			case tcell.KeyEscape, tcell.KeyCtrlC:
				// TODO: Replace with call to ZMachine.Shutdown(0)
				s.screen.Fini()
				os.Exit(0)
			case tcell.KeyEnter:
				// TODO: In V5+ this should check for terminating characters rather than only newline
				stopReading = true
			case tcell.KeyRune:
				buffer.WriteRune(eventType.Rune())
			}
		}
	}

	return buffer.String()
}

func (s *Screen) PrintText(text string) {
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
