package gocui

import (
	"fmt"
	"github.com/herval/cloudsearch"
	"github.com/jroimartin/gocui"
	"github.com/skratchdot/open-golang/open"
	"strconv"
)

type ResultList struct {
	x       int
	y       int
	w       int
	h       int
	v       *gocui.View
	results []cloudsearch.Result
}

func NewResultsList(x0, y0, w, h int) *ResultList {
	return &ResultList{
		x0, y0, w, h, nil, []cloudsearch.Result{},
	}
}

func (r *ResultList) Layout(g *gocui.Gui) error {
	if v, err := g.SetView("results", r.x, r.y, r.w, r.h); err != nil {
		if err != gocui.ErrUnknownView {
			return err
		}
		v.Frame = false
		v.Highlight = true
		v.Autoscroll = true
		v.SelBgColor = gocui.ColorGreen
		r.v = v
		//fmt.Fprintln(v, "Hello world!")
	}

	return nil
}

func (r *ResultList) Clear() {
	r.v.Clear()
	r.results = []cloudsearch.Result{}
	r.v.SetCursor(0, 0)
}

func (r *ResultList) Append(result cloudsearch.Result) {
	r.results = append(r.results, result)

	// pad right
	l := fmt.Sprintf("%-"+strconv.Itoa(r.w)+"s\n", result.Title)

	r.v.Write([]byte(l))
}

func (r *ResultList) Prev() {
	x, y := r.v.Cursor()
	r.v.SetCursor(x, max(y-1, 0))
}

func (r *ResultList) Next() {
	x, y := r.v.Cursor()
	r.v.SetCursor(x, min(y+1, len(r.results)-1))
}

func (r *ResultList) IsSelected() bool {
	return len(r.results) > 0
}

func (r *ResultList) OpenSelected() {
	x, _ := r.v.Cursor()

	open.Run(r.results[x].Permalink)
}

// ugh...
func min(i int, i2 int) int {
	if i < i2 {
		return i
	}
	return i2
}

func max(i int, i2 int) int {
	if i > i2 {
		return i
	}
	return i2
}
