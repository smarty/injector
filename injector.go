package injector

import (
	"errors"
	"fmt"
	"reflect"
	"strings"

	"github.com/smarty/injector/internal"
	"github.com/smarty/injector/internal/contracts"
	"github.com/smarty/injector/internal/search"
	"github.com/smarty/tries"
)

// Injector is a dependency-injector meant to be single-use at startup and
// used in applications where taking a few microseconds generating a dependency
// is acceptable.
type Injector struct {
	library           search.Cache[contracts.KeyType, *contracts.ObjectInfo]
	nameToKeyTrie     tries.Trie[string, reflect.Type]
	scopePool         internal.StackPool
	verificationError error
	verified          bool
}

// New creates a new injector, preloaded with itself.
//
// Parameters:
//   - cacheStrategy defines an optional internal caching strategy. If no
//     caching strategy is chosen, the injector defaults to using Map.
//     If more than one strategy is selected, the first strategy will be used.
//
// Returns:
//   - Injector with self already registered as a singleton.
func New(cacheStrategy ...CacheStrategy) *Injector {
	// default to Map
	strategy := Map
	if len(cacheStrategy) > 0 {
		strategy = cacheStrategy[0]
	}

	nameToKeyTrie, _ := tries.NewTrie[string, reflect.Type](func(in byte) (out byte, use bool) {
		if (in >= 'A' && in <= 'Z') || (in >= 'a' && in <= 'z') || (in >= '0' && in <= '9') { // only alpha-numerics are considered
			return in, true
		}

		return 0, false
	})

	di := &Injector{
		library:       generateCache(strategy),
		nameToKeyTrie: nameToKeyTrie,
		verified:      false,
	}

	RegisterSingleton[*Injector](di, func() *Injector { return di })
	return di
}

// Call checks a function's signature then calls the function by injecting all
// the arguments. Call is used for any function that has no return values.
//
// Parameters:
//   - function is the function to be called with injected arguments.
//
// Returns:
//   - err returns any error encountered during the call.
//
// Error:
//   - if calling Get on any of the argument types would error.
//   - if the function provided is not a function.
//   - if the function provided is variadic.
//   - if the function provided has an incongruent number of return values.
func (this *Injector) Call(function any) (err error) {
	_, err = this.callN(function, 0)
	return err
}

// Call1 checks a function's signature then calls the function by injecting all
// the arguments. Call1 is used for any function that has exactly one return
// value.
//
// Parameters:
//   - function is the function to be called with injected arguments.
//
// Returns:
//   - r1 is return value 1.
//   - err returns any error encountered during the call.
//
// Errors:
//   - if calling Get on any of the argument types would error.
//   - if the function provided is not a function.
//   - if the function provided is variadic.
//   - if the function provided has an incongruent number of return values.
func (this *Injector) Call1(function any) (r1 any, err error) {
	var returns []any
	returns, err = this.callN(function, 1)
	return returns[0], err
}

// Call2 checks a function's signature then calls the function by injecting all
// the arguments. Call2 is used for any function that has exactly two return
// values.
//
// Parameters:
//   - function is the function to be called with injected arguments.
//
// Returns:
//   - r1 is return value 1.
//   - r2 is return value 2.
//   - err returns any error encountered during the call.
//
// Errors:
//   - if calling Get on any of the argument types would error.
//   - if the function provided is not a function.
//   - if the function provided is variadic.
//   - if the function provided has an incongruent number of return values.
func (this *Injector) Call2(function any) (r1, r2 any, err error) {
	var returns []any
	returns, err = this.callN(function, 2)
	return returns[0], returns[1], err
}

// Call3 checks a function's signature then calls the function by injecting all
// the arguments. Call3 is used for any function that has exactly three return
// values.
//
// Parameters:
//   - function is the function to be called with injected arguments.
//
// Returns:
//   - r1 is return value 1.
//   - r2 is return value 2.
//   - r3 is return value 3.
//   - err returns any error encountered during the call.
//
// Returns:
//   - if calling Get on any of the argument types would error.
//   - if the function provided is not a function.
//   - if the function provided is variadic.
//   - if the function provided has an incongruent number of return values.
func (this *Injector) Call3(function any) (r1, r2, r3 any, err error) {
	var returns []any
	returns, err = this.callN(function, 3)
	return returns[0], returns[1], returns[2], err
}

// Call4 checks a function's signature then calls the function by injecting all
// the arguments. Call4 is used for any function that has exactly four return
// values.
//
// Parameters:
//   - function is the function to be called with injected arguments.
//
// Returns:
//   - r1 is return value 1.
//   - r2 is return value 2.
//   - r3 is return value 3.
//   - r4 is return value 4.
//   - err returns any error encountered during the call.
//
// Errors:
//   - if calling Get on any of the argument types would error.
//   - if the function provided is not a function.
//   - if the function provided is variadic.
//   - if the function provided has an incongruent number of return values.
func (this *Injector) Call4(function any) (r1, r2, r3, r4 any, err error) {
	var returns []any
	returns, err = this.callN(function, 4)
	return returns[0], returns[1], returns[2], returns[3], err
}

