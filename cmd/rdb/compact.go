package main

import (
	"fmt"
	"log"

	"github.com/codegangsta/cli"
	"github.com/unigraph/rdb"
)

func init() {
	app.Commands = append(app.Commands, cli.Command{
		Name:   "compact",
		Usage:  "compact rdb database",
		Action: createDb,

		Flags: []cli.Flag{
			compression_type,
			num_levels,
		},
	})

}

func compactDb(c *cli.Context) {
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
	db.CompactRange(rdb.Range{})
	fmt.Println("done")
	fmt.Println([]byte(db.GetProperty("rocksdb.stats")))
}
