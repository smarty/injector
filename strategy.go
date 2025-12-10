package injector

import (
	"github.com/smarty/injector/internal/contracts"
	"github.com/smarty/injector/internal/search"
)

// CachingStrategy is any one of a few predefined type-caching backends for
// the injector. Each caching strategy has its strengths and weaknesses.
type CacheStrategy int

const (
	// Map uses a Go map for the cache. Good for truly random access.
	Map CacheStrategy = iota

	// BubbleList uses a slice that reorders based on the number of times
	// each element is accessed. Good for highly stable access patterns that
	// don't change during runtime.
	BubbleList

	// PriorityList uses a linked-list that promotes an item all the way to
	// the front whenever it is accessed. Good for fairly stable access
	// patterns that can change over time.
	PriorityList
)

func generateCache(strategy CacheStrategy) search.Cache[contracts.KeyType, *contracts.ObjectInfo] {
	switch strategy {
	case Map:
		return search.NewMap[contracts.KeyType, *contracts.ObjectInfo]()
	case BubbleList:
		return new(search.BubbleList[contracts.KeyType, *contracts.ObjectInfo])
	case PriorityList:
		return new(search.PriorityList[contracts.KeyType, *contracts.ObjectInfo])
	default:
		return nil
	}
}
