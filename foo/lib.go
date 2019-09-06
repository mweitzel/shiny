package foo

import (
	"reflect"
	"strings"
	"sync"
)

func F(fn interface{}) mySpecial {
	i := impl{fn: fn, wg: sync.WaitGroup{}}
	return (&i).special()
}

type impl struct {
	fn      interface{}
	async   bool
	lastOut interface{}
	wg      sync.WaitGroup
	bArgs   []interface{}
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
		i.bArgs = append([]interface{}{}, i.bArgs...)
		i.bArgs = append(i.bArgs, bArgs.([]interface{})...)
		return i
	}

	if arg["await"] == true {
		i.wg.Wait()
		//		return i.lastOut
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
				i.lastOut = outs[0].Interface()
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
	return i.lastOut
}

func (ms mySpecial) Async() mySpecial {
	return func(args ...interface{}) interface{} {
		ms(cheat{"async": true})
		return ms(args...)
	}
}

func (ms mySpecial) Bind(bArgs ...interface{}) mySpecial {
	// problematic, this does not duplicate impl
	i := ms(cheat{"bind": bArgs}).(*impl)
	return i.special()
}

type cheat map[string]interface{}

type special func(...interface{}) interface{}

//func Async

func Digf(obj interface{}, path string) mySpecial {
	return F(Dig(obj, path))
}

func Dig(obj interface{}, path string) interface{} {
	current := obj
	for _, v := range strings.Split(path, ".") {
		rStruct := reflect.ValueOf(current)
		if rStruct.Kind() == reflect.Struct {
			field := rStruct.FieldByName(v)
			if field.Kind() != reflect.Invalid {
				if !IsZero(field) {
					current = field.Interface()
					continue
				}
			}
			method := rStruct.MethodByName(v)
			if method.Kind() != reflect.Invalid {
				//			fmt.Println("..............", method)
				//			fmt.Println("..............", method.Kind())
				//			fmt.Println("..............", method.IsValid())
				//			fmt.Println("..............", method.CanInterface())
				//			fmt.Printf(".............%#v\n", method.Interface())
				//			fmt.Printf(".............%#v\n", method.Interface())
				current = method.Interface()
				continue
			}
		}
		return nil
	}
	return current
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
