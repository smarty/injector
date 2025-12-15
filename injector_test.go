package injector

import (
	"errors"
	"reflect"
	"testing"

	"github.com/smarty/assertions/should"
	"github.com/smarty/gunit"
	"github.com/smarty/injector/internal/contracts"
	. "github.com/smarty/injector/internal/test"
)

func TestInjectorFixture(t *testing.T) {
	gunit.Run(new(InjectorFixture), t)
}

type InjectorFixture struct {
	*gunit.Fixture
}

func (this *InjectorFixture) Setup() {
}

func (this *InjectorFixture) TestTypeAlreadyRegistered() {
	di := New()
	err := RegisterSingleton[Car](di, NewRegularCar)
	this.So(err, should.BeNil)

	err = RegisterSingleton[Car](di, NewRegularCar)
	this.So(err, should.NotBeNil)
	this.So(err, should.Wrap, ErrorAlreadyRegistered)
}

func (this *InjectorFixture) TestBadStateErrors() {
	di := New()
	err := RegisterSingleton[Car](di, NewRegularCar)
	this.So(err, should.BeNil)

	_, err = Get[Car](di)
	this.So(err, should.NotBeNil)
	this.So(err, should.Wrap, ErrorBadState)
}

func (this *InjectorFixture) TestDependencyLoop() {
	di := New()
	err := RegisterSingleton[Car](di, NewRegularCar)
	this.So(err, should.BeNil)

	err = RegisterSingleton[Driver](di, NewLoopDriver)
	this.So(err, should.BeNil)

	err = Verify(di)
	this.So(err, should.NotBeNil)
	this.So(err, should.Wrap, ErrorDependencyLoop)
}

func (this *InjectorFixture) TestNoReturnValues() {
	di := New()
	err := RegisterSingleton[Car](di, func() {})
	this.So(err, should.NotBeNil)
	this.So(err, should.Wrap, ErrorNoReturns)
}

func (this *InjectorFixture) TestNotAFunction() {
	di := New()
	err := RegisterSingleton[Car](di, NewRegularCar(nil))
	this.So(err, should.NotBeNil)
	this.So(err, should.Wrap, ErrorNotAFunction)
}

func (this *InjectorFixture) TestNotAssignable() {
	di := New()
	err := RegisterSingleton[Car](di, NewRegularDriver)
	this.So(err, should.NotBeNil)
	this.So(err, should.Wrap, ErrorNotAssignable)
}

func (this *InjectorFixture) TestNotRegistered() {
	di := New()
	err := RegisterSingleton[Car](di, NewRegularCar)
	this.So(err, should.BeNil)

	err = Verify(di)
	this.So(err, should.NotBeNil)
	this.So(err, should.Wrap, ErrorNotRegistered)
	this.So(err.Error(), should.ContainSubstring, "Driver")
}

func (this *InjectorFixture) TestGetUnregisteredType_ReturnsErrorNotRegistered() {
	di := New()

	err := Verify(di)
	this.So(err, should.BeNil)

	_, getErr := Get[Driver](di)
	this.So(getErr, should.NotBeNil)
	this.So(getErr, should.Wrap, ErrorNotRegistered)
	this.So(getErr.Error(), should.ContainSubstring, "Driver")
}

func (this *InjectorFixture) TestNotAStructOrInterface() {
	di := New()
	err := RegisterSingleton[int](di, NewRegularCar)
	this.So(err, should.NotBeNil)
	this.So(err, should.Wrap, ErrorNotStructOrInterface)
}

func (this *InjectorFixture) TestTooManyReturns() {
	di := New()
	err := RegisterSingleton[Car](di, func() (Car, error) { return NewRegularCar(nil), nil })
	this.So(err, should.NotBeNil)
	this.So(err, should.Wrap, ErrorTooManyReturns)
}

func (this *InjectorFixture) TestVariadicArguments() {
	di := New()
	err := RegisterSingleton[Car](di, func(driver ...Driver) Car { return NewRegularCar(nil) })
	this.So(err, should.NotBeNil)
	this.So(err, should.Wrap, ErrorVariadicArguments)
}

