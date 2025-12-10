# Injector

A lightweight, high-performance dependency injection library for Go.

## Overview

`injector` is a compile-time-safe dependency injection container designed for Go applications that need reliable, fast dependency resolution at startup. It provides multiple caching strategies to optimize for different access patterns and supports flexible lifecycle management for registered types.

## Features

- **Type-Safe API**: Generic functions for compile-time type safety
- **Multiple Lifecycle Options**: Singleton, Scope, and Transient lifecycles
- **Error Handling**: Constructors can optionally return an error value
- **Flexible Registration**: Register by type with custom constructors
- **Named Lookups**: Retrieve dependencies by type name
- **Caching Strategies**: Choose between Map, BubbleList, and PriorityList caches
- **Dependency Verification**: Validate your dependency graph before runtime
- **Scoped Instances**: Create isolated scopes for request-specific dependencies
- **Function Injection**: Automatically inject dependencies into functions

## Installation

```bash
go get github.com/smarty/injector
```

## Quick Start

### Basic Usage

```go
package main

import (
	"github.com/smarty/injector"
)

type Database interface {
	Query(sql string) ([]string, error)
}

type MySQLDB struct{}

func (d *MySQLDB) Query(sql string) ([]string, error) {
	// Implementation
	return []string{}, nil
}

func main() {
	di := injector.New()

	// Register a singleton instance
	injector.RegisterSingleton[Database](di, func() Database {
		return &MySQLDB{}
	})

	// Verify the dependency graph
	if err := injector.Verify(di); err != nil {
		panic(err)
	}

	// Retrieve the database
	db, err := injector.Get[Database](di)
	if err != nil {
		panic(err)
	}

	// Use the database
	_ = db
}
```

### Lifecycle Options

#### Singleton
Instantiated once and reused across the entire application lifetime.

```go
injector.RegisterSingleton[MyType](di, constructor)
```

#### Scope
Instantiated once per `Get()` call, allowing for request-scoped or operation-scoped instances.

```go
injector.RegisterScope[MyType](di, constructor)
```

#### Transient
Instantiated every time it's requested as a dependency.

```go
injector.RegisterTransient[MyType](di, constructor)
```

### Error Handling in Constructors

Constructors can return an error in addition to the instance:

```go
injector.RegisterSingletonError[MyType](di, func() (MyType, error) {
	// Return instance and error
	return MyType{}, nil
})
```

### Function Injection

Automatically inject dependencies into functions:

```go
func setupDatabase(db Database) error {
	// Setup logic
	return nil
}

// Call the function with injected dependencies
err := injector.Call(di, setupDatabase)
```

### Named Lookups

Register types that can be retrieved by name:

```go
db, err := injector.GetByName(di, "Database")
```

### Caching Strategies

Choose the caching strategy that best fits your access patterns:

```go
// Map: Good for random access patterns (default)
di := injector.New(injector.Map)

// BubbleList: Good for stable access patterns
di := injector.New(injector.BubbleList)

// PriorityList: Good for stable but changing access patterns
di := injector.New(injector.PriorityList)
```

## API Reference

### Core Methods

- **`New(cacheStrategy ...CacheStrategy) *Injector`**: Create a new injector instance
- **`Get[T](di *Injector) (T, error)`**: Retrieve a dependency by type
- **`GetByName(di *Injector, name string) (any, error)`**: Retrieve a dependency by name
- **`Call(di *Injector, function any) error`**: Call a function with injected dependencies
- **`Call1` through `Call4`**: Call functions with specific return value counts
- **`CallN(di *Injector, function any) ([]any, error)`**: Call a function with any number of returns
- **`Verify(di *Injector) error`**: Validate the dependency graph

### Registration Methods

- **`RegisterSingleton[T](di *Injector, constructor any) error`**: Register a singleton
- **`RegisterSingletonError[T](di *Injector, constructor any) error`**: Register a singleton with error handling
- **`RegisterScope[T](di *Injector, constructor any) error`**: Register a scoped instance
- **`RegisterScopeError[T](di *Injector, constructor any) error`**: Register a scoped instance with error handling
- **`RegisterTransient[T](di *Injector, constructor any) error`**: Register a transient instance
- **`RegisterTransientError[T](di *Injector, constructor any) error`**: Register a transient with error handling

## Error Handling

The injector provides specific error types for different failure scenarios:

- `ErrorAlreadyRegistered`: A type has already been registered
- `ErrorBadState`: Injector is in an invalid state for the requested operation
- `ErrorDependencyLoop`: A circular dependency has been detected
- `ErrorNotRegistered`: A required dependency has not been registered
- `ErrorNotStructOrInterface`: A type is not suitable for registration
- `ErrorVariadicArguments`: A function has a variadic signature

## Performance Considerations

- The injector is optimized for **startup-time usage**. Generating dependencies takes a few microseconds per call.
- Choose a caching strategy based on your access patterns:
  - **Map**: Random access, no reordering overhead
  - **BubbleList**: Stable patterns, benefits from reordering on stable workloads
  - **PriorityList**: Changing patterns that settle over time

## Testing

The library includes comprehensive tests covering:
- Type registration and validation
- Dependency resolution
- Circular dependency detection
- Error handling
- Performance benchmarks

Run tests with:
```bash
go test ./...
```

Run benchmarks with:
```bash
go test -bench=. ./...
```

## Roadmap

### Planned Features

- **Struct Tag Support for Automatic Field Injection** - Allow dependency injection directly into struct fields via tags (e.g., `di:"inject"`).

- **Factory Pattern Support** - Register factory functions that can take parameters to create multiple instances with different configurations. Useful for creating variants of the same type based on input parameters.

## License

MIT License - See [LICENSE](LICENSE) for details
