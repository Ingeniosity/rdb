package shard

import (
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"regexp"
	"sync"

	"github.com/unigraph/rdb"
)

var ShardNameFn = func(i uint) string { return fmt.Sprintf("%03d", i) }

type Shard struct {
	dbs []*rdb.DB
}

func Open(opts *rdb.Options, name string, shardsNum uint) (*Shard, error) {
	if err := checkValid(name, shardsNum); err != nil {
		return nil, err
	}
	s := &Shard{}
	for i := uint(0); i < shardsNum; i++ {
		sName := filepath.Join(name, ShardNameFn(i))
		db, err := rdb.OpenDb(opts, sName)
		if err != nil {
			s.Close()
			return nil, err
		}
		s.dbs = append(s.dbs, db)
	}
	return s, nil
}

func OpenForReadOnly(opts *rdb.Options, name string, shardsNum uint, errorIfLogFileExist bool) (*Shard, error) {
	if err := checkValid(name, shardsNum); err != nil {
		return nil, err
	}
	s := &Shard{}
	for i := uint(0); i < shardsNum; i++ {
		sName := filepath.Join(name, ShardNameFn(i))
		db, err := rdb.OpenDbForReadOnly(opts, sName, errorIfLogFileExist)
		if err != nil {
			s.Close()
			return nil, err
		}
		s.dbs = append(s.dbs, db)
	}
	return s, nil
}

type errors []error

func (e errors) Error() string {
	res := ""
	for _, err := range e {
		res += err.Error()
	}
	return res
}

func (s *Shard) Flush(opts *rdb.FlushOptions) error {
	wg := sync.WaitGroup{}
	wg.Add(len(s.dbs))
	err := errors(nil)
	l := sync.RWMutex{}
	for _, db := range s.dbs {
		go func(db *rdb.DB) {
			if e := db.Flush(opts); e != nil {
				l.Lock()
				err = append(err, e)
				l.Unlock()
			}
			wg.Done()
		}(db)
	}
	wg.Wait()
	if len(err) == 0 {
		return nil
	}
	return err
}

func (s *Shard) CompactRange(r rdb.Range) {
	wg := sync.WaitGroup{}
	wg.Add(len(s.dbs))
	for _, db := range s.dbs {
		go func(db *rdb.DB) {
			db.CompactRange(r)
			wg.Done()
		}(db)
	}
	wg.Wait()
}

func (s *Shard) DBs() []*rdb.DB {
	return append([]*rdb.DB(nil), s.dbs...)
}

func (s *Shard) Close() {
	wg := sync.WaitGroup{}
	wg.Add(len(s.dbs))
	for _, db := range s.dbs {
		go func(db *rdb.DB) {
			db.Close()
			wg.Done()
		}(db)
	}
	wg.Wait()
}

func GetShardNum(name string) uint {
	if files, err := ioutil.ReadDir(name); os.IsNotExist(err) {
		return 0
	} else {
		shards := map[string]bool{}
		for _, file := range files {
			if matched, _ := regexp.MatchString(`\d{3}`, file.Name()); matched {
				shards[file.Name()] = true
			}
		}
		return uint(len(shards))
	}
}

func checkValid(name string, shardsNum uint) error {
	if shardsNum == 0 || shardsNum > 999 {
		return fmt.Errorf("Number of shards has to be bigger than 0 and lower than 1000")
	}
	files, err := ioutil.ReadDir(name)
	if os.IsNotExist(err) { // does not exists, let's create empty
		return os.Mkdir(name, 0700)
	} else if err != nil { // some other error related to ReadDir
		return err
	} else { // exists, let's check the content
		shards := map[string]bool{}
		for _, file := range files {
			if matched, _ := regexp.MatchString(`\d{3}`, file.Name()); matched {
				shards[file.Name()] = true
			}
		}
		if len(shards) != 0 {
			if uint(len(shards)) != shardsNum {
				return fmt.Errorf("Wrong number of shards provided (%v)", len(shards))
			}
			for i := uint(0); i < shardsNum; i++ {
				sName := ShardNameFn(i)
				if !shards[sName] {
					return fmt.Errorf("Wrong number of shards provided (%v)", len(shards))
				}
			}
		}
	}
	return nil
}