func (this *InjectorFixture) TestRegisterSingletonError_WrongNumberOfReturns() {
	di := New()
	err := RegisterSingletonError[Car](di, func() Car { return NewRegularCar(nil) })
	this.So(err, should.NotBeNil)
	this.So(err, should.Wrap, ErrorWrongNumberOfReturns)
}

func (this *InjectorFixture) TestInjectorRegisterSingletonError_WrongNumberOfReturns() {
	di := New()
	err := di.RegisterSingletonError(reflect.TypeFor[Car](), func() Car { return NewRegularCar(nil) })
	this.So(err, should.NotBeNil)
	this.So(err, should.Wrap, ErrorWrongNumberOfReturns)
}

func (this *InjectorFixture) TestRegisterSingletonError_RuntimeErrorPropagates() {
	di := New()
	err := RegisterSingletonError[Car](di, func() (Car, error) { return nil, errors.New("boom") })
	this.So(err, should.BeNil)

	err = Verify(di)
	this.So(err, should.BeNil)

	_, getErr := Get[Car](di)
	this.So(getErr, should.NotBeNil)
	this.So(getErr.Error(), should.ContainSubstring, "boom")
}

func (this *InjectorFixture) TestTransient() {
	di := New()
	err := RegisterTransient[Counter](di, NewCallCounter)
	this.So(err, should.BeNil)

	err = Verify(di)
	this.So(err, should.BeNil)

	skipError(Get[Counter](di)).CallMe()
	skipError(Get[Counter](di)).CallMe()
	skipError(Get[Counter](di)).CallMe()
	skipError(Get[Counter](di)).CallMe()
	skipError(Get[Counter](di)).CallMe()
	this.So(skipError(Get[Counter](di)).GetCount(), should.Equal, 0)
}

func (this *InjectorFixture) TestSingleton() {
	di := New()
	err := RegisterSingleton[Counter](di, NewCallCounter)
	this.So(err, should.BeNil)

	err = Verify(di)
	this.So(err, should.BeNil)

	skipError(Get[Counter](di)).CallMe()
	skipError(Get[Counter](di)).CallMe()
	skipError(Get[Counter](di)).CallMe()
	skipError(Get[Counter](di)).CallMe()
	skipError(Get[Counter](di)).CallMe()
	this.So(skipError(Get[Counter](di)).GetCount(), should.Equal, 5)
}

func (this *InjectorFixture) TestScope() {
	di := New()
	err := RegisterScope[Counter](di, NewCallCounter)
	this.So(err, should.BeNil)
	err = RegisterTransient[CounterWrapper](di, NewCallCounterWrapper)
	this.So(err, should.BeNil)

	err = Verify(di)
	this.So(err, should.BeNil)

	cw := skipError(Get[CounterWrapper](di))
	cw.CallLeft()
	cw.CallLeft()
	this.So(cw.GetRightCount(), should.Equal, 2)

	cw = skipError(Get[CounterWrapper](di))
	this.So(cw.GetLeftCount(), should.Equal, 0)
	cw.CallLeft()
	cw.CallLeft()
	this.So(cw.GetRightCount(), should.Equal, 2)
}

func (this *InjectorFixture) TestGetCorrectlyFillsInstance() {
	di := New()
	err := RegisterTransient[Car](di, NewRegularCar)
	this.So(err, should.BeNil)
	err = RegisterTransient[Driver](di, NewRegularDriver)
	this.So(err, should.BeNil)
	err = Verify(di)
	this.So(err, should.BeNil)

	car := skipError(Get[Car](di))
	this.So(car, should.NotBeNil)
	_, ok := car.(*RegularCar)
	this.So(ok, should.BeTrue)
	_, ok = car.GetDriver().(*RegularDriver)
	this.So(ok, should.BeTrue)
}

func (this *InjectorFixture) TestAddSelf() {
	di := New()
	err := Verify(di)
	this.So(err, should.BeNil)

	inj2 := skipError(Get[*Injector](di))
	this.So(inj2, should.Equal, di)
}

func (this *InjectorFixture) TestGetByNameNotRegistered() {
	di := New()
	err := RegisterTransient[Driver](di, NewRegularDriver)
	this.So(err, should.BeNil)
	err = Verify(di)
	this.So(err, should.BeNil)

	var byName any
	byName, err = GetByName(di, "Car")
	this.So(err, should.NotBeNil)
	this.So(err, should.Wrap, ErrorNotRegistered)
	this.So(byName, should.BeNil)
}

