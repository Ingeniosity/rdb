#pragma once


extern ROCKSDB_LIBRARY_API const char rocksdb_key_may_exist(
		rocksdb_t* db, 
		const rocksdb_readoptions_t* options,
		const char* key, size_t keylen, 
		char** errptr);


typedef struct rocksdb_switchable_memtable_factory rocksdb_switchable_memtable_factory;

extern ROCKSDB_LIBRARY_API rocksdb_switchable_memtable_factory* rocksdb_options_set_switchable_memtable_factory(rocksdb_options_t *opt);
extern ROCKSDB_LIBRARY_API void useSkipList(rocksdb_switchable_memtable_factory *f);
extern ROCKSDB_LIBRARY_API void useVector(rocksdb_switchable_memtable_factory *f);
