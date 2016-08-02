// +build !embed

package rdb

// #cgo LDFLAGS: -lrocksdb -lstdc++ -lm -lz -lbz2 -lsnappy
// #cgo CXXFLAGS: -std=c++11
import "C"
