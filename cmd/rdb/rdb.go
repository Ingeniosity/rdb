package main

import (
	"os"

	"github.com/codegangsta/cli"
)

var app = cli.NewApp()

func init() {
	app.Flags = []cli.Flag{
		cli.StringFlag{
			Name:  "db",
			Usage: "database location (required)",
		},
	}
}

func main() {
	app.Name = "RodksDB Command Line Tool"
	app.Usage = "tool to manipulate RocksDB databases"
	app.Version = "dev"
	app.Run(os.Args)
}
