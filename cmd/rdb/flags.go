package main

import (
	"encoding/json"
	"fmt"
	"log"
	"os"
	"reflect"

	"github.com/codegangsta/cli"
	"github.com/unigraph/rdb"
)

const (
	kB = 1024
	MB = kB * 1024
	GB = MB * 1024
)

// https://github.com/facebook/rocksdb/blob/master/include/rocksdb/options.h
var (
	compression_type = cli.StringFlag{
		Name:  "compression_type",
		Value: DefaultOptions.CompressionType,
		Usage: "(none, lz4, snappy, zlib, bzip"}

	num_levels = cli.IntFlag{
		Name:  "num_levels",
		Value: DefaultOptions.NumLevels,
		Usage: "It is safe for num_levels to be bigger than expected number of levels in the database. Some higher levels may be empty, but this will not impact performance in any way. Only change this option if you expect your number of levels will be greater than 7",
	}

	write_buffer_size = cli.IntFlag{
		Name:  "write_buffer_size",
		Value: DefaultOptions.WriteBufferSize,
		Usage: "Size of a single memtable. Once memtable exceeds this size, it is marked immutable and a new one is created."}

	max_write_buffer_number = cli.IntFlag{
		Name:  "max_write_buffer_number",
		Value: DefaultOptions.MaxWriteBufferNumber,
		Usage: "Maximum number of memtables, both active and immutable. If the active memtable fills up and the total number of memtables is larger than max_write_buffer_number we stall further writes. This may happen if the flush process is slower than the write rate.",
	}

	min_write_buffer_number_to_merge = cli.IntFlag{
		Name:  "min_write_buffe_number_to_merge",
		Value: DefaultOptions.MinWriteBufferNumberToMerge,
		Usage: "Minimum number of memtables to be merged before flushing to storage. For example, if this option is set to 2, immutable memtables are only flushed when there are two of them - a single immutable memtable will never be flushed. If multiple memtables are merged together, less data may be written to storage since two updates are merged to a single key. However, every Get() must traverse all immutable memtables linearly to check if the key is there. Setting this option too high may hurt read performance.",
	}

	level0_file_num_compaction_trigger = cli.IntFlag{
		Name:  "level0_file_num_compaction_trigger",
		Value: DefaultOptions.Level0FileNumCompactionTrigger,
		Usage: `Once level 0 reaches this number of files, L0->L1 compaction is triggered. We can therefore estimate level 0 size in stable state as write_buffer_size * min_write_buffer_number_to_merge * level0_file_num_compaction_trigger.`,
	}

	level0_slowdown_writes_trigger = cli.IntFlag{
		Name:  "level0_slowdown_writes_trigger",
		Value: DefaultOptions.Level0SlowdownWritesTrigger,
		Usage: "When the number of level 0 files is greater than the slowdown limit, writes are stalled."}

	level0_stop_writes_trigger = cli.IntFlag{
		Name:  "level0_stop_writes_trigger",
		Value: DefaultOptions.Level0StopWritesTrigger,
		Usage: "When the number is greater than stop limit, writes are fully stopped until compaction is done."}

	target_file_size_base = cli.IntFlag{
		Name:  "target_file_size_base",
		Value: DefaultOptions.TargetFileSizeBase,
		Usage: "Files in level 1 will have target_file_size_base bytes. Each next level's file size will be target_file_size_multiplier bigger than previous one. However, by default target_file_size_multiplier is 1, so files in all L1..Lmax levels are equal. Increasing target_file_size_base will reduce total number of database files, which is generally a good thing. We recommend setting target_file_size_base to be max_bytes_for_level_base / 10, so that there are 10 files in level 1.",
	}

	target_file_size_multiplier = cli.IntFlag{
		Name:  "target_file_size_multiplier",
		Value: DefaultOptions.TargetFileSizeMultiplier,
		Usage: "Each next level's file size will be target_file_size_multiplier bigger than previous one. By default target_file_size_multiplier is 1, which means by default files in different levels will have similar size.",
	}

	max_bytes_for_level_base = cli.IntFlag{
		Name:  "max_bytes_for_level_base",
		Value: DefaultOptions.MaxBytesForLevelBase,
		Usage: "Total size of level 1. As mentioned, we recommend that this be around the size of level 0. ",
	}

	max_bytes_for_level_multiplier = cli.IntFlag{
		Name:  "max_bytes_for_level_multiplier",
		Value: DefaultOptions.MaxBytesForLevelMultiplier,
		Usage: "Each subsequent level is max_bytes_for_level_multiplier larger than previous one. The default is 10 and we do not recommend changing that."}

	bulk = cli.BoolFlag{
		Name:  "bulk",
		Usage: "Sets options for bulk data load. Modifies level0_file_num_compaction_trigger (1GB), level0_slowdown_writes_trigger(1GB), level0_stop_writes_trigger(1GB), target_file_size_base(256MB)"}

	source_compaction_factor = cli.IntFlag{
		Name:  "source_compaction_factor",
		Value: DefaultOptions.SourceCompactionFactor,
		Usage: "Maximum number of bytes in all source files to be compacted in a single compaction run. We avoid picking too many files in the source level so that we do not exceed the total source bytes for compaction to exceed (source_compaction_factor * targetFileSizeLevel()) many bytes. Default:1, i.e. pick maxfilesize amount of data as the source of a compaction."}

	disable_auto_compactions = cli.BoolFlag{
		Name:  "disable_auto_compactions",
		Usage: "disables auto compactions",
	}

	disable_wal = cli.BoolFlag{
		Name:  "disable_wal",
		Usage: "disables wal (write ahead log)",
	}

	config = cli.StringFlag{
		Name:  "config",
		Usage: "rdb configuration file with options (alternative to command line parameters)",
	}
)

