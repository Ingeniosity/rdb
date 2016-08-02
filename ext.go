package rdb

// #include "rocksdb/c.h"
// #include "ext.h"
import "C"

func (db *DB) KeyMayExist(opts *ReadOptions, key []byte) bool {
	var cErr *C.char
	cKey := byteToChar(key)
	return C.rocksdb_key_may_exist(db.c, opts.c, cKey, C.size_t(len(key)), &cErr) != 0
}
