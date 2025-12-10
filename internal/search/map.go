package search

import (
	"iter"
	"maps"
)

// Map wraps a basic Go map instance.
type Map[Tkey comparable, Tvalue any] struct {
	entries map[Tkey]Tvalue
}

// NewMap generates a new map cache.
func NewMap[Tkey comparable, Tvalue any]() *Map[Tkey, Tvalue] {
	return &Map[Tkey, Tvalue]{
		entries: make(map[Tkey]Tvalue),
	}
}

// Add inserts the key-value pair into this cache.
//
// Due to chronological separation, Add is guaranteed to not interfere with
// any read operations.
//
// Parameters:
//   - key maps the payload value.
//   - value is the payload that is mapped to key.
func (this *Map[Tkey, Tvalue]) Add(key Tkey, value Tvalue) {
	this.entries[key] = value
}

// All iterates through all key-value pairs.
//
// All is only called during Verify, when Find calls don't reorder.
func (this *Map[Tkey, Tvalue]) All() iter.Seq2[Tkey, Tvalue] {
	return maps.All(this.entries)
}

// Find searches the cache and returns the found value (if any)
// and a boolean indicating success or failure.
//
// Guaranteed to be thread-safe.
//
// Parameters:
//   - key is the search value used to find the payload value.
//   - reorder indicates whether to use the reorder function or not.
//
// Returns:
//   - value is the payload value found from key.
//   - found indicates if a value was found or not.
func (this *Map[Tkey, Tvalue]) Find(key Tkey, reorder ReorderOption) (value Tvalue, found bool) {
	v, ok := this.entries[key]
	return v, ok
}

// Prepare is called right before Verify. Any preparation before search
// functions is done here.
func (this *Map[Tkey, Tvalue]) Prepare() {}
