package main

import (
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
		Value: "snappy",
		Usage: "(none, lz4, snappy, zlib, bzip"}

	num_levels = cli.IntFlag{
		Name:  "num_levels",
		Value: 7,
		Usage: "It is safe for num_levels to be bigger than expected number of levels in the database. Some higher levels may be empty, but this will not impact performance in any way. Only change this option if you expect your number of levels will be greater than 7",
	}

	write_buffer_size = cli.IntFlag{
		Name:  "write_buffer_size",
		Value: 4 * MB,
		Usage: "Size of a single memtable. Once memtable exceeds this size, it is marked immutable and a new one is created."}

	max_write_buffer_number = cli.IntFlag{
		Name:  "max_write_buffer_number",
		Value: 2,
		Usage: "Maximum number of memtables, both active and immutable. If the active memtable fills up and the total number of memtables is larger than max_write_buffer_number we stall further writes. This may happen if the flush process is slower than the write rate.",
	}

	min_write_buffer_number_to_merge = cli.IntFlag{
		Name:  "min_write_buffe_number_to_merge",
		Value: 1,
		Usage: "Minimum number of memtables to be merged before flushing to storage. For example, if this option is set to 2, immutable memtables are only flushed when there are two of them - a single immutable memtable will never be flushed. If multiple memtables are merged together, less data may be written to storage since two updates are merged to a single key. However, every Get() must traverse all immutable memtables linearly to check if the key is there. Setting this option too high may hurt read performance.",
	}

	level0_file_num_compaction_trigger = cli.IntFlag{
		Name:  "level0_file_num_compaction_trigger",
		Value: 4,
		Usage: `Once level 0 reaches this number of files, L0->L1 compaction is triggered. We can therefore estimate level 0 size in stable state as write_buffer_size * min_write_buffer_number_to_merge * level0_file_num_compaction_trigger.`,
	}

	level0_slowdown_writes_trigger = cli.IntFlag{
		Name:  "level0_slowdown_writes_trigger",
		Value: 20,
		Usage: "When the number of level 0 files is greater than the slowdown limit, writes are stalled."}

	level0_stop_writes_trigger = cli.IntFlag{
		Name:  "level0_stop_writes_trigger",
		Value: 24,
		Usage: "When the number is greater than stop limit, writes are fully stopped until compaction is done."}

	target_file_size_base = cli.IntFlag{
		Name:  "target_file_size_base",
		Value: 2 * MB,
		Usage: "Files in level 1 will have target_file_size_base bytes. Each next level's file size will be target_file_size_multiplier bigger than previous one. However, by default target_file_size_multiplier is 1, so files in all L1..Lmax levels are equal. Increasing target_file_size_base will reduce total number of database files, which is generally a good thing. We recommend setting target_file_size_base to be max_bytes_for_level_base / 10, so that there are 10 files in level 1.",
	}

	target_file_size_multiplier = cli.IntFlag{
		Name:  "target_file_size_multiplier",
		Value: 1,
		Usage: "Each next level's file size will be target_file_size_multiplier bigger than previous one. By default target_file_size_multiplier is 1, which means by default files in different levels will have similar size.",
	}

	max_bytes_for_level_base = cli.IntFlag{
		Name:  "max_bytes_for_level_base",
		Value: 10 * MB,
		Usage: "Total size of level 1. As mentioned, we recommend that this be around the size of level 0. ",
	}

	max_bytes_for_level_multiplier = cli.IntFlag{
		Name:  "max_bytes_for_level_multiplier",
		Value: 10,
		Usage: "Each subsequent level is max_bytes_for_level_multiplier larger than previous one. The default is 10 and we do not recommend changing that."}

	bulk = cli.BoolFlag{
		Name:  "bulk",
		Usage: "Sets options for bulk data load. Modifies level0_file_num_compaction_trigger (1GB), level0_slowdown_writes_trigger(1GB), level0_stop_writes_trigger(1GB), target_file_size_base(256MB)"}

	source_compaction_factor = cli.IntFlag{
		Name:  "source_compaction_factor",
		Value: 1,
		Usage: "Maximum number of bytes in all source files to be compacted in a single compaction run. We avoid picking too many files in the source level so that we do not exceed the total source bytes for compaction to exceed (source_compaction_factor * targetFileSizeLevel()) many bytes. Default:1, i.e. pick maxfilesize amount of data as the source of a compaction."}

	disable_auto_compactions = cli.BoolFlag{
		Name:  "disable_auto_compactions",
		Usage: "disables auto compactions",
	}
)

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
}

func (f flags) setOptions(dbOptions *rdb.Options, c *cli.Context) {
	setCompression(dbOptions, c.GlobalString(compression_type.Name))
	dbOptions.SetNumLevels(c.GlobalInt(num_levels.Name))
	dbOptions.SetWriteBufferSize(c.GlobalInt(write_buffer_size.Name))
	dbOptions.SetMaxWriteBufferNumber(c.GlobalInt(max_write_buffer_number.Name))
	dbOptions.SetMinWriteBufferNumberToMerge(c.GlobalInt(min_write_buffer_number_to_merge.Name))
	//
	dbOptions.SetLevel0FileNumCompactionTrigger(c.GlobalInt(level0_file_num_compaction_trigger.Name))
	dbOptions.SetLevel0SlowdownWritesTrigger(c.GlobalInt(level0_slowdown_writes_trigger.Name))
	dbOptions.SetLevel0StopWritesTrigger(c.GlobalInt(level0_stop_writes_trigger.Name))
	dbOptions.SetMaxBytesForLevelBase(uint64(c.GlobalInt(max_bytes_for_level_base.Name)))
	dbOptions.SetMaxBytesForLevelMultiplier(c.GlobalInt(max_bytes_for_level_multiplier.Name))
	dbOptions.SetTargetFileSizeBase(uint64(c.GlobalInt(target_file_size_base.Name)))
	dbOptions.SetTargetFileSizeMultiplier(c.GlobalInt(target_file_size_multiplier.Name))
	dbOptions.SetSourceCompactionFactor(c.GlobalInt(source_compaction_factor.Name))
	dbOptions.SetDisableAutoCompactions(c.GlobalBool(disable_auto_compactions.Name))

	if c.GlobalBool(bulk.Name) {
		dbOptions.PrepareForBulkLoad()
	}
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
