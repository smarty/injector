package test

// ----- interfaces

type Car interface {
	GetDriver() Driver
}

type Driver interface {
	GetName() string
}

type Counter interface {
	CallMe()
	GetCount() int
}

type CounterWrapper interface {
	CallLeft()
	CallRight()
	GetLeftCount() int
	GetRightCount() int
}

// ----- structs

type RegularCar struct {
	driver Driver
}

type RegularDriver struct {
}

type LoopDriver struct {
	car Car
}

type CallCounter struct {
	count int
}

type CallCounterWrapper struct {
	left  Counter
	right Counter
}

type StringProvider struct {
	Values []string
}

// ----- constructors

func NewRegularCar(driver Driver) Car {
	return &RegularCar{
		driver: driver,
	}
}

func NewRegularDriver() Driver {
	return &RegularDriver{}
}

func NewLoopDriver(car Car) Driver {
	return &LoopDriver{
		car: car,
	}
}

func NewCallCounter() *CallCounter {
	return &CallCounter{
		count: 0,
	}
}

func NewCallCounterWrapper(left Counter, right Counter) *CallCounterWrapper {
	return &CallCounterWrapper{
		left:  left,
		right: right,
	}
}

func NewStringProvider(strings ...string) *StringProvider {
	return &StringProvider{
		Values: strings,
	}
}

// ----- methods

func (this *RegularCar) GetDriver() Driver {
	return this.driver
}

func (this *RegularDriver) GetName() string {
	return "Norman"
}

func (this *LoopDriver) GetName() string {
	return "Lupin"
}

func (this *CallCounter) CallMe() {
	this.count++
}

func (this *CallCounter) GetCount() int {
	return this.count
}

func (this *CallCounterWrapper) CallLeft() {
	this.left.CallMe()
}

func (this *CallCounterWrapper) CallRight() {
	this.right.CallMe()
}

func (this *CallCounterWrapper) GetLeftCount() int {
	return this.left.GetCount()
}

func (this *CallCounterWrapper) GetRightCount() int {
	return this.right.GetCount()
}
