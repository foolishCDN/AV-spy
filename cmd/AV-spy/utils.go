package main

import (
	"fmt"
	"io"
	"strings"

	"github.com/awesome-gocui/gocui"
	"github.com/fatih/color"
	"github.com/gobs/pretty"
)

func setViewTextAndCursor(view *gocui.View, s string) {
	view.Clear()
	s = strings.TrimSpace(s)
	_, _ = fmt.Fprint(view, s)
	_ = view.SetCursor(len(s), 0)
}

func quit(_ *gocui.Gui, _ *gocui.View) error {
	return gocui.ErrQuit
}

func showError(g *gocui.Gui, format string, params ...interface{}) {
	showInfo(g, color.RedString("Error! ")+format, params...)
}

func showWarning(g *gocui.Gui, format string, params ...interface{}) {
	showInfo(g, color.MagentaString(
		"Warning! ")+format, params...)
}

func showNotice(g *gocui.Gui, format string, params ...interface{}) {
	showInfo(g, color.YellowString("Notice! ")+format, params...)
}

func showInfo(g *gocui.Gui, format string, params ...interface{}) {
	submitEvent(func(gui *gocui.Gui) error {
		infoView, _ := g.View(InfoViewName)
		if len(params) == 0 {
			_, _ = fmt.Fprintln(infoView, format)
			return nil
		}
		_, _ = fmt.Fprintf(infoView, format, params...)
		return nil
	})
}

func getViewValue(g *gocui.Gui, name string) string {
	view, err := g.View(name)
	if err != nil {
		return ""
	}
	return strings.TrimSpace(view.Buffer())
}

func scrollViewUpWith(f func(*gocui.Gui) error) func(*gocui.Gui, *gocui.View) error {
	return func(g *gocui.Gui, view *gocui.View) error {
		scrollView(view, -1)
		return f(g)
	}
}

func scrollViewDownWith(f func(*gocui.Gui) error) func(*gocui.Gui, *gocui.View) error {
	return func(g *gocui.Gui, view *gocui.View) error {
		scrollView(view, 1)
		return f(g)
	}
}

func scrollViewUp(_ *gocui.Gui, view *gocui.View) error {
	scrollView(view, -1)
	return nil
}

func scrollViewDown(_ *gocui.Gui, view *gocui.View) error {
	scrollView(view, 1)
	return nil
}

func scrollView(view *gocui.View, dy int) {
	view.Autoscroll = false
	ox, oy := view.Origin()
	if oy+dy < 0 {
		dy = -oy
	}
	if _, err := view.Line(dy); dy > 0 && err != nil {
		dy = 0
	}
	_ = view.SetOrigin(ox, oy+dy)
}

func clearInfoView(g *gocui.Gui, _ *gocui.View) error {
	infoView, _ := g.View(InfoViewName)
	infoView.Clear()
	return nil
}

func clear(_ *gocui.Gui, view *gocui.View) error {
	view.Clear()
	return nil
}

func prettyPrintTo(out io.Writer, i interface{}) {
	switch t := i.(type) {
	case []byte:
		dump(out, t, 16)
		return
	}
	p := &pretty.Pretty{Indent: pretty.DEFAULT_INDENT, Out: out, NilString: pretty.DEFAULT_NIL, Compact: false}
	p.Println(i)
}

func dump(out io.Writer, by []byte, number int) {
	n := len(by)
	rowCount := 0
	stop := (n / number) * number
	k := 0
	for i := 0; i <= stop; i += number {
		k++
		if i+number < n {
			rowCount = number
		} else {
			rowCount = min(k*number, n) % number
		}

		_, _ = fmt.Fprintf(out, color.RedString("%04d ", i))
		for j := 0; j < rowCount; j++ {
			_, _ = fmt.Fprintf(out, "%02x  ", by[i+j])
		}
		for j := rowCount; j < 8; j++ {
			_, _ = fmt.Fprintf(out, "    ")
		}
		_, _ = fmt.Fprintf(out, "  '%s'\n", viewString(by[i:(i+rowCount)]))
	}

}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}

func viewString(b []byte) string {
	r := []rune(string(b))
	for i := range r {
		if r[i] < 32 || r[i] > 126 {
			r[i] = '.'
		}
	}
	return string(r)
}