func (this *InjectorFixture) TestGetByName() {
	di := New()
	err := RegisterTransient[Car](di, NewRegularCar)
	this.So(err, should.BeNil)
	err = RegisterTransient[Driver](di, NewRegularDriver)
	this.So(err, should.BeNil)
	err = Verify(di)
	this.So(err, should.BeNil)

	var byName any
	byName, err = GetByName(di, "Car")
	this.So(err, should.BeNil)
	this.So(byName, should.NotBeNil)
	car, ok := byName.(*RegularCar)
	this.So(ok, should.BeTrue)
	_, ok = car.GetDriver().(*RegularDriver)
	this.So(ok, should.BeTrue)
}

func (this *InjectorFixture) TestGetByNameFromAnotherPackage() {
	di := New()
	err := RegisterTransient[*contracts.ScopedInstance](di, func() *contracts.ScopedInstance { return new(contracts.ScopedInstance) })
	this.So(err, should.BeNil)
	err = Verify(di)
	this.So(err, should.BeNil)

	var byName any
	byName, err = GetByName(di, "ScopedInstance")
	this.So(err, should.BeNil)
	this.So(byName, should.NotBeNil)
	_, ok := byName.(*contracts.ScopedInstance)
	this.So(ok, should.BeTrue)
}

func (this *InjectorFixture) TestCall_Method() {
	strings := []string{"hello", "world", "how", "are", "you"}

	di := New()
	err := RegisterTransient[*StringProvider](di, func() *StringProvider { return NewStringProvider(strings...) })
	this.So(err, should.BeNil)
	err = Verify(di)
	this.So(err, should.BeNil)

	di.Call(func(sp *StringProvider) {})
}

func (this *InjectorFixture) TestCall_Method_Variadic() {
	strings := []string{"hello", "world", "how", "are", "you"}

	di := New()
	err := RegisterTransient[*StringProvider](di, func() *StringProvider { return NewStringProvider(strings...) })
	this.So(err, should.BeNil)
	err = Verify(di)
	this.So(err, should.BeNil)

	err = di.Call(func(sp ...*StringProvider) {})
	this.So(err, should.NotBeNil)
}

func (this *InjectorFixture) TestCall_Method_WrongNumberOfReturns() {
	strings := []string{"hello", "world", "how", "are", "you"}

	di := New()
	err := RegisterTransient[*StringProvider](di, func() *StringProvider { return NewStringProvider(strings...) })
	this.So(err, should.BeNil)
	err = Verify(di)
	this.So(err, should.BeNil)

	err = di.Call(func(sp *StringProvider) any { return sp.Values[0] })
	this.So(err, should.NotBeNil)
}

func (this *InjectorFixture) TestCall_Method_NotAFunction() {
	strings := []string{"hello", "world", "how", "are", "you"}

	di := New()
	err := RegisterTransient[*StringProvider](di, func() *StringProvider { return NewStringProvider(strings...) })
	this.So(err, should.BeNil)
	err = Verify(di)
	this.So(err, should.BeNil)

	err = di.Call(strings)
	this.So(err, should.NotBeNil)
}

func (this *InjectorFixture) TestCall1_Method() {
	strings := []string{"hello", "world", "how", "are", "you"}

	di := New()
	err := RegisterTransient[*StringProvider](di, func() *StringProvider { return NewStringProvider(strings...) })
	this.So(err, should.BeNil)
	err = Verify(di)
	this.So(err, should.BeNil)

	raw1, e := di.Call1(func(sp *StringProvider) any { return sp.Values[0] })
	string1 := raw1.(string)

	this.So(string1, should.Equal, strings[0])
	this.So(e, should.BeNil)
}

func (this *InjectorFixture) TestCall2_Method() {
	strings := []string{"hello", "world", "how", "are", "you"}

	di := New()
	err := RegisterTransient[*StringProvider](di, func() *StringProvider { return NewStringProvider(strings...) })
	this.So(err, should.BeNil)
	err = Verify(di)
	this.So(err, should.BeNil)

	raw1, raw2, e := di.Call2(func(sp *StringProvider) (any, any) { return sp.Values[0], sp.Values[1] })
	string1 := raw1.(string)
	string2 := raw2.(string)

	this.So(string1, should.Equal, strings[0])
	this.So(string2, should.Equal, strings[1])
	this.So(e, should.BeNil)
}