var DefaultOptions = Options{
	CompressionType:                "snappy",
	NumLevels:                      7,
	WriteBufferSize:                4 * MB,
	MaxWriteBufferNumber:           2,
	MinWriteBufferNumberToMerge:    2,
	Level0FileNumCompactionTrigger: 4,
	Level0SlowdownWritesTrigger:    20,
	Level0StopWritesTrigger:        24,
	TargetFileSizeBase:             2 * MB,
	TargetFileSizeMultiplier:       1,
	MaxBytesForLevelBase:           10 * MB,
	MaxBytesForLevelMultiplier:     10,
	SourceCompactionFactor:         1,
	DisableAutoCompactions:         false,
	DisableWriteAheadLog:           false,
	Bulk:                           false,
}

type Options struct {
	CompressionType                string
	NumLevels                      int
	WriteBufferSize                int
	MaxWriteBufferNumber           int
	MinWriteBufferNumberToMerge    int
	Level0FileNumCompactionTrigger int
	Level0SlowdownWritesTrigger    int
	Level0StopWritesTrigger        int
	TargetFileSizeBase             int
	TargetFileSizeMultiplier       int
	MaxBytesForLevelBase           int
	MaxBytesForLevelMultiplier     int
	SourceCompactionFactor         int
	DisableAutoCompactions         bool
	DisableWriteAheadLog           bool
	Bulk                           bool
}

type flags []cli.Flag

var defaultFlags = flags{
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
	source_compaction_factor,
	disable_auto_compactions,
	bulk,
	config,
}

func (o *Options) Update(c *cli.Context) {
	o.CompressionType = c.GlobalString(compression_type.Name)
	o.NumLevels = c.GlobalInt(num_levels.Name)
	o.WriteBufferSize = c.GlobalInt(write_buffer_size.Name)
	o.MaxWriteBufferNumber = c.GlobalInt(max_write_buffer_number.Name)
	o.MinWriteBufferNumberToMerge = c.GlobalInt(min_write_buffer_number_to_merge.Name)
	o.Level0FileNumCompactionTrigger = (c.GlobalInt(level0_file_num_compaction_trigger.Name))
	o.Level0SlowdownWritesTrigger = (c.GlobalInt(level0_slowdown_writes_trigger.Name))
	o.Level0StopWritesTrigger = (c.GlobalInt(level0_stop_writes_trigger.Name))
	o.MaxBytesForLevelBase = (c.GlobalInt(max_bytes_for_level_base.Name))
	o.MaxBytesForLevelMultiplier = (c.GlobalInt(max_bytes_for_level_multiplier.Name))
	o.TargetFileSizeBase = (c.GlobalInt(target_file_size_base.Name))
	o.TargetFileSizeMultiplier = (c.GlobalInt(target_file_size_multiplier.Name))
	o.SourceCompactionFactor = (c.GlobalInt(source_compaction_factor.Name))
	o.DisableAutoCompactions = (c.GlobalBool(disable_auto_compactions.Name))
	if c.GlobalBool(bulk.Name) {
		o.Bulk = true
	}
}

