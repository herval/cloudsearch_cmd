package gocui

import (
	"context"
	"fmt"
	"github.com/herval/cloudsearch/pkg"
	"github.com/jroimartin/gocui"
	"github.com/sirupsen/logrus"
	"strings"
	"time"
)

const DefaultHint = "Type to search, ↑/↓ to move, ENTER to open, C-c to exit"

type SearchBar struct {
	x           int
	y           int
	w           int
	h           int
	engine      *SingleSearchHandler
	results     *ResultList
	hintMessage string
}

func NewSearchBar(x0, y0, w, h int, engine *cloudsearch.SearchEngine, results *ResultList) *SearchBar {
	return &SearchBar{
		x:           x0,
		y:           y0,
		w:           w,
		h:           h,
		results:     results,
		hintMessage: DefaultHint,
		engine: &SingleSearchHandler{
			e: engine,
			r: results,
		},
	}
}

func (s *SearchBar) SetHint(hint string) {
	if hint == "" {
		s.hintMessage = DefaultHint
	} else {
		s.hintMessage = hint
	}
}

func (s *SearchBar) Layout(g *gocui.Gui) error {
	if v, err := g.SetView("search", s.x, s.y, 2, 2); err != nil {
		if err != gocui.ErrUnknownView {
			return err
		}
		v.FgColor = gocui.ColorGreen
		v.Frame = false
		fmt.Fprintln(v, "?")
	}

	if v, err := g.SetView("hints", s.x, s.y+1, s.w-1, 3); err != nil {
		if err != gocui.ErrUnknownView {
			return err
		}
		v.Frame = false
		v.BgColor = gocui.AttrReverse
		fmt.Fprintln(v, s.hintMessage)
	}

	if v, err := g.SetView("search_bar", s.x+2, s.y, s.w-1, 2); err != nil {
		if err != gocui.ErrUnknownView {
			return err
		}

		s.engine.v = v
		s.engine.g = g

		v.Editable = true
		v.Frame = false
		v.Wrap = false
		v.Editor = gocui.EditorFunc(
			func(v *gocui.View, key gocui.Key, ch rune, mod gocui.Modifier) {
				switch key {
				case gocui.KeyArrowUp:
					s.results.Prev()
				case gocui.KeyArrowDown:
					s.results.Next()
				case gocui.KeyEnter:
					if s.results.IsSelected() {
						s.results.OpenSelected()
					} else {
						s.engine.Search()
						// TODO handle err
					}
				case gocui.KeyEsc:
					s.clearInput(v) // TODO doesn't capture?
				default:
					gocui.DefaultEditor.Edit(v, key, ch, mod)

					s.engine.Search()
					// TODO handle err
				}
			},
		)
	}

	return nil
}

func (s *SearchBar) clearInput(v *gocui.View) error {
	v.Clear()
	v.SetCursor(0, 0)
	return nil
}

func (s *SearchBar) Focus(g *gocui.Gui) error {
	g.Cursor = true
	g.SetCurrentView("search_bar")
	return nil
}

// handle multiple searches happening as user types, cancelling stale ones, etc
type SingleSearchHandler struct {
	e                   *cloudsearch.SearchEngine
	v                   *gocui.View
	g                   *gocui.Gui
	r                   *ResultList
	re                  *cloudsearch.Registry
	currentSearchCancel context.CancelFunc
}

func (s *SingleSearchHandler) Search() error {
	data := strings.TrimRightFunc(s.v.Buffer(), func(c rune) bool {
		return c == '\r' || c == '\n'
	})

	s.r.Clear()

	if s.currentSearchCancel != nil {
		logrus.Debug("Canceling previous query!")
		s.currentSearchCancel()
	}

	if len(data) < 2 {
		// no searching yet
		logrus.Debug("Input too short, ignoring...")
		return nil
	}

	ctx, cancel := context.WithCancel(context.Background())
	s.currentSearchCancel = cancel

	id := cloudsearch.NewId()
	res := s.e.Search(
		cloudsearch.ParseQuery(data, id, s.re),
		ctx,
	)

	done := false
	buff := make(chan cloudsearch.Result, 100)

	go func() {
		for !done {
			select {
			case r, ok := <-res:
				buff <- r
				if !ok {
					done = true
				}
			case <-ctx.Done():
				done = true
				return
			}
		}
	}()

	// update UI every 200ms, while search is ongoing
	go func() {
		for !done {
			time.Sleep(time.Millisecond * 200)
			s.g.Update(func(gui *gocui.Gui) error {
				// flush everything currently buffered to the ui
				buffDone := false
				for !buffDone {
					select {
					case r := <-buff:
						s.r.Append(r)
					default:
						buffDone = true
					}
				}

				return nil
			})
		}
	}()

	return nil
}