// CallN checks a function's signature then calls the function by injecting all
// the arguments. CallN is used for any function with any number of return
// values.
//
// Parameters:
//   - function is the function to be called with injected arguments.
//
// Returns:
//   - returns contains all return values.
//   - err returns any error encountered during the call.
//
// Errors:
//   - if calling Get on any of the argument types would error.
//   - if the function provided is not a function.
//   - if the function provided is variadic.
func (this *Injector) CallN(function any) (returns []any, err error) {
	return this.callN(function, reflect.TypeOf(function).NumOut())
}

// Get retrieves the given type using the registered constructor or instance.
//
// Parameters:
//   - key is the type to look for a registered instance or constructor for.
//
// Returns:
//   - The registered instance or the result of the registered constructor.
//   - err is nil unless an error occurred during retrieval.
//
// Errors:
//   - if Verify() has not been called.
//   - if Verify() returned an error.
func (this *Injector) Get(key reflect.Type) (value any, err error) {
	err = assertValidState(this)
	if err != nil {
		return nil, err
	}

	scopedStack := this.scopePool.CheckOut()
	defer this.scopePool.CheckIn(scopedStack)

	var objAsAny any
	objAsAny, err = get(this, key, &scopedStack)
	if err != nil {
		return nil, err
	}

	switch o := objAsAny.(type) {
	case reflect.Value:
		return o.Interface(), nil
	default:
		return objAsAny, nil
	}
}

// GetByName retrieves the named type using the registered constructor or
// instance. The name is expected to be truncated down to type name only,
// package should not be included in the name. Pointer "*" symbols are also
// assumed to be stripped from the name and should not be included.
//
// Parameters:
//   - name is the name of the type that should have been registered. Expected
//     to be truncated down to type name only, package should not be included.
//     Pointer "*" symbols are also assumed to be stripped from the name and
//     should not be included.
//
// Returns:
//   - value is the registered instance or the result of the registered
//     constructor.
//   - err is nil unless the named type cannot be found.
//
// Errors:
//   - ErrorNotRegistered is returned if the named type cannot be found.
//
// Errors:
//   - if Verify() has not been called.
//   - if Verify() returned an error.
func (this *Injector) GetByName(name string) (value any, err error) {
	key, found := this.nameToKeyTrie.Find(name)
	if !found {
		return nil, fmt.Errorf(
			"%w: no keys that match the string pattern %q have been registered",
			ErrorNotRegistered,
			name,
		)
	}

	return this.Get(key)
}

// RegisterScope adds a constructor for the given type.
// Every time the type is requested in a unique Get() call, the same instance is
// always returned. If it's requested again in a new Get() call, the constructor
// is called and a new instance is returned.
//
// Notes:
//   - Constructor is expected to be a function that returns exactly one value.
//     if the constructor also returns an error, use
//     [Injector.RegisterScopeError].
//
// Parameters:
//   - key is the registered type that the constructor will be registered with.
//   - constructor is the requisite function to generate the type.
//
// Errors:
//   - ErrorAlreadyRegistered is returned when a type has already been
//     registered.
//   - ErrorNoReturns is returned when a constructor has no return value.
//   - ErrorNotAFunction is returned when a non-function is passed as a
//     constructor.
//   - ErrorNotAssignable is returned when a constructor returns a type that
//     cannot be assigned to the key type.
//   - ErrorNotStructOrInterface is returned when a type is not registerable.
//   - ErrorTooManyReturns is returned when a constructor has more than 1
//     return value.
//   - ErrorVariadicArguments is returned when a constructor has a variadic
//     signature.
func (this *Injector) RegisterScope(key reflect.Type, constructor any) error {
	info := &contracts.ObjectInfo{
		ConstructorType:  contracts.ConstructorType(reflect.TypeOf(constructor)),
		ConstructorValue: contracts.ConstructorValue(reflect.ValueOf(constructor)),
		Lifecycle:        contracts.Scope,
	}

	return register(this, key, info)
}

