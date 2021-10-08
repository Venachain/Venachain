package types

import "errors"

type DBEngineType int32

const (
	//UnknownDb don't know database
	UnknownDb DBEngineType = iota
	//LevelDb LevelDb
	LevelDb
	// PebbleDb PebbleDb
	PebbleDb
)

const (
	LevelDbStr  = "leveldb"
	PebbleDbStr = "pebbledb"
)

func ParseDbType(str string) (DBEngineType, error) {
	switch str {
	case LevelDbStr:
		return LevelDb, nil
	case PebbleDbStr:
		return PebbleDb, nil
	default:
		return UnknownDb, errors.New("unknown db type:" + str)
	}
}
