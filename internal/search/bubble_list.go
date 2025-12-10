package search

import (
	"iter"
	"sync"
)

type bubbleNode[Tkey comparable, Tvalue any] struct {
	key      Tkey
	value    Tvalue
	accesses int
}

// BubbleList contains a slice of entries that bubble up based on access frequency.
type BubbleList[Tkey comparable, Tvalue any] struct {
	sync.Mutex
	entries []bubbleNode[Tkey, Tvalue]
}

// Add inserts the key-value pair into this cache.
//
// Due to chronological separation, Add is guaranteed to not interfere with
// any read operations.
//
// Parameters:
//   - key maps the payload value.
//   - value is the payload that is mapped to key.
func (this *BubbleList[Tkey, Tvalue]) Add(key Tkey, value Tvalue) {
	this.entries = append(this.entries, bubbleNode[Tkey, Tvalue]{key: key, value: value})
}

// All iterates through all key-value pairs.
//
// All is only called during Verify, when Find calls don't reorder.
func (this *BubbleList[Tkey, Tvalue]) All() iter.Seq2[Tkey, Tvalue] {
	return func(yield func(Tkey, Tvalue) bool) {
		for _, node := range this.entries {
			if !yield(node.key, node.value) {
				return
			}
		}
	}
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
func (this *BubbleList[Tkey, Tvalue]) Find(key Tkey, reorder ReorderOption) (value Tvalue, found bool) {
	found = false

	defer this.Unlock()
	this.Lock()
	for i, entry := range this.entries {
		if entry.key == key {
			value = entry.value
			found = true
			if bool(reorder) {
				this.reorder(i, &entry)
			}

			break
		}
	}

	return value, found
}

// Prepare is called right before Verify. Any preparation before search
// functions is done here.
func (this *BubbleList[Tkey, Tvalue]) Prepare() {}

func (this *BubbleList[Tkey, Tvalue]) reorder(index int, entry *bubbleNode[Tkey, Tvalue]) {
	entry.accesses++
	if index > 0 && this.entries[index].accesses > this.entries[index-1].accesses {
		this.entries[index-1], this.entries[index] = this.entries[index], this.entries[index-1]
	}
}