// RegisterScopeError adds a constructor for the given type.
// Every time the type is requested in a unique Get() call, the same instance is
// always returned. If it's requested again in a new Get() call, the constructor
// is called and a new instance is returned.
//
// Notes:
//   - Constructor is expected to return (Tkey, error). If your constructor
//     does not return an error, use [Injector.RegisterScope] instead.
//
// Parameters:
//   - key is the registered type that the constructor will be registered with.
//   - constructor is the requisite function to generate the type.
//
// Errors:
//   - ErrorAlreadyRegistered is returned when a type has already been
//     registered.
//   - ErrorNoReturns is returned when a constructor has no return value.
//   - ErrorNotAFunction is returned when a non-function is passed as a
//     constructor.
//   - ErrorNotAssignable is returned when a constructor returns a type that
//     cannot be assigned to the key type.
//   - ErrorNotStructOrInterface is returned when a type is not registerable.
//   - ErrorTooManyReturns is returned when a constructor has more than 1
//     return value.
//   - ErrorVariadicArguments is returned when a constructor has a variadic
//     signature.
func (this *Injector) RegisterScopeError(key reflect.Type, constructor any) error {
	info := &contracts.ObjectInfo{
		ConstructorType:         contracts.ConstructorType(reflect.TypeOf(constructor)),
		ConstructorValue:        contracts.ConstructorValue(reflect.ValueOf(constructor)),
		Lifecycle:               contracts.Scope,
		ConstructorReturnsError: true,
	}

	return register(this, key, info)
}

// RegisterSingleton adds a constructor for the given type.
// Every time the type is requested, the same instance is always returned.
//
// Notes:
//   - Constructor is expected to be a function that returns exactly one value.
//     if the constructor also returns an error, use
//     [Injector.RegisterSingletonError]
//
// Parameters:
//   - key is the registered type that the constructor will be registered with.
//   - constructor is the requisite function to generate the type.
//
// Errors:
//   - ErrorAlreadyRegistered is returned when a type has already been
//     registered.
//   - ErrorNoReturns is returned when a constructor has no return value.
//   - ErrorNotAFunction is returned when a non-function is passed as a
//     constructor.
//   - ErrorNotAssignable is returned when a constructor returns a type that
//     cannot be assigned to the key type.
//   - ErrorNotStructOrInterface is returned when a type is not registerable.
//   - ErrorTooManyReturns is returned when a constructor has more than 1
//     return value.
//   - ErrorVariadicArguments is returned when a constructor has a variadic
//     signature.
func (this *Injector) RegisterSingleton(key reflect.Type, constructor any) error {
	info := &contracts.ObjectInfo{
		ConstructorType:  contracts.ConstructorType(reflect.TypeOf(constructor)),
		ConstructorValue: contracts.ConstructorValue(reflect.ValueOf(constructor)),
		Lifecycle:        contracts.Singleton,
	}

	return register(this, key, info)
}

// RegisterSingletonError adds a constructor for the given type.
// Every time the type is requested, the same instance is always returned.
//
// Notes:
//   - Constructor is expected to return (Tkey, error). If your constructor
//     does not return an error, use [Injector.RegisterSingleton] instead.
//
// Parameters:
//   - key is the registered type that the constructor will be registered with.
//   - constructor is the requisite function to generate the type.
//
// Errors:
//   - ErrorAlreadyRegistered is returned when a type has already been
//     registered.
//   - ErrorNoReturns is returned when a constructor has no return value.
//   - ErrorNotAFunction is returned when a non-function is passed as a
//     constructor.
//   - ErrorNotAssignable is returned when a constructor returns a type that
//     cannot be assigned to the key type.
//   - ErrorNotStructOrInterface is returned when a type is not registerable.
//   - ErrorTooManyReturns is returned when a constructor has more than 1
//     return value.
//   - ErrorVariadicArguments is returned when a constructor has a variadic
//     signature.
func (this *Injector) RegisterSingletonError(key reflect.Type, constructor any) error {
	info := &contracts.ObjectInfo{
		ConstructorType:         contracts.ConstructorType(reflect.TypeOf(constructor)),
		ConstructorValue:        contracts.ConstructorValue(reflect.ValueOf(constructor)),
		Lifecycle:               contracts.Singleton,
		ConstructorReturnsError: true,
	}

	return register(this, key, info)
}

// RegisterTransient adds a constructor for the given type.
// Every time the type is requested, the constructor is always called and a new
// instance is returned.
//
// Notes:
//   - Constructor is expected to be a function that returns exactly one value.
//     if the constructor also returns an error, use
//     [Injector.RegisterTransientError].
//
// Parameters:
//   - key is the registered type that the constructor will be registered with.
//   - constructor is the requisite function to generate the type.
//
// Errors:
//   - ErrorAlreadyRegistered is returned when a type has already been
//     registered.
//   - ErrorNoReturns is returned when a constructor has no return value.
//   - ErrorNotAFunction is returned when a non-function is passed as a
//     constructor.
//   - ErrorNotAssignable is returned when a constructor returns a type that
//     cannot be assigned to the key type.
//   - ErrorNotStructOrInterface is returned when a type is not registerable.
//   - ErrorTooManyReturns is returned when a constructor has more than 1
//     return value.
//   - ErrorVariadicArguments is returned when a constructor has a variadic
//     signature.
func (this *Injector) RegisterTransient(key reflect.Type, constructor any) error {
	info := &contracts.ObjectInfo{
		ConstructorType:  contracts.ConstructorType(reflect.TypeOf(constructor)),
		ConstructorValue: contracts.ConstructorValue(reflect.ValueOf(constructor)),
		Lifecycle:        contracts.Transient,
	}

	return register(this, key, info)
}

