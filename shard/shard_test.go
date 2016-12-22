package shard

import (
	"io/ioutil"
	"log"
	"os"
	"testing"

	"github.com/unigraph/rdb"
)

func TestOpen(t *testing.T) {
	dir := tmpLocation()
	defer os.RemoveAll(dir)

	dbOpts := rdb.NewDefaultOptions()
	dbOpts.SetCreateIfMissing(true)

	if _, err := Open(dbOpts, dir, 0); err == nil {
		t.Errorf("Expecting error")
	}

	sh, err := Open(dbOpts, dir, 5)
	if err != nil {
		t.Fatalf(err.Error())
	}
	if err := sh.Flush(rdb.NewDefaultFlushOptions()); err != nil {
		t.Errorf("Should not get error: %v", err)
	}
	sh.Close()
}

func TestOpenWrongPath(t *testing.T) {
	dbOpts := rdb.NewDefaultOptions()
	dbOpts.SetCreateIfMissing(true)
	if _, err := Open(dbOpts, "this#is&/some*wrong(path", 10); err == nil {
		t.Errorf("Expecting error")
	}
}

func TestCreateAndOpenWithWrongShardsNum(t *testing.T) {
	dir := tmpLocation()
	defer os.RemoveAll(dir)

	dbOpts := rdb.NewDefaultOptions()
	dbOpts.SetCreateIfMissing(true)

	sh, err := Open(dbOpts, dir, 5)
	if err != nil {
		t.Fatalf(err.Error())
	}
	sh.Close()
	if _, err := Open(dbOpts, dir, 6); err == nil {
		t.Errorf("Expecting error")
	}
	if _, err := OpenForReadOnly(dbOpts, dir, 1, false); err == nil {
		t.Errorf("Expecting error")
	}
	if n := GetShardNum(dir); n != 5 {
		t.Errorf("Wrong shards number returned, expected 5 got %v", n)
	}
}

func tmpLocation() string {
	name, err := ioutil.TempDir("", "shard")
	if err != nil {
		log.Fatal(err)
	}
	return name
}
