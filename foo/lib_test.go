package foo_test

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

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
	})
})

func bar(i int) int {
	time.Sleep(20 * time.Millisecond)
	return 2 * i
}
