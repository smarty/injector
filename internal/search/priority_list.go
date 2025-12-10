package search

import (
	"iter"
	"sync"
)

type priorityListNode[Tkey comparable, Tvalue any] struct {
	key   Tkey
	value Tvalue
	next  *priorityListNode[Tkey, Tvalue]
}

// PriorityList is a linked list of items that jump to the front when accessed.
type PriorityList[Tkey comparable, Tvalue any] struct {
	sync.Mutex
	head *priorityListNode[Tkey, Tvalue]
}

// Add inserts the key-value pair into this cache.
//
// Due to chronological separation, Add is guaranteed to not interfere with
// any read operations.
//
// Parameters:
//   - key maps the payload value.
//   - value is the payload that is mapped to key.
func (this *PriorityList[Tkey, Tvalue]) Add(key Tkey, value Tvalue) {
	node := &priorityListNode[Tkey, Tvalue]{
		key:   key,
		value: value,
		next:  this.head,
	}

	this.head = node
}

// All iterates through all key-value pairs.
//
// All is only called during Verify, when Find calls don't reorder.
func (this *PriorityList[Tkey, Tvalue]) All() iter.Seq2[Tkey, Tvalue] {
	return func(yield func(Tkey, Tvalue) bool) {
		defer this.Unlock()
		this.Lock()
		current := this.head
		for current != nil {
			if !yield(current.key, current.value) {
				return
			}

			current = current.next
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
func (this *PriorityList[Tkey, Tvalue]) Find(key Tkey, reorder ReorderOption) (value Tvalue, found bool) {
	found = false
	var previous *priorityListNode[Tkey, Tvalue] = nil

	defer this.Unlock()
	this.Lock()
	current := this.head
	for current != nil {
		if current.key == key {
			if bool(reorder) {
				this.reorder(previous, current)
			}

			value = current.value
			found = true
			break
		}

		previous = current
		current = current.next
	}

	return value, found
}

// Prepare is called right before Verify. Any preparation before search
// functions is done here.
func (this *PriorityList[Tkey, Tvalue]) Prepare() {}

func (this *PriorityList[Tkey, Tvalue]) reorder(previous *priorityListNode[Tkey, Tvalue], current *priorityListNode[Tkey, Tvalue]) {
	if previous != nil {
		previous.next = current.next
		current.next = this.head
		this.head = current
	}
}
