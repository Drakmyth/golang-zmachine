package screen

import (
	"os"

	"github.com/Drakmyth/golang-zmachine/assert"
	"github.com/gdamore/tcell/v2"
)

type Screen struct {
	screen     tcell.Screen
	Events     chan tcell.Event
	QuitEvents chan struct{}
}

func NewScreen() *Screen {
	s, err := tcell.NewScreen()
	assert.NoError(err, "Error initializing screen")

	s.Init()
	s.Clear()
	s.Show()

	quit := make(chan struct{})
	events := make(chan tcell.Event)

	go s.ChannelEvents(events, quit)

	return &Screen{
		screen:     s,
		Events:     events,
		QuitEvents: quit,
	}
}

func (s Screen) HandleEvent(ev tcell.Event) {
	switch ev := ev.(type) {
	case *tcell.EventResize:
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
	// width, height := s.screen.Size()
	// fmt.Printf("%d, %d", width, height)
	for i, r := range text {
		s.screen.SetContent(i, 0, r, []rune{}, tcell.StyleDefault)
	}
}