// RegisterTransientError adds a constructor for the given type.
// Every time the type is requested, the constructor is always called and a new
// instance is returned.
//
// Notes:
//   - Constructor is expected to return (Tkey, error). If your constructor
//     does not return an error, use [Injector.RegisterTransient] instead.
//
// Parameters:
//   - key is the registered type that the constructor will be registered with.
//   - constructor is the requisite function to generate the type.
//
// Errors:
//   - ErrorAlreadyRegistered is returned when a type has already been
//     registered.
//   - ErrorNoReturns is returned when a constructor has no return value.
//   - ErrorNotAFunction is returned when a non-function is passed as a
//     constructor.
//   - ErrorNotAssignable is returned when a constructor returns a type that
//     cannot be assigned to the key type.
//   - ErrorNotStructOrInterface is returned when a type is not registerable.
//   - ErrorTooManyReturns is returned when a constructor has more than 1
//     return value.
//   - ErrorVariadicArguments is returned when a constructor has a variadic
//     signature.
func (this *Injector) RegisterTransientError(key reflect.Type, constructor any) error {
	info := &contracts.ObjectInfo{
		ConstructorType:         contracts.ConstructorType(reflect.TypeOf(constructor)),
		ConstructorValue:        contracts.ConstructorValue(reflect.ValueOf(constructor)),
		Lifecycle:               contracts.Transient,
		ConstructorReturnsError: true,
	}

	return register(this, key, info)
}

// Call checks a function's signature then calls the function by injecting all
// the arguments. Call is used for any function that has no return values.
//
// Parameters:
//   - injector is the dependency injector to use when making the function call.
//   - function is the function to be called with injected arguments.
//
// Returns:
//   - err returns any error encountered during the call.
//
// Errors:
//   - if calling Get on any of the argument types would error.
//   - if the function provided is not a function.
//   - if the function provided is variadic.
//   - if the function provided has an incongruent number of return values.
func Call(injector *Injector, function any) (err error) {
	_, err = injector.callN(function, 0)
	return err
}

// Call1 checks a function's signature then calls the function by injecting all
// the arguments. Call1 is used for any function that has exactly one return
// value.
//
// Parameters:
//   - injector is the dependency injector to use when making the function call.
//   - function is the function to be called with injected arguments.
//
// Returns:
//   - r1 is return value 1.
//   - err returns any error encountered during the call.
//
// Errors:
//   - if calling Get on any of the argument types would error.
//   - if the function provided is not a function.
//   - if the function provided is variadic.
//   - if the function provided has an incongruent number of return values.
func Call1[T1 any](injector *Injector, function any) (r1 T1, err error) {
	var returns []any
	returns, err = injector.callN(function, 1)
	return returns[0].(T1), err
}

// Call2 checks a function's signature then calls the function by injecting all
// the arguments. Call2 is used for any function that has exactly two return
// values.
//
// Parameters:
//   - injector is the dependency injector to use when making the function call.
//   - function is the function to be called with injected arguments.
//
// Returns:
//   - r1 is return value 1.
//   - r2 is return value 2.
//   - err returns any error encountered during the call.
//
// Errors:
//   - if calling Get on any of the argument types would error.
//   - if the function provided is not a function.
//   - if the function provided is variadic.
//   - if the function provided has an incongruent number of return values.
func Call2[T1, T2 any](injector *Injector, function any) (r1 T1, r2 T2, err error) {
	var returns []any
	returns, err = injector.callN(function, 2)
	return returns[0].(T1), returns[1].(T2), err
}

// Call3 checks a function's signature then calls the function by injecting all
// the arguments. Call3 is used for any function that has exactly three return
// values.
//
// Parameters:
//   - injector is the dependency injector to use when making the function call.
//   - function is the function to be called with injected arguments.
//
// Returns:
//   - r1 is return value 1.
//   - r2 is return value 2.
//   - r3 is return value 3.
//   - err returns any error encountered during the call.
//
// Errors:
//   - if calling Get on any of the argument types would error.
//   - if the function provided is not a function.
//   - if the function provided is variadic.
//   - if the function provided has an incongruent number of return values.
func Call3[T1, T2, T3 any](injector *Injector, function any) (r1 T1, r2 T2, r3 T3, err error) {
	var returns []any
	returns, err = injector.callN(function, 3)
	return returns[0].(T1), returns[1].(T2), returns[2].(T3), err
}

