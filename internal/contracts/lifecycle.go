package contracts

type Lifecycle byte

const (
	Transient Lifecycle = iota
	Scope
	Singleton
)
