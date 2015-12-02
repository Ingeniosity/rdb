package main

import "github.com/codegangsta/cli"

const (
	kB = 1024
	MB = kB * 1024
	GB = MB * 1024
)

var (
	compressionTypeFlag = cli.StringFlag{
		Name:  "compression_type",
		Value: "snappy",
		Usage: "(lz4, snappy, zlib)"}

	numLevelsFlag = cli.IntFlag{
		Name:  "num_levels",
		Value: 7,
		Usage: "It is safe for num_levels to be bigger than expected number of levels in the database. Some higher levels may be empty, but this will not impact performance in any way. Only change this option if you expect your number of levels will be greater than 7",
	}

	writeBufferSizeFlag = cli.IntFlag{
		Name:  "write_buffer_size",
		Value: 4 * MB,
		Usage: "sets the size of a single memtable. Once memtable exceeds this size, it is marked immutable and a new one is created."}

	maxWriteBufferNumberFlag = cli.IntFlag{
		Name:  "max_write_buffer_number",
		Value: 2,
		Usage: "sets the maximum number of memtables, both active and immutable. If the active memtable fills up and the total number of memtables is larger than max_write_buffer_number we stall further writes. This may happen if the flush process is slower than the write rate.",
	}

	minWriteBufferNumberToMerge = cli.IntFlag{
		Name:  "min_write_buffer_number_to_merge",
		Value: 1,
		Usage: "is the minimum number of memtables to be merged before flushing to storage. For example, if this option is set to 2, immutable memtables are only flushed when there are two of them - a single immutable memtable will never be flushed. If multiple memtables are merged together, less data may be written to storage since two updates are merged to a single key. However, every Get() must traverse all immutable memtables linearly to check if the key is there. Setting this option too high may hurt read performance.",
	}

	level0_file_num_compaction_trigger = cli.IntFlag{
		Name:  "level0_file_num_compaction_trigger",
		Value: 4,
		Usage: `Once level 0 reaches this number of files, L0->L1 compaction is triggered. We can therefore estimate level 0 size in stable state as write_buffer_size * min_write_buffer_number_to_merge * level0_file_num_compaction_trigger.`,
	}

	level0_slowdown_writes_trigger = cli.IntFlag{Value: 20}
	level0_stop_writes_trigger     = cli.IntFlag{Value: 24}

	target_file_size_base = cli.IntFlag{
		Name:  "target_file_size_base",
		Value: 2 * MB,
		Usage: "Files in level 1 will have target_file_size_base bytes. Each next level's file size will be target_file_size_multiplier bigger than previous one. However, by default target_file_size_multiplier is 1, so files in all L1..Lmax levels are equal. Increasing target_file_size_base will reduce total number of database files, which is generally a good thing. We recommend setting target_file_size_base to be max_bytes_for_level_base / 10, so that there are 10 files in level 1.",
	}

	target_file_size_multiplier = cli.IntFlag{
		Name:  "target_file_size_multiplier",
		Value: 1,
	}

	max_bytes_for_level_base = cli.IntFlag{
		Name:  "max_bytes_for_level_base",
		Value: 10 * MB,
		Usage: "max_bytes_for_level_base is total size of level 1. As mentioned, we recommend that this be around the size of level 0. ",
	}
	max_bytes_for_level_multiplier = cli.IntFlag{
		Name:  "max_bytes_for_level_multiplier",
		Value: 10,
		Usage: "Each subsequent level is max_bytes_for_level_multiplier larger than previous one. The default is 10 and we do not recommend changing that."}
)
