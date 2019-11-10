package shiny_test

import (
	"fmt"
	. "github.com/mweitzel/shiny/shiny"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"time"
)

func get(apiKey string, url string) string {
	time.Sleep(1 * time.Millisecond)
	return "payload for url: " + url
}

var _ = Describe("Foos", func() {

	It("returns when invoked", func() {
		fBar := F(bar)
		Ω(fBar(13)).Should(Equal(26))
	})

	Describe("Async", func() {
		It("returns nil when invoked async", func() {
			fBar := F(bar)
			fBar = fBar.Async()

			for i := 0; i < 10000; i++ {
				fBar(3)
			}

			Ω(fBar.Await()).Should(Equal(6))
		})

		It("collects return values", func() {
			fBar := F(bar)
			fBar = fBar.Async()

			for i := 0; i < 5; i++ {
				fBar(i)
			}
			collection := []int{}
			fBar.Await(&collection)

			Ω(collection).Should(ConsistOf(0, 2, 4, 6, 8))
			Ω(collection).Should(HaveLen(5))

			for i := 0; i < 5; i++ {
				fBar(i)
			}
			collection = []int{}
			fBar.Await(&collection)

			Ω(collection).Should(ConsistOf(0, 2, 4, 6, 8))
			Ω(collection).Should(HaveLen(5))
		})
	})

	Describe("Bind/Call/Apply", func() {
		It("bind 1", func() {
			Ω(F(bar).Bind(3)()).Should(Equal(6))
		})
		It("bind 2", func() {
			Ω(F(baz).Bind(1, 1, 1)()).Should(Equal(3))
		})

		It("rick bind example", func() {
			Ω(baz(1, 1, 1)).Should(Equal(3))
			plus := F(baz)
			Ω(plus(1, 1, 1)).Should(Equal(3))
			Ω(plus.Bind(1)(1, 1)).Should(Equal(3))

			plus5 := plus.Bind(5)
			Ω(plus5(4)).Should(Equal(9))

		})

		It("bind non-statefully", func() {
			fBaz := F(baz).Bind(1)
			boundFbaz := fBaz.Bind(1, 1)
			boundFbaz = boundFbaz.Bind(1)
			Ω(fBaz()).Should(Equal(1))
			Ω(boundFbaz()).Should(Equal(4))
			Ω(boundFbaz()).Should(Equal(4))
		})

		It("call", func() {
			fBaz := F(baz)
			Ω(fBaz.Call(1, 1, 1)).Should(Equal(3))
			Ω(fBaz.Bind(1).Call(1, 1, 1)).Should(Equal(4))
		})

		It("apply", func() {
			fBaz := F(baz)
			Ω(fBaz.Apply([]int{1, 1, 1})).Should(Equal(3))
			Ω(fBaz.Bind(1).Apply([]int{1, 1, 1})).Should(Equal(4))
		})

		It("shorthand", func() {
			Ω(F(baz).B(1, 1).C(1, 1, 1)).Should(Equal(5))
			Ω(F(baz).B(1, 1).A([]int{1, 1, 1})).Should(Equal(5))
		})
	})

	Describe("dig bag", func() {
		It("can get stuff", func() {
			type B struct{ Bar int }
			a := struct{ Foo B }{Foo: B{Bar: 5}}
			Ω(Dig(a, `Foo.Bar`)).Should(Equal(5))
			Ω(Dig(a, `Foo.Bat`)).Should(BeNil())
			var b binterface = bstruct{}
			Ω(fmt.Sprintf("%#v", Dig(b, `Foo`))).Should(ContainSubstring("(func(int) int)"))
			Ω(Digf(b, `Foo`)(30)).Should(Equal(60))
			Ω(Digf(b, `Bar`)).Should(BeNil())
		})

		It("can get ptr stuff", func() {
			type B struct{ Bar int }
			a := struct{ Foo B }{Foo: B{Bar: 5}}
			Ω(Dig(&a, `Foo`)).Should(Equal(B{Bar: 5}))
			Ω(Dig(&a, `Foo.Bar`)).Should(Equal(5))
			Ω(Dig(&a, `Foo.Bat`)).Should(BeNil())
			var b binterface = &bstruct{}
			Ω(Digf(b, `Foo`)(30)).Should(Equal(60))
			Ω(fmt.Sprintf("%#v", Dig(b, `XFoo`))).Should(ContainSubstring("(func(int) int)"))
			Ω(Digf(b, `XFoo`)(30)).Should(Equal(60))
			Ω(Digf(b, `Bar`)).Should(BeNil())
		})
	})

	Describe("arity", func() {
		It("detects arity for function with fixed args", func() {
			f := func() {}
			ff := F(f)
			Ω(ff.Arity()).To(Equal(0))
			g := func(i int) {}
			fg := F(g)
			Ω(fg.Arity()).To(Equal(1))
		})
		It("returns negative if the last value is variadic", func() {
			f := func(ii ...int) {}
			ff := F(f)
			Ω(ff.Arity()).To(Equal(-1))
			g := func(i int, bb ...bool) {}
			fg := F(g)
			Ω(fg.Arity()).To(Equal(-2))
		})
	})

	Describe("new args", func() {
		It("detects returns an empty args of input type", func() {
			f := func(i int) {}
			ff := F(f)
			Ω(ff.NewArgs()).To(Equal([]interface{}{
				0,
			}))

			type t struct{ foo string }
			g := func(t) {}
			Ω(F(g).NewArgs()).To(Equal([]interface{}{
				t{},
			}))

			h := func(*t) {}
			Ω(F(h).NewArgs()).To(Equal([]interface{}{
				&t{},
			}))
		})
	})

	Describe("rarity", func() {
		It("detects arity for function returning fixed args", func() {
			f := func() {}
			ff := F(f)
			Ω(ff.Rarity()).To(Equal(0))
			g := func() int { return 23 }
			fg := F(g)
			Ω(fg.Rarity()).To(Equal(1))
		})

		It("returns negative is last return is an error", func() {
			f := func() error { return err("err") }
			ff := F(f)
			Ω(ff.Rarity()).To(Equal(-1))
			g := func() (int, error) { return 23, nil }
			fg := F(g)
			Ω(fg.Rarity()).To(Equal(-2))
		})
	})

	Describe("New", func() {
		It("can Make a nil map non-nil given a pointer to it", func() {
			var foo map[string]interface{}
			New(&foo)
			foo["hi"] = 3
			Ω(foo["hi"]).Should(Equal(3))
		})

		It("can Make a nil struct non-nil given a pointer to it", func() {
			type tfoo struct{ bar string }
			var foo *tfoo
			New(&foo)
			foo.bar = "baz"
			Ω(foo.bar).Should(Equal("baz"))
		})
	})
})

type err string

func (e err) Error() string { return string(e) }

type binterface interface{ Foo(int) int }
type bstruct struct{}

func (b bstruct) Foo(i int) int   { return 2 * i }
func (b *bstruct) XFoo(i int) int { return 2 * i }

func bar(i int) int {
	time.Sleep(20 * time.Millisecond)
	return 2 * i
}

func baz(args ...int) int {
	carry := 0
	for _, a := range args {
		carry += a
	}
	return carry
}
