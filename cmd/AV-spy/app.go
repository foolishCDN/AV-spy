package main

import (
	"context"
	"errors"
	"fmt"
	"io"
	"net/url"

	"github.com/awesome-gocui/gocui"

	"github.com/foolishCDN/AV-spy/container/flv"
	"github.com/mattn/go-runewidth"
)

type App struct {
	viewIndex int

	ctx    context.Context
	cancel context.CancelFunc

	tags        []flv.TagI
	showTagInfo bool
}

func (app *App) Init(g *gocui.Gui) {
	g.Cursor = true
	g.InputEsc = false
	g.Mouse = true
	g.BgColor = gocui.ColorDefault
	g.FgColor = gocui.ColorDefault
	if runewidth.IsEastAsian() {
		g.ASCII = true
	}

	g.SetManagerFunc(app.Layout)
	app.SetKeys(g)
}

func (app *App) Layout(g *gocui.Gui) error {
	maxX, maxY := g.Size()
	if maxX < MinWidth || maxY < MinHeight {
		if view, err := g.SetView(ErrorViewName, 0, 0, maxX-1, maxY-1, 0); err != nil {
			if err != gocui.ErrUnknownView {
				return err
			}
			view.Frame = true
			view.Wrap = true
			view.Title = "Error"
			g.Cursor = false
			_, _ = fmt.Fprintln(view, "Terminal is too small")
			return nil
		}
	}
	if _, err := g.View(ErrorViewName); err == nil {
		_ = g.DeleteView(ErrorViewName)
		_ = app.setView(g)
	}
	return app.InitViews(g)
}

func (app *App) InitViews(g *gocui.Gui) error {
	maxX, maxY := g.Size()
	for _, param := range ViewsParams() {
		x0 := param.Position.x0.getCoordinate(maxX + 1)
		y0 := param.Position.y0.getCoordinate(maxY + 1)
		x1 := param.Position.x1.getCoordinate(maxX + 1)
		y1 := param.Position.y1.getCoordinate(maxY + 1)
		if view, err := g.SetView(param.Name, x0, y0, x1, y1, 0); err != nil {
			if err != gocui.ErrUnknownView {
				return err
			}
			view.Title = param.Title
			view.Subtitle = param.SubTitle
			view.Editable = param.Editable
			view.Visible = param.Visible
			view.Wrap = param.Wrap
			view.Editor = param.Editor
			view.Autoscroll = param.Autoscroll
			setViewTextAndCursor(view, param.Text)
		}
	}

	if app.ctx == nil {
		_, _ = g.SetCurrentView(PathViewName)
	}
	return nil
}

func (app *App) SetKeys(g *gocui.Gui) {
	_ = g.SetKeybinding(AllViewName, gocui.KeyCtrlC, gocui.ModNone, quit)
	_ = g.SetKeybinding(AllViewName, gocui.KeyTab, gocui.ModNone, app.NextView)
	_ = g.SetKeybinding(AllViewName, gocui.KeyEnter, gocui.ModNone, app.SubmitOrStopRequest)
	_ = g.SetKeybinding(AllViewName, gocui.KeyCtrlR, gocui.ModNone, clearInfoView)

	_ = g.SetKeybinding(PathViewName, gocui.KeyEnter, gocui.ModNone, app.SubmitOrStopRequest)

	_ = g.SetKeybinding(TimestampViewName, gocui.KeyArrowUp, gocui.ModNone, scrollViewUpWith(app.showTagContent))
	_ = g.SetKeybinding(TimestampViewName, gocui.KeyArrowDown, gocui.ModNone, scrollViewDownWith(app.showTagContent))
	_ = g.SetKeybinding(TimestampViewName, gocui.KeyEnter, gocui.ModNone, app.switchTagVisible)

	_ = g.SetKeybinding(InfoViewName, gocui.KeyArrowUp, gocui.ModNone, scrollViewUp)
	_ = g.SetKeybinding(InfoViewName, gocui.KeyArrowDown, gocui.ModNone, scrollViewDown)
	_ = g.SetKeybinding(InfoViewName, gocui.KeyCtrlR, gocui.ModNone, clear)

	_ = g.SetKeybinding(TagViewName, gocui.KeyArrowUp, gocui.ModNone, scrollViewUp)
	_ = g.SetKeybinding(TagViewName, gocui.KeyArrowDown, gocui.ModNone, scrollViewDown)
}

