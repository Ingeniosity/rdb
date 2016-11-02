package rdb

// #include <stdlib.h>
// #include "rocksdb/c.h"
// #include "rocksdb_ext.h"
import "C"
import (
	"bytes"
	"errors"
	"reflect"
	"unsafe"
)

// Iterator provides a way to seek to specific keys and iterate through
// the keyspace from that point, as well as access the values of those keys.
//
// For example:
//
//      it := db.NewIterator(readOpts)
//      defer it.Close()
//
//      it.Seek([]byte("foo"))
//		for ; it.Valid(); it.Next() {
//          fmt.Printf("Key: %v Value: %v\n", it.Key().Data(), it.Value().Data())
// 		}
//
//      if err := it.Err(); err != nil {
//          return err
//      }
//
type Iterator struct {
	c       *C.rocksdb_iterator_t
	isValid C.uchar
}

// NewNativeIterator creates a Iterator object.
func NewNativeIterator(c unsafe.Pointer) *Iterator {
	return &Iterator{c: (*C.rocksdb_iterator_t)(c)}
}

// Valid returns false only when an Iterator has iterated past either the
// first or the last key in the database.
func (iter *Iterator) Valid() bool {
	return ucharToBool(iter.isValid)
}

// ValidForPrefix returns false only when an Iterator has iterated past the
// first or the last key in the database or the specified prefix.
func (iter *Iterator) ValidForPrefix(prefix []byte) bool {
	return ucharToBool(iter.isValid) && bytes.HasPrefix(iter.Key(), prefix)
}

// Consider changing as here
// https://github.com/siddontang/ledisdb/blob/master/store/rocksdb/iterator.go#L20:L38
//
// Key returns the key the iterator currently holds.
func (iter *Iterator) Key() []byte {
	var cLen C.size_t
	cKey := C.rocksdb_iter_key(iter.c, &cLen)
	if cKey == nil {
		return nil
	}
	return slice(unsafe.Pointer(cKey), int(C.int(cLen)))
	// return &Slice{cKey, cLen, true}
}

// Value returns the value in the database the iterator currently holds.
func (iter *Iterator) Value() []byte {
	var cLen C.size_t
	cVal := C.rocksdb_iter_value(iter.c, &cLen)
	if cVal == nil {
		return nil
	}
	// return &Slice{cVal, cLen, true}
	return slice(unsafe.Pointer(cVal), int(C.int(cLen)))
}

// Next moves the iterator to the next sequential key in the database.
func (iter *Iterator) Next() {
	iter.isValid = C.rocksdb_iter_next_ext(iter.c)
}

// Prev moves the iterator to the previous sequential key in the database.
func (iter *Iterator) Prev() {
	iter.isValid = C.rocksdb_iter_prev_ext(iter.c)
}

// SeekToFirst moves the iterator to the first key in the database.
func (iter *Iterator) SeekToFirst() {
	iter.isValid = C.rocksdb_iter_seek_to_first_ext(iter.c)
}

// SeekToLast moves the iterator to the last key in the database.
func (iter *Iterator) SeekToLast() {
	iter.isValid = C.rocksdb_iter_seek_to_last_ext(iter.c)
}

// Seek moves the iterator to the position greater than or equal to the key.
func (iter *Iterator) Seek(key []byte) {
	cKey := byteToChar(key)
	iter.isValid = C.rocksdb_iter_seek_ext(iter.c, cKey, C.size_t(len(key)))
}

// Err returns nil if no errors happened during iteration, or the actual
// error otherwise.
func (iter *Iterator) Err() error {
	var cErr *C.char
	C.rocksdb_iter_get_error(iter.c, &cErr)
	if cErr != nil {
		defer C.rocksdb_free(unsafe.Pointer(cErr))
		return errors.New(C.GoString(cErr))
	}
	return nil
}

// Close closes the iterator.
func (iter *Iterator) Close() {
	C.rocksdb_iter_destroy(iter.c)
	iter.c = nil
}

func ucharToBool(uc C.uchar) bool {
	if uc == C.uchar(0) {
		return false
	}
	return true
}
func slice(p unsafe.Pointer, n int) []byte {
	var b []byte
	pbyte := (*reflect.SliceHeader)(unsafe.Pointer(&b))
	pbyte.Data = uintptr(p)
	pbyte.Len = n
	pbyte.Cap = n
	return b
}
