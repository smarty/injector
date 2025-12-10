package search

import "iter"

// Cache is a very basic cache implementation allowing inserts, finds,
// and full iterations.
type Cache[Tkey comparable, Tvalue any] interface {
	// Add inserts the key-value pair into this cache.
	//
	// Due to chronological separation, Add is guaranteed to not interfere with
	// any read operations.
	//
	// Parameters:
	//   - key maps the payload value.
	//   - value is the payload that is mapped to key.
	Add(key Tkey, value Tvalue)

	// All iterates through all key-value pairs.
	//
	// All is only called during Verify, when Find calls don't reorder.
	All() iter.Seq2[Tkey, Tvalue]

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
	Find(key Tkey, reorder ReorderOption) (value Tvalue, found bool)

	// Prepare is called right before Verify. Any preparation before search
	// functions is done here.
	Prepare()
}
