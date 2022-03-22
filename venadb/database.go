package venadb

import (
	"errors"

	"github.com/Venachain/Venachain/log"
	"github.com/Venachain/Venachain/venadb/dbhandle"
	"github.com/Venachain/Venachain/venadb/leveldb"
	"github.com/Venachain/Venachain/venadb/pebbledb"
	"github.com/Venachain/Venachain/venadb/types"
)

// New news database via the giving db type.
func New(dbType, file string, cache, handles int) (dbhandle.Database, error) {
	log.Info("new database", "dbtype", dbType)

	dbtype, err := types.ParseDbType(dbType)
	if err != nil {
		return nil, err
	}

	if t := types.GetExistDBType(file); t != types.UnknownDb && t != dbtype {
		dbtype = t
		log.Warn("db already exist", "type", types.GetDbName(t))
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
