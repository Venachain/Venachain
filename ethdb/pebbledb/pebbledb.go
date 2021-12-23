package pebbledb

import (
	"time"

	"github.com/Venachain/Venachain/ethdb/dbhandle"
	"github.com/Venachain/Venachain/log"
	"github.com/cockroachdb/pebble"
	"github.com/cockroachdb/pebble/bloom"
)

const (
	writePauseWarningThrottler = 1 * time.Minute
)

// pebble write options
var pebbleWO = pebble.NoSync

type PebbleDatabase struct {
	fn  string     // filename for reporting
	db  *pebble.DB // pebble instance
	log log.Logger // Contextual logger tracking the database path
}

// NewPebbleDatabase returns a LevelDB wrapped object.
func NewPebbleDB(file string, cache int, handles int) (*PebbleDatabase, error) {

	logger := log.New("database", file)
	logger.Info("Allocated cache and file handles", "file", file, "cache", cache, "handles", handles)

	// options
	minHandles := 16384
	var minCache int64 = 1 << 30

	if handles < minHandles { // handles配置默认值75
		handles = minHandles
	}

	var cacheSize int64
	cacheSize = int64(cache)
	if cacheSize < minCache {
		cacheSize = minCache
	}
	pebblecache := pebble.NewCache(cacheSize)

	opts := &pebble.Options{
		Cache:                       pebblecache, // 1GB, 单位B,pebble默认值8MB，venachain 默认1024MB
		DisableWAL:                  false,
		FormatMajorVersion:          pebble.FormatNewest,
		L0CompactionThreshold:       2,
		L0StopWritesThreshold:       1000,
		LBaseMaxBytes:               64 << 20, // 64 MB
		Levels:                      make([]pebble.LevelOptions, 7),
		MaxConcurrentCompactions:    3,
		MaxOpenFiles:                handles,  // 16384, pebble默认值1000，venachain 默认75
		MemTableSize:                64 << 20, // 64MB, venachain 默认 256MB
		MemTableStopWritesThreshold: 4,
	}
	for i := 0; i < len(opts.Levels); i++ {
		l := &opts.Levels[i]
		l.BlockSize = 32 << 10       // 32 KB // 单位B，pebble默认值4MB，venachain 默认4MB
		l.IndexBlockSize = 256 << 10 // 256 KB
		l.FilterPolicy = bloom.FilterPolicy(10)
		l.FilterType = pebble.TableFilter
		if i > 0 {
			l.TargetFileSize = opts.Levels[i-1].TargetFileSize * 2
		}
		l.EnsureDefaults()
	}
	opts.Levels[6].FilterPolicy = nil
	opts.FlushSplitBytes = opts.Levels[0].TargetFileSize
	opts.EnsureDefaults()

	// Open the db and recover any potential corruptions
	db, err := pebble.Open(file, opts)

	if err != nil {
		return nil, err
	}
	return &PebbleDatabase{
		fn:  file,
		db:  db,
		log: logger,
	}, nil
}

func (db *PebbleDatabase) Path() string {
	return db.fn
}

func (db *PebbleDatabase) Put(key []byte, value []byte) error {
	return db.db.Set(key, value, pebbleWO)
}

func (db *PebbleDatabase) Has(key []byte) (bool, error) {
	data, closer, err := db.db.Get(key)
	defer func() {
		if closer != nil {
			closer.Close()
		}
	}()
	if err != nil {
		return false, err
	}
	if data == nil {
		return false, nil
	}
	return true, nil
}

func (db *PebbleDatabase) Get(key []byte) ([]byte, error) {
	data, closer, err := db.db.Get(key)
	defer func() {
		if closer != nil {
			closer.Close()
		}
	}()

	if err != nil {
		return nil, err
	}

	return data, nil
}

func (db *PebbleDatabase) Delete(key []byte) error {
	return db.db.Delete(key, pebbleWO)
}

func (db *PebbleDatabase) NewIterator() *pebble.Iterator {
	return db.db.NewIter(nil)
}

func (db *PebbleDatabase) NewIteratorWithPrefix(prefix []byte) *pebble.Iterator {
	keyUpperBound := func(b []byte) []byte {
		end := make([]byte, len(b))
		copy(end, b)
		for i := len(end) - 1; i >= 0; i-- {
			end[i] = end[i] + 1
			if end[i] != 0 {
				return end[:i+1]
			}
		}
		return nil
	}

	prefixIterOptions := func(prefix []byte) *pebble.IterOptions {
		return &pebble.IterOptions{
			LowerBound: prefix,
			UpperBound: keyUpperBound(prefix),
		}
	}

	return db.db.NewIter(prefixIterOptions(prefix))
}

func (db *PebbleDatabase) Close() {
	err := db.db.Close()
	if err == nil {
		db.log.Info("Database closed")
	} else {
		db.log.Error("Failed to close database", "err", err)
	}
}

func (db *PebbleDatabase) PDB() *pebble.DB {
	return db.db
}

func (db *PebbleDatabase) NewBatch() dbhandle.Batch {
	return &pdbBatch{db: db.db, b: new(pebble.Batch)}
}

type pdbBatch struct {
	db   *pebble.DB
	b    *pebble.Batch
	size int
}

func (b *pdbBatch) Put(key, value []byte) error {
	b.b.Set(key, value, pebbleWO)
	b.size += 1
	return nil
}

func (b *pdbBatch) Delete(key []byte) error {
	b.b.Delete(key, pebbleWO)
	b.size += 1
	return nil
}

func (b *pdbBatch) Write() error {
	return b.db.Apply(b.b, nil)
}

func (b *pdbBatch) ValueSize() int {
	return b.size
}

func (b *pdbBatch) Reset() {
	b.b.Reset()
	b.size = 0
}
