//  Copyright (c) 2011-present, Facebook, Inc.  All rights reserved.
//  This source code is licensed under the BSD-style license found in the
//  LICENSE file in the root directory of this source tree. An additional grant
//  of patent rights can be found in the PATENTS file in the same directory.
//
// Copyright (c) 2011 The LevelDB Authors. All rights reserved.
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file. See the AUTHORS file for names of contributors.


#include "rocksdb/db.h"
#include "rocksdb/memtablerep.h"
#include <string>
#include <iostream>

using namespace rocksdb;

extern "C" {
	struct rocksdb_t                 { DB*               rep; };
	struct rocksdb_readoptions_t {
		ReadOptions rep;
		Slice upper_bound; // stack variable to set pointer to in ReadOptions
	};
	struct rocksdb_options_t         { Options           rep; };



	unsigned char rocksdb_key_may_exist(
			rocksdb_t* db,
			const rocksdb_readoptions_t* options,
			const char* key, size_t keylen,
			char** errptr) {
		std::string tmp;
		rocksdb::ColumnFamilyHandle* cf  = db->rep->DefaultColumnFamily();
		return db->rep->KeyMayExist(options->rep, cf, Slice(key, keylen), &tmp, nullptr);
	}

	class SwitchableMemTableFactory: public MemTableRepFactory {
		public:
			char isSkipList;
			MemTableRepFactory *vector, *skipList;
		public:
			explicit SwitchableMemTableFactory(): isSkipList(true) {
				this->vector = new VectorRepFactory();
				this->skipList = new SkipListFactory();
			}

			virtual MemTableRep* CreateMemTableRep(
					const MemTableRep::KeyComparator& a,
					MemTableAllocator* b,
					const SliceTransform* c,
					Logger* d) override {
				if (this->isSkipList) {
					return this->skipList->CreateMemTableRep(a,b,c,d);
				} else { 
					return this->vector->CreateMemTableRep(a,b,c,d);
				}
			}

			virtual const char* Name() const override {
				return "SwitchableMemTableFactory";
			}
	};

	struct rocksdb_switchable_memtable_factory { SwitchableMemTableFactory* rep; };

	rocksdb_switchable_memtable_factory* rocksdb_options_set_switchable_memtable_factory(rocksdb_options_t *opt) {
		SwitchableMemTableFactory *factory = new SwitchableMemTableFactory();
		opt->rep.memtable_factory.reset(factory);
		rocksdb_switchable_memtable_factory* res = new rocksdb_switchable_memtable_factory;
		res->rep = factory;
		return res;
	}

	void useSkipList(rocksdb_switchable_memtable_factory *f) {
		f->rep->isSkipList = 1;
	}

	void useVector(rocksdb_switchable_memtable_factory *f) {
		f->rep->isSkipList = 0;
	}


}