// Call4 checks a function's signature then calls the function by injecting all
// the arguments. Call4 is used for any function that has exactly four return
// values.
//
// Parameters:
//   - injector is the dependency injector to use when making the function call.
//   - function is the function to be called with injected arguments.
//
// Returns:
//   - r1 is return value 1.
//   - r2 is return value 2.
//   - r3 is return value 3.
//   - r4 is return value 4.
//   - err returns any error encountered during the call.
//
// Errors:
//   - if calling Get on any of the argument types would error.
//   - if the function provided is not a function.
//   - if the function provided is variadic.
//   - if the function provided has an incongruent number of return values.
func Call4[T1, T2, T3, T4 any](injector *Injector, function any) (r1 T1, r2 T2, r3 T3, r4 T4, err error) {
	var returns []any
	returns, err = injector.callN(function, 4)
	return returns[0].(T1), returns[1].(T2), returns[2].(T3), returns[3].(T4), err
}

// CallN checks a function's signature then calls the function by injecting all
// the arguments. CallN is used for any function with any number of return
// values.
//
// Parameters:
//   - injector is the dependency injector to use when making the function call.
//   - function is the function to be called with injected arguments.
//
// Returns:
//   - returns contains all return values.
//   - err returns any error encountered during the call.
//
// Errors:
//   - if calling Get on any of the argument types would error.
//   - if the function provided is not a function.
//   - if the function provided is variadic.
func CallN(injector *Injector, function any) (returns []any, err error) {
	return injector.callN(function, reflect.TypeOf(function).NumOut())
}

// Get retrieves the given type using the registered constructor or instance.
//
// Parameters:
//   - injector is the dependency injector to get the instance from.
//
// Returns:
//   - value is the registered instance or the result of the registered
//     constructor.
//   - err is nil unless an error occurred during retrieval.
//
// Errors:
//   - if Verify() has not been called.
//   - if Verify() returned an error.
func Get[Tkey any](injector *Injector) (value Tkey, err error) {
	key := reflect.TypeFor[Tkey]()
	var rawValue any
	rawValue, err = injector.Get(key)
	if err != nil {
		return value, err
	}

	return rawValue.(Tkey), nil
}

// GetByName retrieves the named type using the registered constructor or
// instance. The name is expected to be truncated down to type name only,
// package should not be included in the name. Pointer "*" symbols are also
// assumed to be stripped from the name and should not be included.
//
// Parameters:
//   - injector is the dependency injector to get the instance from.
//   - name is the name of the type that should have been registered. Expected
//     to be truncated down to type name only, package should not be included.
//     Pointer "*" symbols are also assumed to be stripped from the name and
//     should not be included.
//
// Returns:
//   - value is the registered instance or the result of the registered
//     constructor.
//   - err is nil unless the named type cannot be found.
//
// Errors:
//   - ErrorNotRegistered is returned if the named type cannot be found.
//
// Errors:
//   - if Verify() has not been called.
//   - if Verify() returned an error.
func GetByName(injector *Injector, name string) (value any, err error) {
	return injector.GetByName(name)
}

// RegisterScope adds a constructor for the given type.
// Every time the type is requested in a unique Get() call, the same instance is
// always returned. If it's requested again in a new Get() call, the constructor
// is called and a new instance is returned.
//
// Notes:
//   - Constructor is expected to be a function that returns exactly one value.
//     if the constructor also returns an error, use RegisterScopeError instead.
//
// Parameters:
//   - target is the Injector to register the type in.
//   - constructor is the requisite function to generate the type.
//
// Errors:
//   - ErrorAlreadyRegistered is returned when a type has already been
//     registered.
//   - ErrorNoReturns is returned when a constructor has no return value.
//   - ErrorNotAFunction is returned when a non-function is passed as a
//     constructor.
//   - ErrorNotAssignable is returned when a constructor returns a type that
//     cannot be assigned to the key type.
//   - ErrorNotStructOrInterface is returned when a type is not registerable.
//   - ErrorTooManyReturns is returned when a constructor has more than 1
//     return value.
//   - ErrorVariadicArguments is returned when a constructor has a variadic
//     signature.
func RegisterScope[Tkey any](target *Injector, constructor any) error {
	return target.RegisterScope(reflect.TypeFor[Tkey](), constructor)
}

