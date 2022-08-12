package main

import (
	"github.com/awesome-gocui/gocui"
)

const (
	MinWidth  = 60
	MinHeight = 20
)

const (
	AllViewName   = ""
	ErrorViewName = "error"

	PathViewName            = "path"
	InfoViewName            = "info"
	TimestampViewName       = "timestamp"
	LatestTimestampViewName = "latest_timestamp"
	TagViewName             = "tag"
	NetworkViewName         = "network"
)

var ViewsNames = []string{
	PathViewName,
	InfoViewName,
	TimestampViewName,
}

var ViewsParams = func() []ViewParam {
	return []ViewParam{
		{
			Name:     PathViewName,
			Title:    "URL",
			SubTitle: "Press Enter to Request or Stop",
			Editable: true,
			Wrap:     true,
			Visible:  true,
			Editor:   &Editor{gocui.DefaultEditor},
			Position: ViewPosition{
				x0: position{0.0, 0},
				y0: position{0.0, 0},
				x1: position{1.0, -2},
				y1: position{0.0, 3},
			},
			Text: *URL,
		},
		{
			Name:       InfoViewName,
			Title:      "Info",
			SubTitle:   "Press Ctrl-R to Clear",
			Editor:     gocui.DefaultEditor,
			Autoscroll: true,
			Wrap:       true,
			Visible:    true,
			Position: ViewPosition{
				x0: position{0.0, 0},
				y0: position{0.0, 3},
				x1: position{0.5, -2},
				y1: position{1.0, -2},
			},
		},
		{
			Name:     TimestampViewName,
			Title:    "Timestamp",
			SubTitle: "Press Enter to Show More Info",
			Editor:   gocui.DefaultEditor,
			Wrap:     true,
			Visible:  true,
			Position: ViewPosition{
				position{0.5, -2},
				position{0.0, 3},
				position{1.0, -2},
				position{1.0, -5}},
		},
		{
			Name:       LatestTimestampViewName,
			Title:      "Latest Timestamp",
			Editor:     gocui.DefaultEditor,
			Autoscroll: true,
			Wrap:       true,
			Visible:    true,
			Position: ViewPosition{
				position{0.5, -2},
				position{1.0, -5},
				position{1.0, -2},
				position{1.0, -2}},
		},
		{
			Name:   TagViewName,
			Title:  "Tag Info",
			Editor: gocui.DefaultEditor,
			Wrap:   true,
			Position: ViewPosition{
				x0: position{0.0, 0},
				y0: position{0.0, 3},
				x1: position{0.5, -2},
				y1: position{1.0, -2},
			},
		},
		{
			Name:   NetworkViewName,
			Title:  "Network",
			Editor: gocui.DefaultEditor,
			Wrap:   true,
			Position: ViewPosition{
				x0: position{0.25, 0},
				y0: position{0.15, 0},
				x1: position{0.75, 0},
				y1: position{1.0, -2},
			},
		},
	}
}

type ViewParam struct {
	Name     string
	Title    string
	SubTitle string

	Editable   bool
	Autoscroll bool
	Wrap       bool
	Visible    bool

	gocui.Editor
	Text     string
	Position ViewPosition
}

type position struct {
	pct float32
	abs int
}

func (p position) getCoordinate(max int) int {
	return int(p.pct*float32(max)) + p.abs
}

type ViewPosition struct {
	x0, y0, x1, y1 position
}
