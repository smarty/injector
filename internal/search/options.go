package search

// ReorderOption defines whether to reorder during a search or not.
type ReorderOption bool

const (
	// NoReorder prevents the search from reordering elements.
	NoReorder ReorderOption = false

	// Reorder allows the search to reorder elements.
	Reorder ReorderOption = true
)
