package migrate

import (
	"fmt"
	"path/filepath"
	"runtime"
	"sync"

	"github.com/trivigy/migrate/v2/types"
)

// Registry is the container holding registered migrations.
var Registry registry

type registry struct {
	db sync.Map
}

// Filter iterates over all registered cluster migrations.
func (r *registry) Filter(glob string, fn func(element interface{})) types.RangeFunc {
	return func(key, value interface{}) bool {
		absPath, err := filepath.Abs(glob)
		if err != nil {
			panic(err)
		}
		if ok, err := filepath.Match(absPath, key.(string)); ok {
			fn(value)
		} else if err != nil {
			panic(err)
		}
		return true
	}
}

// Collect iterates over all regirstered cluster migrations and adds them to
// the specified migration.
func (r *registry) Collect(glob string, collection interface{}) {
	r.Range(r.Filter(glob, func(element interface{}) {
		switch elm := element.(type) {
		case types.Migration:
			col, ok := collection.(*types.Migrations)
			if !ok {
				panic(fmt.Errorf(
					"collection type %T missmatch element type %T",
					collection, element,
				))
			}

			*col = append(*col, elm)
		case types.Release:
			col, ok := collection.(*types.Releases)
			if !ok {
				panic(fmt.Errorf(
					"collection type %T missmatch element type %T",
					collection, element,
				))
			}

			*col = append(*col, elm)
		default:
			panic(fmt.Errorf("element type %T not supported", elm))
		}
	}))
}

func (r *registry) Load(filename string) (value interface{}, ok bool) {
	absPath, err := filepath.Abs(filename)
	if err != nil {
		panic(err)
	}
	return r.db.Load(absPath)
}

func (r *registry) Range(fn func(key, value interface{}) bool) {
	r.db.Range(fn)
}

func (r *registry) Store(value interface{}) {
	_, filename, _, _ := runtime.Caller(1)
	r.db.Store(filename, value)
}
