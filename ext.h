#pragma once


extern ROCKSDB_LIBRARY_API const char rocksdb_key_may_exist(
		rocksdb_t* db, 
		const rocksdb_readoptions_t* options,
		const char* key, size_t keylen, 
		char** errptr);