// RegisterScopeError adds a constructor for the given type.
// Every time the type is requested in a unique Get() call, the same instance is
// always returned. If it's requested again in a new Get() call, the constructor
// is called and a new instance is returned.
//
// Notes:
//   - Constructor is expected to return (Tkey, error). If your constructor
//     does not return an error, use RegisterScope instead.
//
// Parameters:
//   - target is the Injector to register the type in.
//   - constructor is the requisite function to generate the type.
//
// Errors:
//   - ErrorAlreadyRegistered is returned when a type has already been
//     registered.
//   - ErrorNoReturns is returned when a constructor has no return value.
//   - ErrorNotAFunction is returned when a non-function is passed as a
//     constructor.
//   - ErrorNotAssignable is returned when a constructor returns a type that
//     cannot be assigned to the key type.
//   - ErrorNotStructOrInterface is returned when a type is not registerable.
//   - ErrorTooManyReturns is returned when a constructor has more than 1
//     return value.
//   - ErrorVariadicArguments is returned when a constructor has a variadic
//     signature.
func RegisterScopeError[Tkey any](target *Injector, constructor any) error {
	return target.RegisterScopeError(reflect.TypeFor[Tkey](), constructor)
}

// RegisterSingleton adds a constructor for the given type.
// Every time the type is requested, the same instance is always returned.
//
// Notes:
//   - Constructor is expected to be a function that returns exactly one value.
//     if the constructor also returns an error, use RegisterSingletonError
//
// Parameters:
//   - target is the Injector to register the type in.
//   - constructor is the requisite function to generate the type.
//
// Errors:
//   - ErrorAlreadyRegistered is returned when a type has already been
//     registered.
//   - ErrorNoReturns is returned when a constructor has no return value.
//   - ErrorNotAFunction is returned when a non-function is passed as a
//     constructor.
//   - ErrorNotAssignable is returned when a constructor returns a type that
//     cannot be assigned to the key type.
//   - ErrorNotStructOrInterface is returned when a type is not registerable.
//   - ErrorTooManyReturns is returned when a constructor has more than 1
//     return value.
//   - ErrorVariadicArguments is returned when a constructor has a variadic
//     signature.
func RegisterSingleton[Tkey any](target *Injector, constructor any) error {
	return target.RegisterSingleton(reflect.TypeFor[Tkey](), constructor)
}

// RegisterSingletonError adds a constructor for the given type.
// Every time the type is requested, the same instance is always returned.
//
// Notes:
//   - Constructor is expected to return (Tkey, error). If your constructor
//     does not return an error, use RegisterSingleton instead.
//
// Parameters:
//   - target is the Injector to register the type in.
//   - constructor is the requisite function to generate the type.
//
// Errors:
//   - ErrorAlreadyRegistered is returned when a type has already been
//     registered.
//   - ErrorNoReturns is returned when a constructor has no return value.
//   - ErrorNotAFunction is returned when a non-function is passed as a
//     constructor.
//   - ErrorNotAssignable is returned when a constructor returns a type that
//     cannot be assigned to the key type.
//   - ErrorNotStructOrInterface is returned when a type is not registerable.
//   - ErrorTooManyReturns is returned when a constructor has more than 1
//     return value.
//   - ErrorVariadicArguments is returned when a constructor has a variadic
//     signature.
func RegisterSingletonError[Tkey any](target *Injector, constructor any) error {
	return target.RegisterSingletonError(reflect.TypeFor[Tkey](), constructor)
}

// RegisterTransient adds a constructor for the given type.
// Every time the type is requested, the constructor is always called and a new
// instance is returned.
//
// Notes:
//   - Constructor is expected to be a function that returns exactly one value.
//     if the constructor also returns an error, use RegisterTransientError
//
// Parameters:
//   - target is the Injector to register the type in.
//   - constructor is the requisite function to generate the type.
//
// Errors:
//   - ErrorAlreadyRegistered is returned when a type has already been
//     registered.
//   - ErrorNoReturns is returned when a constructor has no return value.
//   - ErrorNotAFunction is returned when a non-function is passed as a
//     constructor.
//   - ErrorNotAssignable is returned when a constructor returns a type that
//     cannot be assigned to the key type.
//   - ErrorNotStructOrInterface is returned when a type is not registerable.
//   - ErrorTooManyReturns is returned when a constructor has more than 1
//     return value.
//   - ErrorVariadicArguments is returned when a constructor has a variadic
//     signature.
func RegisterTransient[Tkey any](target *Injector, constructor any) error {
	return target.RegisterTransient(reflect.TypeFor[Tkey](), constructor)
}

// RegisterTransientError adds a constructor for the given type.
// Every time the type is requested, the constructor is always called and a new
// instance is returned.
//
// Notes:
//   - Constructor is expected to return (Tkey, error). If your constructor
//     does not return an error, use RegisterTransient instead.
//
// Parameters:
//   - target is the Injector to register the type in.
//   - constructor is the requisite function to generate the type.
//
// Errors:
//   - ErrorAlreadyRegistered is returned when a type has already been
//     registered.
//   - ErrorNoReturns is returned when a constructor has no return value.
//   - ErrorNotAFunction is returned when a non-function is passed as a
//     constructor.
//   - ErrorNotAssignable is returned when a constructor returns a type that
//     cannot be assigned to the key type.
//   - ErrorNotStructOrInterface is returned when a type is not registerable.
//   - ErrorTooManyReturns is returned when a constructor has more than 1
//     return value.
//   - ErrorVariadicArguments is returned when a constructor has a variadic
//     signature.
func RegisterTransientError[Tkey any](target *Injector, constructor any) error {
	return target.RegisterTransientError(reflect.TypeFor[Tkey](), constructor)
}