func (this *InjectorFixture) TestCall3_Method() {
	strings := []string{"hello", "world", "how", "are", "you"}

	di := New()
	err := RegisterTransient[*StringProvider](di, func() *StringProvider { return NewStringProvider(strings...) })
	this.So(err, should.BeNil)
	err = Verify(di)
	this.So(err, should.BeNil)

	raw1, raw2, raw3, e := di.Call3(func(sp *StringProvider) (any, any, any) { return sp.Values[0], sp.Values[1], sp.Values[2] })
	string1 := raw1.(string)
	string2 := raw2.(string)
	string3 := raw3.(string)

	this.So(string1, should.Equal, strings[0])
	this.So(string2, should.Equal, strings[1])
	this.So(string3, should.Equal, strings[2])
	this.So(e, should.BeNil)
}

func (this *InjectorFixture) TestCall4_Method() {
	strings := []string{"hello", "world", "how", "are", "you"}

	di := New()
	err := RegisterTransient[*StringProvider](di, func() *StringProvider { return NewStringProvider(strings...) })
	this.So(err, should.BeNil)
	err = Verify(di)
	this.So(err, should.BeNil)

	raw1, raw2, raw3, raw4, e := di.Call4(func(sp *StringProvider) (any, any, any, any) {
		return sp.Values[0], sp.Values[1], sp.Values[2], sp.Values[3]
	})
	string1 := raw1.(string)
	string2 := raw2.(string)
	string3 := raw3.(string)
	string4 := raw4.(string)

	this.So(string1, should.Equal, strings[0])
	this.So(string2, should.Equal, strings[1])
	this.So(string3, should.Equal, strings[2])
	this.So(string4, should.Equal, strings[3])
	this.So(e, should.BeNil)
}

func (this *InjectorFixture) TestCallN_Method() {
	strings := []string{"hello", "world", "how", "are", "you"}

	di := New()
	err := RegisterTransient[*StringProvider](di, func() *StringProvider { return NewStringProvider(strings...) })
	this.So(err, should.BeNil)
	err = Verify(di)
	this.So(err, should.BeNil)

	raws, e := di.CallN(func(sp *StringProvider) (any, any, any, any, any) {
		return sp.Values[0], sp.Values[1], sp.Values[2], sp.Values[3], sp.Values[4]
	})
	string1 := raws[0].(string)
	string2 := raws[1].(string)
	string3 := raws[2].(string)
	string4 := raws[3].(string)
	string5 := raws[4].(string)

	this.So(string1, should.Equal, strings[0])
	this.So(string2, should.Equal, strings[1])
	this.So(string3, should.Equal, strings[2])
	this.So(string4, should.Equal, strings[3])
	this.So(string5, should.Equal, strings[4])
	this.So(e, should.BeNil)
}

func (this *InjectorFixture) TestCall_Function_Variadic() {
	strings := []string{"hello", "world", "how", "are", "you"}

	di := New()
	err := RegisterTransient[*StringProvider](di, func() *StringProvider { return NewStringProvider(strings...) })
	this.So(err, should.BeNil)
	err = Verify(di)
	this.So(err, should.BeNil)

	err = Call(di, func(sp ...*StringProvider) {})
	this.So(err, should.NotBeNil)
}

func (this *InjectorFixture) TestCall_Function_WrongNumberOfReturns() {
	strings := []string{"hello", "world", "how", "are", "you"}

	di := New()
	err := RegisterTransient[*StringProvider](di, func() *StringProvider { return NewStringProvider(strings...) })
	this.So(err, should.BeNil)
	err = Verify(di)
	this.So(err, should.BeNil)

	err = Call(di, func(sp *StringProvider) any { return sp.Values[0] })
	this.So(err, should.NotBeNil)
}

func (this *InjectorFixture) TestCall_Function_NotAFunction() {
	strings := []string{"hello", "world", "how", "are", "you"}

	di := New()
	err := RegisterTransient[*StringProvider](di, func() *StringProvider { return NewStringProvider(strings...) })
	this.So(err, should.BeNil)
	err = Verify(di)
	this.So(err, should.BeNil)

	err = Call(di, strings)
	this.So(err, should.NotBeNil)
}