func (app *App) NextView(g *gocui.Gui, _ *gocui.View) error {
	app.viewIndex = (app.viewIndex + 1) % len(ViewsNames)
	if app.showTagInfo {
		if ViewsNames[app.viewIndex] == InfoViewName {
			_, _ = g.SetCurrentView(TagViewName)
			return nil
		}
	}
	_ = app.setView(g)
	return nil
}

func (app *App) SubmitOrStopRequest(g *gocui.Gui, _ *gocui.View) error {
	if app.ctx != nil {
		select {
		case <-app.ctx.Done():
		default:
			app.cancel()
			return nil
		}
	}
	return app.SubmitRequest(g)
}

func (app *App) SubmitRequest(g *gocui.Gui) error {
	timestampView, _ := g.View(TimestampViewName)
	timestampView.Clear()
	latestTimestampView, _ := g.View(LatestTimestampViewName)
	latestTimestampView.Clear()
	app.tags = app.tags[:0]

	ctx, cancel := context.WithCancel(context.Background())
	app.ctx = ctx
	app.cancel = cancel
	go func(ctx context.Context) {
		defer func() {
			cancel()
		}()
		path := getViewValue(g, PathViewName)
		u, err := url.Parse(path)
		if err != nil {
			showError(g, err.Error())
			return
		}
		showInfo(g, "Sending request to %s\n", path)
		r, err := doRequest(ctx, u.String())
		if err != nil {
			showError(g, err.Error())
			return
		}
		defer func() {
			_ = r.Close()
		}()
		demuxer := new(flv.Demuxer)
		header, err := demuxer.ReadHeader(r)
		if err != nil {
			showError(g, "Parse flv header failed,  error: %v\n", err)
			return
		}
		showNotice(g, "Flv Header:\n\t\tVersion: %d\n\t\tHasVideo: %t\n\t\tHasAudio: %t\n\t\tHeaderSize: %d\n", header.Version, header.HasVideo, header.HasAudio, header.DataOffset)

		for {
			select {
			case <-ctx.Done():
				return
			default:
			}
			tag, err := demuxer.ReadTag(r)
			if err != nil {
				if errors.Is(err, io.EOF) {
					showWarning(g, "Receive EOF")
				} else if errors.Is(err, context.Canceled) {
					showInfo(g, "Stop request")
				} else {
					showError(g, "Parse flv tag failed,  error: %v\n", err)
				}
				return
			}
			app.onTag(g, tag)
		}
	}(ctx)
	return nil
}

func (app *App) setView(g *gocui.Gui) error {
	view, err := g.SetCurrentView(ViewsNames[app.viewIndex])
	if err == nil {
		view.Highlight = true
	}
	return nil
}

func (app *App) onTag(g *gocui.Gui, tag flv.TagI) {
	onTag(g, tag, nil)
	app.tags = append(app.tags, tag)
}

func (app *App) hiddenTagView(g *gocui.Gui) {
	app.showTagInfo = false
	tagView, _ := g.View(TagViewName)
	tagView.Clear()
	tagView.Visible = false
}

func (app *App) switchTagVisible(g *gocui.Gui, view *gocui.View) error {
	if app.showTagInfo {
		app.hiddenTagView(g)
		return nil
	}
	app.showTagInfo = true
	return app.showTagContent(g, view)
}

func (app *App) showTagContent(g *gocui.Gui, view *gocui.View) error {
	if !app.showTagInfo {
		return nil
	}
	app.showTagInfo = true

	_, lineIndex := view.Origin()
	if lineIndex >= len(app.tags) {
		showWarning(g, "please select tag!\n")
		return nil
	}
	_ = view.SetHighlight(lineIndex, true)
	_, err := view.Line(lineIndex)
	if err != nil {
		showWarning(g, "get line content fail, err %v\n", err)
		return nil
	}
	tag := app.tags[lineIndex]

	tagView, _ := g.View(TagViewName)
	tagView.Clear()
	tagView.Visible = true
	onTag(g, tag, tagView)
	return nil
}