// Verify examines all registered types and their corresponding constructors
// and validates them, otherwise an error is returned.
//
// Parameters:
//   - target is the Injector to explore and verify all the registered types in.
//
// Errors:
//   - ErrorDependencyLoop indicates that an unsolvable dependency injection
//     loop.
//   - ErrorNotRegistered indicates that a required dependency does not appear
//     in the registered list.
func Verify(injector *Injector) error {
	injector.verified = false
	injector.verificationError = nil
	injector.library.Prepare()
	for key := range injector.library.All() {
		if err := verify(injector, key); err != nil {
			injector.verificationError = err
			return err
		}
	}

	injector.verified = true
	return nil
}

func (this *Injector) callN(function any, expectedReturnCount int) (returns []any, err error) {
	functionType := reflect.TypeOf(function)
	functionValue := reflect.ValueOf(function)
	if functionType.Kind() != reflect.Func {
		return nil, fmt.Errorf(
			"%w: for value type with name '%s'",
			ErrorNotAFunction,
			functionType.Name())
	}

	if functionType.NumOut() != expectedReturnCount {
		return nil, fmt.Errorf(
			"%w: expected passed function to have [%d] return values, but it has [%d] return values",
			ErrorWrongNumberOfReturns,
			expectedReturnCount,
			functionType.NumOut())
	}

	if functionType.IsVariadic() {
		return nil, ErrorVariadicArguments
	}

	parameterCount := functionType.NumIn()
	values := make([]reflect.Value, parameterCount)
	parametersInfo := make([]contracts.ConstructorType, parameterCount)
	for iParameter := 0; iParameter < parameterCount; iParameter++ {
		parametersInfo[iParameter] = functionType.In(iParameter)
	}

	scopedStack := this.scopePool.CheckOut()
	defer this.scopePool.CheckIn(scopedStack)

	returnValues := func(scopedList *[]contracts.ScopedInstance) []reflect.Value {
		for iParameter := 0; iParameter < parameterCount; iParameter++ {
			rawValue, e := get(this, parametersInfo[iParameter], scopedList)
			if e != nil {
				err = errors.Join(err, e)
			}

			values[iParameter] = rawValue.(reflect.Value)
		}

		return reflect.Value(functionValue).Call(values)
	}(&scopedStack)

	if err != nil {
		return nil, err
	}

	toReturn := make([]any, len(returnValues))
	for iReturn := range returnValues {
		toReturn[iReturn] = returnValues[iReturn].Interface()
	}

	return toReturn, nil
}

func assertValidState(injector *Injector) (err error) {
	if !injector.verified {
		if injector.verificationError != nil {
			return fmt.Errorf(
				"%w: injector is in a bad state with verification error: %w",
				ErrorBadState,
				injector.verificationError)
		}

		return fmt.Errorf(
			"%w: injector is not in a verified state, call Verify() on injector after registering all types",
			ErrorBadState)
	}

	return nil
}

func get(injector *Injector, key contracts.KeyType, scoped *[]contracts.ScopedInstance) (returnValue any, err error) {
	info, found := injector.library.Find(key, search.Reorder)
	if !found {
		return nil, fmt.Errorf("%w: type '%s'", ErrorNotRegistered, key.Name())
	}

	switch info.Lifecycle {
	case contracts.Scope:
		for _, scopedItem := range *scoped {
			if scopedItem.Type == key {
				return scopedItem.Value, nil
			}
		}

		if info.ConstructorFunction != nil {
			obj, e := info.ConstructorFunction(scoped)
			if e != nil {
				return nil, e
			}

			*scoped = append(*scoped, contracts.ScopedInstance{Type: key, Value: obj})
			return obj, nil
		}
	case contracts.Singleton:
		if info.Singleton != nil {
			return info.Singleton, nil
		}
	case contracts.Transient:
		if info.ConstructorFunction != nil {
			return info.ConstructorFunction(scoped)
		}
	}

	parameterCount := info.ConstructorType.NumIn()
	values := make([]reflect.Value, parameterCount)
	parametersInfo := make([]contracts.ConstructorType, parameterCount)
	for iParameter := 0; iParameter < parameterCount; iParameter++ {
		parametersInfo[iParameter] = info.ConstructorType.In(iParameter)
	}

	info.ConstructorFunction = func(scopedList *[]contracts.ScopedInstance) (value any, err error) {
		for iParameter := 0; iParameter < parameterCount; iParameter++ {
			var rawValue any
			rawValue, err = get(injector, parametersInfo[iParameter], scopedList)
			if err != nil {
				return nil, err
			}

			values[iParameter] = rawValue.(reflect.Value)
		}

		returns := reflect.Value(info.ConstructorValue).Call(values)
		if info.ConstructorReturnsError {
			errorRaw := returns[1].Interface()
			if errorRaw != nil {
				return returns[0], errorRaw.(error)
			}

			return returns[0], nil
		}

		return returns[0], nil
	}

	obj, e := info.ConstructorFunction(scoped)
	if e != nil {
		return nil, e
	}

	switch info.Lifecycle {
	case contracts.Scope:
		*scoped = append(*scoped, contracts.ScopedInstance{Type: key, Value: obj})
	case contracts.Singleton:
		info.Singleton = obj
	}

	return obj, nil
}

