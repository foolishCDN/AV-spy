package main

import (
	"flag"

	"github.com/awesome-gocui/gocui"
	"github.com/sirupsen/logrus"
)

var URL = flag.String("i", "", "input url")

var eventChan chan func(*gocui.Gui) error

func main() {
	logrus.SetReportCaller(true)

	flag.Parse()

	var g *gocui.Gui
	var err error
	for _, outputMode := range []gocui.OutputMode{gocui.Output256, gocui.Output216, gocui.OutputTrue, gocui.OutputNormal, gocui.OutputGrayscale} {
		g, err = gocui.NewGui(outputMode, true)
		if err == nil {
			break
		}
		logrus.Debugf("set output mode %v fail, err %v", outputMode, err)
	}
	if err != nil {
		logrus.Panicln(err)
	}
	defer g.Close()

	g.Mouse = false

	eventChan = make(chan func(*gocui.Gui) error, 100)
	go update(g)

	app := &App{}
	app.Init(g)

	if err := g.MainLoop(); err != nil && err != gocui.ErrQuit {
		logrus.Panicln(err)
	}
}

func submitEvent(event func(gui *gocui.Gui) error) {
	eventChan <- event
}

func update(g *gocui.Gui) {
	for event := range eventChan {
		g.UpdateAsync(event)
	}
}
