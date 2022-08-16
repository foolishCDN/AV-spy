package main

import (
	"context"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/http/httputil"
	"net/url"
	"time"

	"github.com/fatih/color"

	"github.com/awesome-gocui/gocui"

	"github.com/foolishCDN/AV-spy/container/flv"
	"github.com/mattn/go-runewidth"
)

type App struct {
	viewIndex int

	ctx    context.Context
	cancel context.CancelFunc

	avc        []*flv.VideoTag
	aac        []*flv.AudioTag
	videoTags  []*flv.VideoTag
	audioTags  []*flv.AudioTag
	scriptTags []*flv.ScriptTag

	tags          []flv.TagI
	isShowTagInfo bool
	isShowNetwork bool
}

func (app *App) Init(g *gocui.Gui) {
	g.Cursor = true
	g.InputEsc = false
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
	_ = g.SetKeybinding(AllViewName, gocui.KeyCtrlC, gocui.ModNone, app.quit)
	_ = g.SetKeybinding(AllViewName, gocui.KeyTab, gocui.ModNone, app.NextView)
	_ = g.SetKeybinding(AllViewName, gocui.KeyEnter, gocui.ModNone, app.SubmitOrStopRequest)
	_ = g.SetKeybinding(AllViewName, gocui.KeyCtrlR, gocui.ModNone, clearInfoView)
	_ = g.SetKeybinding(AllViewName, gocui.KeyCtrlQ, gocui.ModNone, app.switchViewVisible(g, NetworkViewName))

	_ = g.SetKeybinding(PathViewName, gocui.KeyEnter, gocui.ModNone, app.SubmitOrStopRequest)

	_ = g.SetKeybinding(TimestampViewName, gocui.KeyArrowUp, gocui.ModNone, scrollViewUpWith(app.showTagInfo))
	_ = g.SetKeybinding(TimestampViewName, gocui.KeyArrowDown, gocui.ModNone, scrollViewDownWith(app.showTagInfo))
	_ = g.SetKeybinding(TimestampViewName, gocui.KeyEnter, gocui.ModNone, app.switchViewVisible(g, TagViewName))

	_ = g.SetKeybinding(InfoViewName, gocui.KeyArrowUp, gocui.ModNone, scrollViewUp)
	_ = g.SetKeybinding(InfoViewName, gocui.KeyArrowDown, gocui.ModNone, scrollViewDown)
	_ = g.SetKeybinding(InfoViewName, gocui.KeyCtrlR, gocui.ModNone, clear)

	_ = g.SetKeybinding(TagViewName, gocui.KeyArrowUp, gocui.ModNone, scrollViewUp)
	_ = g.SetKeybinding(TagViewName, gocui.KeyArrowDown, gocui.ModNone, scrollViewDown)

	_ = g.SetKeybinding(NetworkViewName, gocui.KeyArrowUp, gocui.ModNone, scrollViewUp)
	_ = g.SetKeybinding(NetworkViewName, gocui.KeyArrowDown, gocui.ModNone, scrollViewDown)
}

func (app *App) NextView(g *gocui.Gui, _ *gocui.View) error {
	app.viewIndex = (app.viewIndex + 1) % len(ViewsNames)
	if app.isShowTagInfo {
		if ViewsNames[app.viewIndex] == InfoViewName {
			_, _ = g.SetCurrentView(TagViewName)
			return nil
		}
	}
	_ = app.setView(g)
	return nil
}

