package main

import "github.com/urfave/cli"

var (
	Version   string
	CommitID  string
	BuildDate string
	Compiler  string
)

func main() {
	app := cli.NewApp()
	app.Name = "AV-spy"
	app.Usage = ""
	app.UsageText = ""
	app.Version = Version
}
