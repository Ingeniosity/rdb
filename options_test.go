package rdb

import (
	"fmt"
	"io/ioutil"
	"runtime"
	"sync"
	"testing"
	"time"

	"github.com/facebookgo/ensure"
)

func Test_DynamicOptions(t *testing.T) {
	dir, err := ioutil.TempDir("", "gorocksdb-")
	fmt.Println(dir)
	ensure.Nil(t, err)

	opts := opts()
	f := opts.SetSwitchableMemtable()
	db, err := OpenDb(opts, dir)
	ensure.Nil(t, err)
	woptions := NewDefaultWriteOptions()
	woptions.DisableWAL(true)
	fmt.Println("start")

	insert := func() {
		wg := sync.WaitGroup{}
		for a := 0; a < runtime.NumCPU(); a++ {
			wg.Add(1)
			a := a
			go func() {
				for i := 0; i < 1000000; i++ {
					db.Put(woptions, []byte(fmt.Sprintf("%v.%v", i, a)), nil)
				}
				wg.Done()
			}()
		}
		wg.Wait()
	}
	start := time.Now()
	insert()
	fmt.Println("\tinsert", time.Since(start))

	f.UseVector()
	start = time.Now()
	db.Flush(NewDefaultFlushOptions())
	fmt.Println("\tflush", time.Since(start))

	start = time.Now()
	insert()
	fmt.Println("\tinsert", time.Since(start))

	start = time.Now()
	db.Flush(NewDefaultFlushOptions())
	fmt.Println("\tflush", time.Since(start))
}

func opts() *Options {
	// global options
	opt := NewDefaultOptions()
	opt.SetCreateIfMissing(true)
	opt.SetAllowMmapReads(true)
	opt.SetMaxOpenFiles(-1)
	opt.SetKeepLogFileNum(2)
	// write options (also affects the size of WAL if used)
	opt.SetWriteBufferSize(10 << 30)      // size per memtable
	opt.SetMaxWriteBufferNumber(8)        // num of memtables in memory
	opt.SetMinWriteBufferNumberToMerge(2) // merge n tables into L0 file
	// compaction options
	opt.SetDisableAutoCompactions(true)
	opt.SetNumLevels(2)                            // L0 and L1 only
	opt.SetCompression(NoCompression)              // no compression for quadstore terms
	opt.SetLevel0FileNumCompactionTrigger(1 << 30) // trigger compaction if num file L0 equals n
	opt.SetLevel0SlowdownWritesTrigger(1 << 30)    // never slow down
	opt.SetLevel0StopWritesTrigger(1 << 30)        // never stop
	// Level sizes options (based on L1 settings)
	opt.SetMaxBytesForLevelBase(256 << 30) // L1 size
	opt.SetTargetFileSizeBase(4 << 30)     // L1 file size
	opt.SetTargetFileSizeMultiplier(10)    // each additional level size multiplier
	opt.SetMaxBackgroundFlushes(2)         // flush threads
	return opt
}
