package main

import (
	"os"

	"github.com/codegangsta/cli"
)

func main() {
	app := cli.NewApp()
	app.Name = "slangouts"
	app.Usage = "Slack up front, Hangouts in the rear."
	app.Version = "0.0.1"
	app.Author = "gpavlidi"
	app.Email = "https://github.com/gpavlidi/slangouts"
	app.Flags = []cli.Flag{
		cli.IntFlag{
			Name:  "poll",
			Value: 10,
			Usage: "polling frequency for new Hangout messages",
		},
		cli.StringFlag{
			Name:  "config",
			Value: "~/.slangouts/config.json",
			Usage: "path to config file",
		},
	}
	app.Action = func(c *cli.Context) {
		path := c.String("config")
		// if config path not specifically set, let user.HomeDir find it
		if !c.IsSet("config") {
			path = ""
		}
		runSlangouts(c.Int("poll"), path)
	}

	app.Run(os.Args)

}
