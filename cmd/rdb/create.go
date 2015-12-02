package main

import (
	"log"

	"github.com/codegangsta/cli"
	"ingen.io/rdb"
)

func init() {
	app.Commands = append(app.Commands, cli.Command{
		Name:   "create",
		Usage:  "create a new rdb database",
		Action: createDb,
	})
}

func createDb(c *cli.Context) {
	dbName := c.GlobalString("db")
	if dbName == "" {
		cli.ShowAppHelp(c)
		return
	}
	dbOptions := rdb.NewDefaultOptions()
	dbOptions.SetCreateIfMissing(true)
	defaultFlags.setOptions(dbOptions, c)

	db, err := rdb.OpenDb(dbOptions, dbName)
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()
}
