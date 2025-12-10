package injector

import (
	"errors"
	"fmt"
)

var (
	// InjectorError is a generic injector error, usable for detecting if an
	// error was injector related.
	InjectorError = errors.New("injector error")

	// ErrorAlreadyRegistered is returned when a type has already been
	// registered.
	ErrorAlreadyRegistered = fmt.Errorf("%w, already registered", InjectorError)

	// ErrorBadState is a panicking error when an access attempt is made on an
	// injector that is in a bad state.
	ErrorBadState = fmt.Errorf("%w, bad injector state", InjectorError)

	// ErrorDependencyLoop indicates that an unsolvable dependency injection
	// loop.
	ErrorDependencyLoop = fmt.Errorf("%w, dependency loop detected", InjectorError)

	// ErrorNoReturns is returned when a constructor has no return value.
	ErrorNoReturns = fmt.Errorf("%w, no return values, must be exactly 1 return value", InjectorError)

	// ErrorNotAFunction is returned when a non-function is passed as a function.
	ErrorNotAFunction = fmt.Errorf("%w, value is not a function", InjectorError)

	// ErrorNotAssignable is returned when a constructor returns a type that
	// cannot be assigned to the key type.
	ErrorNotAssignable = fmt.Errorf("%w, value is not assignable", InjectorError)

	// ErrorNotRegistered indicates that a required dependency does not appear
	// in the registered list.
	ErrorNotRegistered = fmt.Errorf("%w, not registered", InjectorError)

	// ErrorNotStructOrInterface is returned when a type is not registerable.
	ErrorNotStructOrInterface = fmt.Errorf("%w, key type is not a struct or interface", InjectorError)

	// ErrorTooManyReturns is returned when a constructor has more than 1 return
	// value.
	ErrorTooManyReturns = fmt.Errorf("%w, too many return values, must be exactly 1 return value", InjectorError)

	// ErrorVariadicArguments is returned when a function has a variadic
	// signature.
	ErrorVariadicArguments = fmt.Errorf("%w, function has a variadic signature", InjectorError)

	// ErrorWrongNumberOfReturns is returned when a function has more or less
	// than the expected number of return values.
	ErrorWrongNumberOfReturns = fmt.Errorf("%w, wrong number of return values", InjectorError)
)
