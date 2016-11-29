// +build !embed

package rdb

// #cgo LDFLAGS: -lrocksdb -lstdc++ -lm -lz -lbz2 -lsnappy -llz4
// #cgo CXXFLAGS: -std=c++11
import "C"
