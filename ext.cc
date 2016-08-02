//  Copyright (c) 2011-present, Facebook, Inc.  All rights reserved.
//  This source code is licensed under the BSD-style license found in the
//  LICENSE file in the root directory of this source tree. An additional grant
//  of patent rights can be found in the PATENTS file in the same directory.
//
// Copyright (c) 2011 The LevelDB Authors. All rights reserved.
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file. See the AUTHORS file for names of contributors.


#include "rocksdb/db.h"
#include <string>

using namespace rocksdb;


extern "C" {
	struct rocksdb_t                 { DB*               rep; };
	struct rocksdb_readoptions_t {
		ReadOptions rep;
		Slice upper_bound; // stack variable to set pointer to in ReadOptions
	};

	unsigned char rocksdb_key_may_exist(
			rocksdb_t* db,
			const rocksdb_readoptions_t* options,
			const char* key, size_t keylen,
			char** errptr) {
		std::string tmp;
		rocksdb::ColumnFamilyHandle* cf  = db->rep->DefaultColumnFamily();
		return db->rep->KeyMayExist(options->rep, cf, Slice(key, keylen), &tmp, nullptr);
	}
}
