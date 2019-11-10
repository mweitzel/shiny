package shiny

import (
	"github.com/google/uuid"
	"reflect"
	"strings"
	"sync"
)

func F(fn interface{}) mySpecial {
	i := impl{fn: fn, wg: sync.WaitGroup{}, collector: sync.Map{}}
	return (&i).special()
}

type impl struct {
	fn        interface{}
	async     bool
	lastOut   interface{}
	wg        sync.WaitGroup
	bArgs     []interface{}
	collector sync.Map
}

func (i *impl) handleCase(arg cheat) *impl {
	if arg["async"] == true {
		i.async = true
	} else {
		i.async = false
	}

	bArgs, ok := arg["bind"]
	if ok {
		ix := *i
		i = &ix
		i.collector = sync.Map{}
		i.bArgs = append([]interface{}{}, i.bArgs...)
		i.bArgs = append(i.bArgs, bArgs.([]interface{})...)
		return i
	}

	if arg["await"] == true {
		i.wg.Wait()
		return i
	}

	if arg["inspect"] == true {
		return i
	}
	return i
}

func (i *impl) special() func(args ...interface{}) interface{} {
	return func(args ...interface{}) interface{} {
		if len(args) != 0 {
			a, ok := args[0].(cheat)
			if ok {
				return i.handleCase(a)
			}
		}

		xInputs := []interface{}{}
		xInputs = append(xInputs, i.bArgs...)
		xInputs = append(xInputs, args...)
		if i.async {
			i.wg.Add(1)
			go func() {
				defer i.wg.Done()
				inputs := []reflect.Value{}
				for _, arg := range xInputs {
					inputs = append(inputs, reflect.ValueOf(arg))
				}
				outs := reflect.ValueOf(i.fn).Call(inputs)
				if len(outs) == 0 {
					return
				}
				lOut := outs[0].Interface()
				i.lastOut = lOut
				i.collector.Store(uuid.New().String(), lOut)
			}()
			return nil
		} else {
			inputs := []reflect.Value{}
			for _, arg := range xInputs {
				inputs = append(inputs, reflect.ValueOf(arg))
			}
			outs := reflect.ValueOf(i.fn).Call(inputs)
			if len(outs) == 0 {
				return nil
			}
			return outs[0].Interface()
		}
	}
}

type mySpecial func(...interface{}) interface{}

func (ms mySpecial) Await(args ...interface{}) interface{} {
	i := ms(cheat{"await": true}).(*impl)
	if len(args) > 0 {
		collect := []interface{}{}
		i.collector.Range(func(_, v interface{}) bool {
			collect = append(collect, v)
			return true
		})
		ptrToCollectSlice := args[0]
		rSlice := reflect.ValueOf(ptrToCollectSlice).Elem()
		for _, v := range collect {
			rSlice.Set(reflect.Append(rSlice, reflect.ValueOf(v)))
		}
		i.collector = sync.Map{}
	}
	return i.lastOut
}

func (ms mySpecial) Async() mySpecial {
	return func(args ...interface{}) interface{} {
		ms(cheat{"async": true})
		return ms(args...)
	}
}

func (ms mySpecial) B(bArgs ...interface{}) mySpecial { return ms.Bind(bArgs...) }
func (ms mySpecial) Bind(bArgs ...interface{}) mySpecial {
	i := ms(cheat{"bind": bArgs}).(*impl)
	return i.special()
}

func (ms mySpecial) C(bArgs ...interface{}) interface{} { return ms.Call(bArgs...) }
func (ms mySpecial) Call(bArgs ...interface{}) interface{} {
	return ms(bArgs...)
}

func (ms mySpecial) A(argSlice interface{}) interface{} { return ms.Apply(argSlice) }
func (ms mySpecial) Apply(argSlice interface{}) interface{} {
	rSlice := reflect.ValueOf(argSlice)
	args := []interface{}{}
	for i := 0; i < rSlice.Len(); i++ {
		args = append(args, rSlice.Index(i).Interface())
	}
	return ms(args...)
}

