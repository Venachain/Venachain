package pebbledb

import (
	"bytes"
	"fmt"
	"github.com/cockroachdb/pebble"
	"testing"
)

func TestNewPebbleDBDatabase(t *testing.T) {
	db, err := NewPebbleDB("./data", 126, 10)
	if err != nil || db == nil {
		t.Errorf("new PebbleDB fail:%s\n", err)
		return
	}

	if db != nil {
		db.Close()
	}
}

func TestPebbleOper(t *testing.T) {
	db, err := NewPebbleDB("./data", 126, 10)
	if err != nil || db == nil {
		t.Errorf("new PebbleDB fail:%s\n", err)
		return
	}
	defer func() {
		if db != nil {
			db.Close()
		}
	}()

	k := []byte("TestPebble")
	v := []byte("PebbleValue123*&^")

	t.Run("pebble put", func(t *testing.T) {
		if err = db.Put(k, v); err != nil {
			t.Errorf("pebble put data fail:%s", err)
			return
		}

		has, _ := db.Has(k)
		if !has {
			t.Errorf("pebble del data wrong,Has")
		}
	})

	t.Run("pebble get", func(t *testing.T) {
		value, err := db.Get(k)
		if err != nil {
			t.Errorf("pebble get data fail:%s", err)
			return
		}
		if !bytes.Equal(value, v) {
			t.Errorf("pebble get data err, value:%v", value)
			return
		}
	})

	t.Run("pebble del", func(t *testing.T) {
		if err = db.Delete(k); err != nil {
			t.Errorf("pebble del data fail:%s", err)
			return
		}

		has, _ := db.Has(k)
		if has {
			t.Errorf("pebble del data wrong,Has")
		}

		value, err := db.Get(k)
		if err != pebble.ErrNotFound {
			t.Errorf("pebble del data wrong:%s,value:%v", err, value)
			return
		}
	})
}

func TestPebbleBatchOper(t *testing.T) {
	db, err := NewPebbleDB("./data", 126, 10)
	if err != nil || db == nil {
		t.Errorf("new PebbleDB fail:%s\n", err)
		return
	}
	defer func() {
		if db != nil {
			db.Close()
		}
	}()

	datas := [][2][]byte{
		{[]byte("k1"), []byte("v1")},
		{[]byte("k2"), []byte("v2")},
		{[]byte("k3"), []byte("v3")},
		{[]byte("k4"), []byte("v4")},
		{[]byte("k5"), []byte("v5")},
	}
	batch := db.NewBatch()

	t.Run("pebble batch put", func(t *testing.T) {
		for _, data := range datas {
			if err := batch.Put(data[0], data[1]); err != nil {
				t.Errorf("pebble batch put err:%s, key:%v", err, data[0])
				return
			}
		}

		if s := batch.ValueSize(); s != len(datas) {
			t.Errorf("pebble batch size error,batch size:%d", s)
			return
		}

		if err = batch.Write(); err != nil {
			t.Errorf("pebble batch write put fail:%s\n", err)
			return
		}

		iter := db.NewIterator()
		buf := bytes.Buffer{}
		for valid := iter.First(); valid; valid = iter.Next() {
			fmt.Fprintf(&buf, "%s: %s,", iter.Key(), iter.Value())
		}
		t.Logf("after batch write put:%s\n", buf.String())
	})

	t.Run("pebble batch del", func(t *testing.T) {
		batch.Reset()
		if s := batch.ValueSize(); s != 0 {
			t.Errorf("pebble batch reset fail,batch size:%d", s)
			return
		}

		delNum := 3
		for i := 0; i < delNum; i++ {
			if err := batch.Delete(datas[i][0]); err != nil {
				t.Errorf("pebble batch delete err:%s,key:%v\n", err, datas[i][0])
				return
			}
		}
		if s := batch.ValueSize(); s != delNum {
			t.Errorf("pebble batch del num error,batch size:%d", s)
			return
		}

		if err = batch.Write(); err != nil {
			t.Errorf("pebble batch write del fail:%s\n", err)
			return
		}

		iter := db.NewIterator()
		buf := bytes.Buffer{}
		for valid := iter.First(); valid; valid = iter.Next() {
			fmt.Fprintf(&buf, "%s: %s,", iter.Key(), iter.Value())
		}
		t.Logf("after batch write del:%s\n", buf.String())
	})
}

