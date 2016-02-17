package rdb

// #include <stdlib.h>
// #include "rocksdb/c.h"
import "C"
import "unsafe"

// Slice is used as a wrapper for non-copy values
type Slice struct {
	data  *C.char
	size  C.size_t
	freed bool
}

// NewSlice returns a slice with the given data.
func NewSlice(data *C.char, size C.size_t) *Slice {
	return &Slice{data, size, false}
}

// Data returns the data of the slice.
func (s *Slice) Data() []byte {
	return charToByte(s.data, s.size)
}

// Size returns the size of the data.
func (s *Slice) Size() int {
	return int(s.size)
}

func (self *Slice) Free() {
	if !self.freed {
		C.rocksdb_free(unsafe.Pointer(self.data))
		self.freed = true
	}
}
