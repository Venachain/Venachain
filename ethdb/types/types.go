package types

import (
	"errors"
	"path/filepath"
)

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

var dbMeta = map[DBEngineType]struct{
	id     DBEngineType
	name   string
	suffix string
}{
	UnknownDb: {id: UnknownDb},
	LevelDb:   {id: LevelDb, name: LevelDbStr, suffix: ".ldb"},
	PebbleDb:  {id: PebbleDb, name: PebbleDbStr, suffix: ".sst"},
}

// db类型string->int
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

// db类型int->string
func GetDbName(t DBEngineType) (string) {
	if val, ok := dbMeta[t]; ok {
		return val.name
	}
	return ""
}

// 判断路径下是否存在数据库文件
func GetExistDBType(file string) DBEngineType {
	for i, d := range dbMeta {
		if i == 0 {
			continue // 对应的UnknownDb，无需做匹配
		}
		matches, _ := filepath.Glob(file+"/../chaindata/*"+ d.suffix)
		if len(matches) > 0 {
			return i
		}
	}
	return UnknownDb
}