func TestPebbleDatabase_NewIteratorWithPrefix(t *testing.T) {
	db, err := NewPebbleDB("./data", 126, 10)
	if err != nil || db == nil {
		t.Errorf("new PebbleDB fail:%s\n", err)
		return
	}
	defer func() {
		if db != nil {
			db.Close()
		}
	}()

	datas := [][2][]byte{
		{[]byte("abcbbccc1"), []byte("v1")},
		{[]byte("abbcccbb2"), []byte("v2")},
		{[]byte("abcbccbdd3"), []byte("v3")},
		{[]byte("ceeeddff4"), []byte("v4")},
		{[]byte("ccssssdd5"), []byte("v5")},
	}
	batch := db.NewBatch()
	for _, data := range datas {
		if err := batch.Put(data[0], data[1]); err != nil {
			t.Errorf("pebble batch put err:%s, key:%v", err, data[0])
			return
		}
	}
	if err = batch.Write(); err != nil {
		t.Errorf("pebble batch write put fail:%s\n", err)
		return
	}

	t.Run("pebble iter", func(t *testing.T) {
		iter := db.NewIterator()
		buf := bytes.Buffer{}
		for valid := iter.First(); valid; valid = iter.Next() {
			fmt.Fprintf(&buf, "%s: %s,", iter.Key(), iter.Value())
		}
		t.Logf("pebble iter:%s\n", buf.String())
	})

	t.Run("pebble iter with prefix", func(t *testing.T) {
		{
			iter := db.NewIteratorWithPrefix([]byte("k"))
			buf := bytes.Buffer{}
			for valid := iter.First(); valid; valid = iter.Next() {
				fmt.Fprintf(&buf, "%s: %s,", iter.Key(), iter.Value())
			}
			t.Logf("pebble iter with prefix k:%s\n", buf.String())
		}

		{
			iter := db.NewIteratorWithPrefix([]byte("a"))
			buf := bytes.Buffer{}
			for valid := iter.First(); valid; valid = iter.Next() {
				fmt.Fprintf(&buf, "%s: %s,", iter.Key(), iter.Value())
			}
			t.Logf("pebble iter with prefix a:%s\n", buf.String())
		}
		{
			iter := db.NewIteratorWithPrefix([]byte("ab"))
			buf := bytes.Buffer{}
			for valid := iter.First(); valid; valid = iter.Next() {
				fmt.Fprintf(&buf, "%s: %s,", iter.Key(), iter.Value())
			}
			t.Logf("pebble iter with prefix ab:%s\n", buf.String())
		}
		{
			iter := db.NewIteratorWithPrefix([]byte("abc"))
			buf := bytes.Buffer{}
			for valid := iter.First(); valid; valid = iter.Next() {
				fmt.Fprintf(&buf, "%s: %s,", iter.Key(), iter.Value())
			}
			t.Logf("pebble iter with prefix abc:%s\n", buf.String())
		}
		{
			iter := db.NewIteratorWithPrefix([]byte("c"))
			buf := bytes.Buffer{}
			for valid := iter.First(); valid; valid = iter.Next() {
				fmt.Fprintf(&buf, "%s: %s,", iter.Key(), iter.Value())
			}
			t.Logf("pebble iter with prefix c:%s\n", buf.String())
		}
		{
			iter := db.NewIteratorWithPrefix([]byte("ce"))
			buf := bytes.Buffer{}
			for valid := iter.First(); valid; valid = iter.Next() {
				fmt.Fprintf(&buf, "%s: %s,", iter.Key(), iter.Value())
			}
			t.Logf("pebble iter with prefix ce:%s\n", buf.String())
		}
		{
			iter := db.NewIteratorWithPrefix([]byte("cc"))
			buf := bytes.Buffer{}
			for valid := iter.First(); valid; valid = iter.Next() {
				fmt.Fprintf(&buf, "%s: %s,", iter.Key(), iter.Value())
			}
			t.Logf("pebble iter with prefix cc:%s\n", buf.String())
		}
	})
}
