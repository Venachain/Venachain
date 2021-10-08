package ethdb

import (
	"errors"
	"github.com/PlatONEnetwork/PlatONE-Go/ethdb/dbhandle"
	"github.com/PlatONEnetwork/PlatONE-Go/ethdb/leveldb"
	"github.com/PlatONEnetwork/PlatONE-Go/ethdb/pebbledb"
	"github.com/PlatONEnetwork/PlatONE-Go/ethdb/types"
	"github.com/PlatONEnetwork/PlatONE-Go/log"
)

// New news database via the giving db type.
func New(dbType, file string, cache, handles int) (dbhandle.Database, error) {
	log.Info("new database","dbtype", dbType)

	dbtype, err := types.ParseDbType(dbType)
	if err != nil {
		return nil, err
	}

	switch dbtype {
	case types.LevelDb:
		return leveldb.NewLDBDatabase(file, cache, handles)
	case types.PebbleDb:
		return pebbledb.NewPebbleDB(file, cache, handles)
	default:
		return nil, errors.New("not support yet")
	}
}
