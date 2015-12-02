package main

import (
	"fmt"
	"log"

	"github.com/codegangsta/cli"
	"ingen.io/rdb"
)

func init() {
	app.Commands = append(app.Commands, cli.Command{
		Name:   "create",
		Usage:  "create a new rdb database",
		Action: createDb,

		Flags: []cli.Flag{
			compression_type,
			num_levels,
			write_buffer_size,
			max_write_buffer_number,
			min_write_buffer_number_to_merge,
			level0_file_num_compaction_trigger,
			level0_slowdown_writes_trigger,
			level0_stop_writes_trigger,
			max_bytes_for_level_base,
			max_bytes_for_level_multiplier,
			target_file_size_base,
			target_file_size_multiplier,
		},
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
	// dbOptions.SetNumLevels(c.Int(numLevelsFlag.Name))
	// dbOptions.SetWriteBufferSize(c.Int(writeBufferSizeFlag.Name))

	switch c.String(compression_type.Name) {
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
	fmt.Println(db.GetProperty("rocksdb.stats"))
}