func (o *Options) SetOptions(dbOptions *rdb.Options) {
	setCompression(dbOptions, o.CompressionType)
	if o.Bulk {
		dbOptions.PrepareForBulkLoad()
	}
	dbOptions.SetNumLevels(o.NumLevels)
	dbOptions.SetWriteBufferSize(o.WriteBufferSize)
	dbOptions.SetMaxWriteBufferNumber(o.MaxWriteBufferNumber)
	dbOptions.SetMinWriteBufferNumberToMerge(o.MinWriteBufferNumberToMerge)
	//
	dbOptions.SetLevel0FileNumCompactionTrigger(o.Level0FileNumCompactionTrigger)
	dbOptions.SetLevel0SlowdownWritesTrigger(o.Level0SlowdownWritesTrigger)
	dbOptions.SetLevel0StopWritesTrigger(o.Level0StopWritesTrigger)
	dbOptions.SetMaxBytesForLevelBase(uint64(o.MaxBytesForLevelBase))
	dbOptions.SetMaxBytesForLevelMultiplier(o.MaxBytesForLevelMultiplier)
	dbOptions.SetTargetFileSizeBase(uint64(o.TargetFileSizeBase))
	dbOptions.SetTargetFileSizeMultiplier(o.TargetFileSizeMultiplier)
	dbOptions.SetSourceCompactionFactor(o.SourceCompactionFactor)
	dbOptions.SetDisableAutoCompactions(o.DisableAutoCompactions)

}

func (f flags) setOptions(dbOptions *rdb.Options, c *cli.Context) {
	configOptions := map[string]interface{}{}
	if configFile := c.GlobalString("config"); configFile != "" {
		file, err := os.Open(configFile)
		if err != nil {
			log.Fatal(err)
		}
		dec := json.NewDecoder(file)
		err = dec.Decode(&configOptions)
		if err != nil {
			log.Fatal(err)
		}
	}

	if c.GlobalBool("bulk") {
		configOptions["NumLevels"] = float64(2)
		configOptions["Level0FileNumCompactionTrigger"] = float64(1 * GB)
		configOptions["Level0SlowdownWritesTrigger"] = float64(1 * GB)
		configOptions["Level0StopWritesTrigger"] = float64(1 * GB)
		configOptions["TargetFileSizeBase"] = float64(256 * MB)
		configOptions["DisableAutoCompactions"] = true
		configOptions["WriteBufferSize"] = float64(512 * MB)
		configOptions["MinWriteBufferNumberToMerge"] = float64(8)
		configOptions["MaxWriteBufferNumber"] = float64(16)
		configOptions["SourceCompactionFactor"] = float64(1 * GB)
		// DefaultOptions.CompressionType = "lz4"
		// configOptions["CompressionType"] = "lz4"
	}

	DefaultOptions.Update(c)
	val := reflect.ValueOf(&DefaultOptions).Elem()
	for k, v := range configOptions {
		switch a := v.(type) {
		case float64:
			val.FieldByName(k).SetInt(int64(a))
		case string:
			val.FieldByName(k).SetString(a)
		case bool:
			val.FieldByName(k).SetBool(a)
		default:
			fmt.Println(reflect.TypeOf(a))
		}
	}
	DefaultOptions.SetOptions(dbOptions)
}

func setCompression(dbOptions *rdb.Options, compressionType string) {
	switch compressionType {
	case "none":
		dbOptions.SetCompression(rdb.NoCompression)
	case "snappy":
		dbOptions.SetCompression(rdb.SnappyCompression)
	case "zlib":
		dbOptions.SetCompression(rdb.ZLibCompression)
	case "bzip":
		dbOptions.SetCompression(rdb.Bz2Compression)
	case "lz4":
		dbOptions.SetCompression(rdb.Lz4Compression)
	default:
		dbOptions.SetCompression(rdb.Lz4Compression)
	}
}
