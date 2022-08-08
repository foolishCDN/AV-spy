package main

import "github.com/awesome-gocui/gocui"

type Editor struct {
	editor gocui.Editor
}

func (e Editor) Edit(view *gocui.View, key gocui.Key, ch rune, mod gocui.Modifier) {
	switch {
	case (ch != 0 || key == gocui.KeySpace) && mod == 0:
		e.editor.Edit(view, key, ch, mod)
		// At the end of the line the default gcui editor adds a whitespace
		// Force him to remove
		ox, _ := view.Cursor()
		if ox > 1 && ox >= len(view.Buffer())-2 {
			view.EditDelete(false)
		}
		return
	//case key == gocui.KeyEnter:
	//	return
	case key == gocui.KeyHome || key == gocui.KeyCtrlA:
		view.SetCursor(0, 0)
		return
	case key == gocui.KeyEnd || key == gocui.KeyCtrlE:
		view.SetCursor(len(view.Buffer()), 0)
		return
	}
	e.editor.Edit(view, key, ch, mod)
}
