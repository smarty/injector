package contracts

type ObjectInfo struct {
	ConstructorType         ConstructorType
	ConstructorValue        ConstructorValue
	Lifecycle               Lifecycle
	Singleton               any
	ConstructorFunction     func(*[]ScopedInstance) (value any, err error)
	ConstructorReturnsError bool
}