func (ms mySpecial) Arity() int {
	i := ms(cheat{"inspect": true}).(*impl)
	rTyp := reflect.TypeOf(i.fn)
	n := rTyp.NumIn()
	if n != 0 && rTyp.IsVariadic() {
		n = -n
	}
	return n
}

func (ms mySpecial) NewArgs() []interface{} {
	i := ms(cheat{"inspect": true}).(*impl)
	outs := []interface{}{}
	for j := 0; j < reflect.TypeOf(i.fn).NumIn(); j++ {
		var out interface{}
		outType := reflect.TypeOf(i.fn).In(j)
		if outType.Kind() == reflect.Ptr {
			out = reflect.New(outType.Elem()).Interface()
		} else {
			out = reflect.New(outType).Elem().Interface()
		}
		outs = append(outs, out)
	}
	return outs
}

func (ms mySpecial) Rarity() int {
	i := ms(cheat{"inspect": true}).(*impl)
	rTyp := reflect.TypeOf(i.fn)
	n := rTyp.NumOut()
	if n != 0 {
		errorInterface := reflect.TypeOf((*error)(nil)).Elem()
		last := rTyp.Out(n - 1)
		if last.Implements(errorInterface) {
			return -n
		}
	}
	return n
}

type cheat map[string]interface{}

type special func(...interface{}) interface{}

func Digf(obj interface{}, path string) mySpecial {
	maybeFn := Dig(obj, path)
	if maybeFn == nil {
		return nil
	}
	return F(maybeFn)
}

type Bar struct{}

type Foo interface {
	Shit(int) error
}

func New(ptrToThing interface{}) {
	rpThing := reflect.ValueOf(ptrToThing)
	indirectThing := reflect.Indirect(rpThing)

	switch rpThing.Elem().Kind() {
	case reflect.Ptr: // because pointer to struct?
		newThing := reflect.New(indirectThing.Type().Elem())
		indirectThing.Set(newThing)
	case reflect.Map:
		newThing := reflect.MakeMap(indirectThing.Type())
		indirectThing.Set(newThing)
	}
}

func wat(interface{}) {}

func Dig(obj interface{}, path string) interface{} {
	current := obj
	for _, v := range strings.Split(path, ".") {
		rStruct := reflect.ValueOf(current)

		var found bool
		current, found = getField(rStruct, v)
		if found {
			continue
		}

		current, found = getMethod(rStruct, v)
		if found {
			continue
		}

		return nil
	}
	return current
}

func getField(rVal reflect.Value, name string) (interface{}, bool) {
	if rVal.Kind() == reflect.Struct {
		field := rVal.FieldByName(name)
		if field.Kind() != reflect.Invalid {
			return field.Interface(), true
		}
	} else if rVal.Kind() == reflect.Ptr {
		return getField(reflect.Indirect(rVal), name)
	}
	return nil, false
}

func getMethod(rVal reflect.Value, name string) (interface{}, bool) {
	if rVal.Kind() == reflect.Struct {
		method := rVal.MethodByName(name)
		if method.Kind() != reflect.Invalid {
			return method.Interface(), true
		}
	} else if rVal.Kind() == reflect.Ptr {
		method := rVal.MethodByName(name)

		if method.Kind() != reflect.Invalid {
			return method.Interface(), true
		}
	}
	return nil, false
}

// go 1.11 bullshit, bump go and use reflect.Value.IsZero()
func IsZero(v reflect.Value) bool {
	switch v.Kind() {
	case reflect.Array, reflect.String:
		return v.Len() == 0
	case reflect.Bool:
		return !v.Bool()
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		return v.Int() == 0
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64, reflect.Uintptr:
		return v.Uint() == 0
	case reflect.Float32, reflect.Float64:
		return v.Float() == 0
	case reflect.Interface, reflect.Map, reflect.Ptr, reflect.Slice:
		return v.IsNil()
	}
	return false
}