func isStructLike(key contracts.KeyType) bool {
	return key.Kind() == reflect.Struct || key.Kind() == reflect.Interface
}

func register(target *Injector, key reflect.Type, info *contracts.ObjectInfo) error {
	target.verified = false
	if !isStructLike(key) && !validPointerKey(key) {
		return fmt.Errorf(
			"%w: type '%s'",
			ErrorNotStructOrInterface,
			key.Name())
	}

	if info.ConstructorType.Kind() != reflect.Func {
		return fmt.Errorf(
			"%w: constructor for type '%s'",
			ErrorNotAFunction,
			key.Name())
	}

	if info.ConstructorReturnsError {
		if info.ConstructorType.NumOut() != 2 {
			return fmt.Errorf(
				"%w: constructor for type '%s' should have exactly two return values",
				ErrorWrongNumberOfReturns,
				key.Name())
		}
	} else {
		if info.ConstructorType.NumOut() > 1 {
			return fmt.Errorf(
				"%w: constructor for type '%s'",
				ErrorTooManyReturns,
				key.Name())
		}

		if info.ConstructorType.NumOut() == 0 {
			return fmt.Errorf(
				"%w: constructor for type '%s'",
				ErrorNoReturns,
				key.Name())
		}
	}

	if !info.ConstructorType.Out(0).AssignableTo(key) {
		return fmt.Errorf(
			"%w: constructor's return type '%s' is not assignable to type '%s'",
			ErrorNotAssignable,
			info.ConstructorType.Out(0).Name(),
			key.Name())
	}

	if info.ConstructorType.IsVariadic() {
		return fmt.Errorf(
			"%w: constructor for type '%s'",
			ErrorVariadicArguments,
			key.Name())
	}

	if _, ok := target.library.Find(key, search.Reorder); ok {
		return fmt.Errorf(
			"%w: constructor for type '%s'",
			ErrorAlreadyRegistered,
			key.Name())
	}

	nameParts := strings.Split(key.String(), ".")
	target.nameToKeyTrie.Add(nameParts[len(nameParts)-1], key)
	target.library.Add(key, info)
	return nil
}

func validPointerKey(key contracts.KeyType) bool {
	keyKind := key.Kind()
	if keyKind != reflect.Pointer {
		return false
	}

	keyElemType := key.Elem()
	if isStructLike(keyElemType) {
		return true
	}

	return false
}

func verify(injector *Injector, key contracts.KeyType) error {
	info, _ := injector.library.Find(key, search.NoReorder)
	stack := make([]contracts.ConstructorType, 0)
	stack = append(stack, info.ConstructorType)
	err := verifyStack(injector, &stack)
	if err != nil {
		sb := &strings.Builder{}
		for iRequirement, requirement := range stack {
			if iRequirement > 0 {
				sb.WriteString(" -> ")
			}

			sb.WriteString(requirement.Name())
		}

		return fmt.Errorf("%w\n\t%s", err, sb.String())
	}

	return nil
}

func verifyStack(injector *Injector, stack *[]contracts.ConstructorType) error {
	focus := (*stack)[len(*stack)-1]
	parameterCount := focus.NumIn()
	for iParameter := 0; iParameter < parameterCount; iParameter++ {
		parameterType := focus.In(iParameter)
		parameterInfo, ok := injector.library.Find(parameterType, search.NoReorder)
		if !ok {
			return fmt.Errorf(
				"%w: constructor for type '%s'",
				ErrorNotRegistered,
				parameterType.Name())
		}

		for _, requirement := range *stack {
			if requirement == parameterInfo.ConstructorType {
				return ErrorDependencyLoop
			}
		}

		*stack = append(*stack, parameterInfo.ConstructorType)
		err := verifyStack(injector, stack)
		if err != nil {
			return err
		}
	}

	*stack = (*stack)[:len(*stack)-1]
	return nil
}
