package main

import (
	"log"
	"os"

	"ingen.io/rdb"

	"github.com/codegangsta/cli"
)

var app = cli.NewApp()

var (
	compressionTypeFlag = cli.StringFlag{
		Name:  "compression_type",
		Value: "snappy",
		Usage: "(lz4, snappy, zlib)",
	}
	writeBufferSizeFlag = cli.IntFlag{
		Name:  "write_buffer_size",
		Value: 5 * 1024,
		Usage: "buffer size in MB",
	}
	numLevelsFlag = cli.IntFlag{
		Name:  "num_levels",
		Value: 7,
	}
	minWriteBufferNumberToMerge = cli.IntFlag{
		Name:  "min_write_buffer_number_to_merge",
		Value: 4,
	}
)

func init() {
	app.Commands = append(app.Commands, cli.Command{
		Name:   "create",
		Usage:  "create a new rdb database",
		Action: createDb,

		Flags: []cli.Flag{
			compressionTypeFlag,
			writeBufferSizeFlag,
			numLevelsFlag,
			minWriteBufferNumberToMerge,
		},
	})
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
	app.Run(os.Args)
}

func createDb(c *cli.Context) {
	dbName := c.GlobalString("db")
	if dbName == "" {
		cli.ShowAppHelp(c)
		return
	}
	dbOptions := rdb.NewDefaultOptions()
	dbOptions.SetCreateIfMissing(true)
	dbOptions.SetNumLevels(c.Int(numLevelsFlag.Name))
	dbOptions.SetWriteBufferSize(c.Int(writeBufferSizeFlag.Name))

	switch c.String(compressionTypeFlag.Name) {
	case "snappy":
		dbOptions.SetCompression(rdb.SnappyCompression)
	case "zlib":
		dbOptions.SetCompression(rdb.ZLibCompression)
	case "lz4":
		dbOptions.SetCompression(rdb.Lz4Compression)
	default:
		dbOptions.SetCompression(rdb.Lz4Compression)
	}

	db, err := rdb.OpenDb(dbOptions, dbName)
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()
}
