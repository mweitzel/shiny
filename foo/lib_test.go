package foo

import (
	"testing"
	"time"
)

func TestFBar(t *testing.T) {
	fBar := F(bar)
	out := fBar(43)
	if out != 86 {
		t.Log(out)
		t.Fail()
	}
}

func TestAsync(t *testing.T) {
	fBar := F(bar).Async()
	var nilout interface{} = "this will be overwritten"
	for i := 0; i < 100000; i++ {
		nilout = fBar(44)
	}
	out := fBar.Await()
	if nilout != nil {
		t.Log(nilout)
		t.Fail()
	}
	if out != 88 {
		t.Log(out)
		t.Fail()
	}
}

func TestFail(t *testing.T) {
	t.Fail()
}

func bar(i int) int {
	time.Sleep(20 * time.Millisecond)
	return 2 * i
}
