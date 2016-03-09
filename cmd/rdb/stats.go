package main

import (
	"fmt"
	"log"

	"github.com/codegangsta/cli"
	"github.com/unigraph/rdb"
)

func init() {
	app.Commands = append(app.Commands, cli.Command{
		Name:   "stats",
		Usage:  "print rdb stats",
		Action: statsDb,
	})

}

func statsDb(c *cli.Context) {
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
	fmt.Println(db.GetProperty("rocksdb.stats"))
	db.Close()
}
