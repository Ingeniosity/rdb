package main

import (
	"crypto/md5"
	"fmt"
	"hash"
	"log"
	"sync"
	"time"

	"github.com/codegangsta/cli"
	"github.com/unigraph/rdb"
)

func init() {
	app.Commands = append(app.Commands, cli.Command{
		Name:   "generate",
		Usage:  "generates random keys and values",
		Action: randomData,
		Flags: []cli.Flag{
			cli.IntFlag{
				Name:  "valuesize,vs",
				Value: 128,
				Usage: "value size in bytes",
			},
			cli.IntFlag{
				Name:  "batchsize,bs",
				Value: 100000,
				Usage: "batch size",
			},
			cli.IntFlag{
				Name:  "pairs",
				Value: 1 * MB,
				Usage: "key/value pairs to generate",
			},
			cli.BoolFlag{
				Name:  "wal",
				Usage: "disable/enable WAL (write ahead log)",
			},
			cli.BoolFlag{
				Name:  "stats",
				Usage: "disable/enable periodic stats dump",
			},
		},
	})
}

var (
	hashPool = sync.Pool{New: func() interface{} { return md5.New() }}
	termSize = md5.Size // 16bytes
)

func hashOf(s string) []byte {
	h := hashPool.Get().(hash.Hash)
	h.Reset()
	defer hashPool.Put(h)
	key := make([]byte, 0, termSize)
	h.Write([]byte(s))
	key = h.Sum(key)
	return key
}

func randomData(c *cli.Context) error {
	dbName := c.GlobalString("db")
	if dbName == "" {
		cli.ShowAppHelp(c)
		return nil
	}
	dbOptions := rdb.NewDefaultOptions()
	dbOptions.SetCreateIfMissing(true)
	defaultFlags.setOptions(dbOptions, c)

	db, err := rdb.OpenDb(dbOptions, dbName)
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()

	value := make([]byte, c.Int("valuesize"))

	iterations := c.Int("pairs")
	woptions := rdb.NewDefaultWriteOptions()

	woptions.DisableWAL(!c.Bool("wal"))
	batchSize := c.Int("batchsize")

	batch := rdb.NewWriteBatch()
	log.Println("starting...")
	if c.Bool("stats") {
		go func() {
			for {
				time.Sleep(1 * time.Second)
				fmt.Println(db.GetProperty("rocksdb.stats"))
			}
		}()
	}
	for i := 0; i < iterations; i++ {
		key := hashOf(fmt.Sprintf("%016d", i))
		batch.Put(key, value)
		if batch.Count() > batchSize {
			db.Write(woptions, batch)
			batch.Clear()
		}
		// db.Put(woptions, key, key)
	}
	db.Write(woptions, batch)
	db.Flush(rdb.NewDefaultFlushOptions())
	fmt.Println(db.GetProperty("rocksdb.stats"))
	return nil
}