func (app *App) quit(_ *gocui.Gui, _ *gocui.View) error {
	if app.ctx != nil {
		select {
		case <-app.ctx.Done():
		default:
			app.cancel()
			return nil
		}
	}
	return gocui.ErrQuit
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

func (app *App) clearPerRequest(g *gocui.Gui) {
	timestampView, _ := g.View(TimestampViewName)
	timestampView.Clear()
	latestTimestampView, _ := g.View(LatestTimestampViewName)
	latestTimestampView.Clear()

	app.hiddenView(g, TagViewName)
	app.hiddenView(g, NetworkViewName)

	app.avc = app.avc[:0]
	app.aac = app.aac[:0]
	app.videoTags = app.videoTags[:0]
	app.audioTags = app.audioTags[:0]
	app.scriptTags = app.scriptTags[:0]

	app.tags = app.tags[:0]
}

func (app *App) SubmitRequest(g *gocui.Gui) error {
	app.clearPerRequest(g)

	ctx, cancel := context.WithCancel(context.Background())
	app.ctx = ctx
	app.cancel = cancel
	go func(ctx context.Context) {
		defer func() {
			cancel()
		}()
		path := getViewValue(g, PathViewName)
		_, err := url.Parse(path)
		if err != nil {
			showError(g, err.Error())
			return
		}
		req, err := http.NewRequest(http.MethodGet, path, nil)
		if err != nil {
			showError(g, err.Error())
			return
		}
		var trace Trace
		req = req.WithContext(ctx)
		req = req.WithContext(WithTrace(req.Context(), &trace))

		submitEvent(func(gui *gocui.Gui) error {
			networkView, _ := g.View(NetworkViewName)
			dump, err := httputil.DumpRequestOut(req, false)
			if err != nil {
				_, _ = fmt.Fprintf(networkView, "%s\n", err.Error())
				return nil
			}
			_, _ = fmt.Fprintf(networkView, "%s", dump)
			return nil
		})

		client := &http.Client{
			Transport: &http.Transport{
				ResponseHeaderTimeout: 10 * time.Second,
			},
		}
		showInfo(g, color.CyanString("Sending request to: ")+
			color.BlueString("\n\t%s\n", path)+
			color.CyanString("Press Ctrl-C or Enter to Stop\n")+
			color.CyanString("Press Ctrl-Q to show request info\n"))
		resp, err := client.Do(req)
		submitEvent(func(gui *gocui.Gui) error {
			networkView, _ := g.View(NetworkViewName)
			_, _ = fmt.Fprint(networkView, trace.Pretty())
			return nil
		})
		if err != nil {
			showError(g, err.Error())
			return
		}
		body := resp.Body
		submitEvent(func(gui *gocui.Gui) error {
			networkView, _ := g.View(NetworkViewName)
			resp.Body = nil
			dump, err := httputil.DumpResponse(resp, false)
			if err != nil {
				_, _ = fmt.Fprintf(networkView, "%s\n", err.Error())
				return nil
			}
			_, _ = fmt.Fprintf(networkView, "%s", dump)
			return nil
		})
		defer func() {
			_ = body.Close()
		}()

		demuxer := new(flv.Demuxer)
		header, err := demuxer.ReadHeader(body)
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
			tag, err := demuxer.ReadTag(body)
			if err != nil {
				if errors.Is(err, io.EOF) {
					showWarning(g, "Receive EOF")
				} else if errors.Is(err, context.Canceled) {
					showInfo(g, color.CyanString("Stop request, Press Ctrl-C to Quit"))
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
	switch t := tag.(type) {
	case *flv.VideoTag:
		if t.PacketType == flv.SequenceHeader {
			if len(app.avc) > 0 {
				showWarning(g, "Receive new avc, %d\n", len(app.avc)+1)
			}
			showNotice(g, "Receive avc, DTS %d PTS %d, size %d\n", t.DTS, t.PTS, len(t.Data()))
			app.avc = append(app.avc, t)
			return
		}
		if len(app.videoTags) > 0 {
			now := int(tag.Timestamp())
			last := int(app.videoTags[len(app.videoTags)-1].Timestamp())
			diff := now - last
			if now < last || diff > 100 {
				showWarning(g, "Video timestamp skip %d, now %d -> last %d\n", diff, now, last)
			}
		}
		app.videoTags = append(app.videoTags, t)
	case *flv.AudioTag:
		if t.SoundFormat == flv.AAC && t.PacketType == flv.SequenceHeader {
			if len(app.aac) > 0 {
				showWarning(g, "Receive new aac, %d\n", len(app.aac)+1)
			}
			showNotice(g, "Receive aac, timestamp %d, size %d\n", t.PTS, len(t.Data()))
			app.aac = append(app.aac, t)
			return
		}
		if len(app.audioTags) > 0 {
			now := int(tag.Timestamp())
			last := int(app.audioTags[len(app.audioTags)-1].Timestamp())
			diff := now - last
			if now < last || diff > 100 {
				showWarning(g, "Audio timestamp skip %d, now %d -> last %d\n", diff, now, last)
			}
		}
		app.audioTags = append(app.audioTags, t)
	case *flv.ScriptTag:
		if len(app.scriptTags) > 0 {
			showWarning(g, "Receive new script tag %d\n", len(app.scriptTags)+1)
		}
		showNotice(g, "Receive script tag, timestamp %d, size: %d\n", t.PTS, len(t.Data()))
		app.scriptTags = append(app.scriptTags, t)
	}
}

func (app *App) hiddenView(g *gocui.Gui, viewName string) {
	switch viewName {
	case TagViewName:
		app.isShowTagInfo = false
	case NetworkViewName:
		app.isShowNetwork = false
	}
	view, _ := g.View(viewName)
	view.Clear()
	view.Visible = false
}

func (app *App) switchViewVisible(g *gocui.Gui, viewName string) func(*gocui.Gui, *gocui.View) error {
	return func(_ *gocui.Gui, _ *gocui.View) error {
		switch viewName {
		case TagViewName:
			if app.isShowTagInfo {
				app.hiddenView(g, TagViewName)
				return nil
			}
			app.isShowTagInfo = true
			return app.showTagInfo(g)
		case NetworkViewName:
			if app.isShowNetwork {
				app.hiddenView(g, NetworkViewName)
				return nil
			}
			g.Cursor = true
			_, _ = g.SetCurrentView(NetworkViewName)
			_, _ = g.SetViewOnTop(NetworkViewName)
			app.isShowNetwork = true
			return app.showNetwork(g)
		}
		return nil
	}
}

func (app *App) showTagInfo(g *gocui.Gui) error {
	if !app.isShowTagInfo {
		return nil
	}
	app.isShowTagInfo = true

	timestampView, _ := g.View(TimestampViewName)
	_, lineIndex := timestampView.Origin()
	if lineIndex >= len(app.tags) {
		showWarning(g, "please select tag!\n")
		return nil
	}
	_ = timestampView.SetHighlight(lineIndex, true)
	_, err := timestampView.Line(lineIndex)
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

func (app *App) showNetwork(g *gocui.Gui) error {
	if !app.isShowNetwork {
		return nil
	}
	app.isShowNetwork = true

	networkView, _ := g.View(NetworkViewName)
	networkView.Visible = true
	return nil
}
