package rdb

// #include "rocksdb/c.h"
// #include "ext.h"
import "C"
import "unsafe"

func (db *DB) KeyMayExist(opts *ReadOptions, key []byte) bool {
	var cErr *C.char
	cKey := byteToChar(key)
	ret := C.rocksdb_key_may_exist(db.c, opts.c, cKey, C.size_t(len(key)), &cErr)
	if cErr != nil {
		defer C.rocksdb_free(unsafe.Pointer(cErr))
		return false
	}
	return ret != 0
}
