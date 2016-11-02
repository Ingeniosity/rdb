package rdb

import (
	"fmt"
	"reflect"
	"testing"
	"unsafe"

	"github.com/facebookgo/ensure"
)

func TestIterator(t *testing.T) {
	db := newTestDB(t, "TestIterator", nil)
	defer db.Close()

	// insert keys
	givenKeys := [][]byte{[]byte("key1"), []byte("key2"), []byte("key3")}
	wo := NewDefaultWriteOptions()
	for _, k := range givenKeys {
		ensure.Nil(t, db.Put(wo, k, []byte("val")))
	}

	ro := NewDefaultReadOptions()
	iter := db.NewIterator(ro)
	defer iter.Close()
	var actualKeys [][]byte
	for iter.SeekToFirst(); iter.Valid(); iter.Next() {
		key := make([]byte, 4)
		copy(key, iter.Key())
		actualKeys = append(actualKeys, key)
	}
	ensure.Nil(t, iter.Err())
	ensure.DeepEqual(t, actualKeys, givenKeys)
}

func XTestIteratorKeyMemoryLeak(t *testing.T) {
	db := newTestDB(t, "TestIteratorKeyMemoryLeak", nil)
	defer db.Close()

	wo := NewDefaultWriteOptions()
	for i := 0; i < 1<<20; i++ {
		ensure.Nil(t, db.Put(wo, []byte(fmt.Sprintf("key_%v", i)), []byte("some val")))
	}

	iter := db.NewIterator(NewDefaultReadOptions())
	keyPointers := map[uintptr]bool{}
	valPointers := map[uintptr]bool{}
	for iter.SeekToFirst(); iter.Valid(); iter.Next() {
		key := iter.Key()
		keyData := key
		val := iter.Value()
		kH := (*reflect.SliceHeader)(unsafe.Pointer(&keyData))
		vH := (*reflect.SliceHeader)(unsafe.Pointer(&val))
		keyPointers[kH.Data] = true
		valPointers[vH.Data] = true
	}
	iter.Close()
	if len(keyPointers) != 1 {
		t.Errorf("Wrong len, expected 1 got %v", len(keyPointers))
	}
	if len(valPointers) != 1<<20 {
		t.Errorf("Wrong len, expected 1<<20 got %v", len(valPointers))
	}
}
