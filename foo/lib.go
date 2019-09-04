package foo

import (
	"reflect"
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
}

func (i *impl) handleCase(arg cheat) interface{} {
	if arg["async"] == true {
		i.async = true
	} else {
		i.async = false
	}

	if arg["await"] == true {
		i.wg.Wait()
		return i.lastOut
	}
	return nil
}

func (i *impl) special() func(args ...interface{}) interface{} {
	return func(args ...interface{}) interface{} {
		if len(args) != 0 {
			a, ok := args[0].(cheat)
			if ok {
				return i.handleCase(a)
			}
		}

		if i.async {
			i.wg.Add(1)
			go func() {
				defer i.wg.Done()
				inputs := []reflect.Value{}
				for _, arg := range args {
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
			for _, arg := range args {
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
	return ms(cheat{"await": true})
}

func (ms mySpecial) Async() mySpecial {
	return func(args ...interface{}) interface{} {
		ms(cheat{"async": true})
		return ms(args...)
	}
}

type cheat map[string]interface{}

type special func(...interface{}) interface{}

//func Async
