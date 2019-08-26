package types

import (
	"fmt"
	"path/filepath"
	"runtime"
	"strings"
	"sync"
)

// Registry holds stored changes of either database migrations or kubernetes
// releases.
type Registry struct {
	db sync.Map
}

// Filter iterates over all registered cluster migrations.
func (r *Registry) Filter(glob string, fn func(element interface{})) RangeFunc {
	return func(key, value interface{}) bool {

		_, filename, _, _ := runtime.Caller(0)
		path := strings.Split(filename, "/")
		root := "/" + filepath.Join(path[:len(path)-3]...)

		fmt.Printf("%+v\n", root)

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
func (r *Registry) Collect(glob string, collection interface{}) {
	r.Range(r.Filter(glob, func(element interface{}) {
		switch elm := element.(type) {
		case *Migration:
			col, ok := collection.(*Migrations)
			if !ok {
				panic(fmt.Errorf(
					"collection type %T missmatch element type %T",
					collection, element,
				))
			}

			*col = append(*col, elm)
		case *Release:
			col, ok := collection.(*Releases)
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

// Load returns the change, or nil if no value is present. The ok result
// indicates whether value was found in the map.
func (r *Registry) Load(filename string) (value interface{}, ok bool) {
	absPath, err := filepath.Abs(filename)
	if err != nil {
		panic(err)
	}
	return r.db.Load(absPath)
}

// Range calls fn sequentially for each stored registry. If fn returns false,
// range stops the iteration. `key` represents the path where the change has
// been registored.
func (r *Registry) Range(fn func(key, value interface{}) bool) {
	r.db.Range(fn)
}

// Store stores a change inside of the registry.
func (r *Registry) Store(value interface{}) {
	_, filename, _, _ := runtime.Caller(1)
	r.db.Store(filename, value)
}
