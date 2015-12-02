package main

import (
	"fmt"
	"log"

	"github.com/codegangsta/cli"

	"ingen.io/rdb"
)

func init() {
	app.Commands = append(app.Commands, cli.Command{
		Name:   "compact",
		Usage:  "compact rdb database",
		Action: createDb,

		Flags: []cli.Flag{
			compressionTypeFlag,
			numLevelsFlag,
		},
	})

}

func compactDb(c *cli.Context) {
	dbName := c.GlobalString("db")
	if dbName == "" {
		cli.ShowAppHelp(c)
		return
	}

	options := rdb.NewDefaultOptions()
	// options.SetNumLevels(c.Int(numLevelsFlag.Name))
	// options.SetCompression(rdb.Lz4Compression)
	// options.SetDisableAutoCompactions(true)
	// options.SetTargetFileSizeBase(5 * GB)
	db, err := rdb.OpenDb(options, dbName)
	if err != nil {
		log.Fatal(err)
	}
	db.CompactRange(rdb.Range{})
	fmt.Println("done")
	fmt.Println([]byte(db.GetProperty("rocksdb.stats")))
}
