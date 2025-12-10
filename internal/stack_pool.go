package internal

import (
	"sync"

	"github.com/smarty/injector/internal/contracts"
)

// StackPool is used for pooling the scoped stacks.
type StackPool struct {
	mutex  sync.Mutex
	stacks [][]contracts.ScopedInstance
}

// CheckIn returns the scoped stack back to this pool.
func (this *StackPool) CheckIn(value []contracts.ScopedInstance) {
	this.mutex.Lock()
	defer this.mutex.Unlock()

	this.stacks = append(this.stacks, value)
}

// CheckOut will find or generate a new scoped stack and return it.
func (this *StackPool) CheckOut() []contracts.ScopedInstance {
	const startSize = 8

	this.mutex.Lock()
	defer this.mutex.Unlock()

	if len(this.stacks) == 0 {
		return make([]contracts.ScopedInstance, 0, startSize)
	}

	value := this.stacks[len(this.stacks)-1]
	this.stacks = this.stacks[:len(this.stacks)-1]
	return value
}
