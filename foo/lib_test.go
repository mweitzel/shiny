package foo_test

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	"fmt"
	. "mymod/foo"
	"time"
)

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

			for i := 0; i < 5; i++ {
				fBar(i)
			}
			collection = []int{}
			fBar.Await(&collection)

			Ω(collection).Should(ConsistOf(0, 2, 4, 6, 8))
		})
	})

	Describe("Bind/Call/Apply", func() {
		It("bind 1", func() {
			Ω(F(bar).Bind(3)()).Should(Equal(6))
		})
		It("bind 2", func() {
			Ω(F(baz).Bind(1, 1, 1)()).Should(Equal(3))
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
		})
	})
})

type binterface interface{ Foo(int) int }
type bstruct struct{}

func (b bstruct) Foo(i int) int { return 2 * i }

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
