package core

import (
	"runtime"
	"testing"
	"time"

	"github.com/panjf2000/ants/v2"

	"github.com/Venachain/Venachain/core/types"
)

func TestDependency(t *testing.T) {
	dag := makeDag()
	indexs, left, all := initDag(dag)
	if len(indexs) != 4 {
		t.Errorf("transaction count mismatch: have %d, want %d", len(indexs), 4)
	}
	if len(all) != 4 {
		t.Errorf("all dependency count mismatch: have %d, want %d", len(all), 4)
	}
	indexs = getNoneDependency(0, left, all)
	if len(indexs) != 0 {
		t.Errorf("transaction count mismatch: have %d, want %d", len(indexs), 0)
	}

	indexs = getNoneDependency(1, left, all)
	if len(indexs) != 0 {
		t.Errorf("transaction count mismatch: have %d, want %d", len(indexs), 0)
	}
	indexs = getNoneDependency(2, left, all)
	if len(indexs) != 1 {
		t.Errorf("transaction count mismatch: have %d, want %d", len(indexs), 1)
	}
	if len(left) != 2 {
		t.Errorf("left dependency count mismatch: have %d, want %d", len(left), 2)
	}
	indexs = getNoneDependency(3, left, all)
	if len(indexs) != 0 {
		t.Errorf("transaction count mismatch: have %d, want %d", len(indexs), 0)
	}
	indexs = getNoneDependency(4, left, all)
	if len(indexs) != 1 {
		t.Errorf("transaction count mismatch: have %d, want %d", len(indexs), 1)
	}
	indexs = getNoneDependency(5, left, all)
	if len(indexs) != 1 {
		t.Errorf("transaction count mismatch: have %d, want %d", len(indexs), 1)
	}
	if len(left) != 0 {
		t.Errorf("left dependency count mismatch: have %d, want %d", len(left), 0)
	}
}

func makeDag() types.DAG {
	var dag = make(types.DAG, 7)
	dag[0] = nil
	dag[1] = nil
	dag[2] = nil
	var dep3 types.Dependency
	dep3 = append(dep3, 1, 2)
	dag[3] = dep3
	dag[4] = nil
	var dep5 types.Dependency
	dep5 = append(dep5, 2, 4)
	dag[5] = dep5
	var dep6 types.Dependency
	dep6 = append(dep6, 5)
	dag[6] = dep6
	return dag
}

func TestPool(t *testing.T) {
	goRoutinePool, err := ants.NewPool(runtime.NumCPU()*2, ants.WithPreAlloc(true))
	if err != nil {
		return
	}
	go func() {
		err := goRoutinePool.Submit(func() {
			panic("test")
		})
		if err != nil {
			println("test panic")
			println(err.Error())
		}
	}()

	time.Sleep(1 * time.Second)
	println("complete")
}