func (this *InjectorFixture) TestCall_Function() {
	strings := []string{"hello", "world", "how", "are", "you"}

	di := New()
	err := RegisterTransient[*StringProvider](di, func() *StringProvider { return NewStringProvider(strings...) })
	this.So(err, should.BeNil)
	err = Verify(di)
	this.So(err, should.BeNil)

	Call(di, func(sp *StringProvider) {})
}

func (this *InjectorFixture) TestCall1_Function() {
	strings := []string{"hello", "world", "how", "are", "you"}

	di := New()
	err := RegisterTransient[*StringProvider](di, func() *StringProvider { return NewStringProvider(strings...) })
	this.So(err, should.BeNil)
	err = Verify(di)
	this.So(err, should.BeNil)

	string1, e := Call1[string](di, func(sp *StringProvider) string { return sp.Values[0] })

	this.So(string1, should.Equal, strings[0])
	this.So(e, should.BeNil)
}

func (this *InjectorFixture) TestCall2_Function() {
	strings := []string{"hello", "world", "how", "are", "you"}

	di := New()
	err := RegisterTransient[*StringProvider](di, func() *StringProvider { return NewStringProvider(strings...) })
	this.So(err, should.BeNil)
	err = Verify(di)
	this.So(err, should.BeNil)

	string1, string2, e := Call2[string, string](di, func(sp *StringProvider) (string, string) { return sp.Values[0], sp.Values[1] })

	this.So(string1, should.Equal, strings[0])
	this.So(string2, should.Equal, strings[1])
	this.So(e, should.BeNil)
}

func (this *InjectorFixture) TestCall3_Function() {
	strings := []string{"hello", "world", "how", "are", "you"}

	di := New()
	err := RegisterTransient[*StringProvider](di, func() *StringProvider { return NewStringProvider(strings...) })
	this.So(err, should.BeNil)
	err = Verify(di)
	this.So(err, should.BeNil)

	string1, string2, string3, e := Call3[string, string, string](di, func(sp *StringProvider) (string, string, string) { return sp.Values[0], sp.Values[1], sp.Values[2] })

	this.So(string1, should.Equal, strings[0])
	this.So(string2, should.Equal, strings[1])
	this.So(string3, should.Equal, strings[2])
	this.So(e, should.BeNil)
}

func (this *InjectorFixture) TestCall4_Function() {
	strings := []string{"hello", "world", "how", "are", "you"}

	di := New()
	err := RegisterTransient[*StringProvider](di, func() *StringProvider { return NewStringProvider(strings...) })
	this.So(err, should.BeNil)
	err = Verify(di)
	this.So(err, should.BeNil)

	string1, string2, string3, string4, e := Call4[string, string, string, string](di, func(sp *StringProvider) (string, string, string, string) {
		return sp.Values[0], sp.Values[1], sp.Values[2], sp.Values[3]
	})

	this.So(string1, should.Equal, strings[0])
	this.So(string2, should.Equal, strings[1])
	this.So(string3, should.Equal, strings[2])
	this.So(string4, should.Equal, strings[3])
	this.So(e, should.BeNil)
}

func (this *InjectorFixture) TestCallN_Function() {
	strings := []string{"hello", "world", "how", "are", "you"}

	di := New()
	err := RegisterTransient[*StringProvider](di, func() *StringProvider { return NewStringProvider(strings...) })
	this.So(err, should.BeNil)
	err = Verify(di)
	this.So(err, should.BeNil)

	raws, e := CallN(di, func(sp *StringProvider) (any, any, any, any, any) {
		return sp.Values[0], sp.Values[1], sp.Values[2], sp.Values[3], sp.Values[4]
	})
	string1 := raws[0].(string)
	string2 := raws[1].(string)
	string3 := raws[2].(string)
	string4 := raws[3].(string)
	string5 := raws[4].(string)

	this.So(string1, should.Equal, strings[0])
	this.So(string2, should.Equal, strings[1])
	this.So(string3, should.Equal, strings[2])
	this.So(string4, should.Equal, strings[3])
	this.So(string5, should.Equal, strings[4])
	this.So(e, should.BeNil)
}

func skipError[T any](value T, err error) T {
	return value
}
