package queue

import (
	"github.com/redresseur/message_bus/open_interface"
	"sync"
	"testing"
)

var (
	rw   open_interface.QueueRW
	once sync.Once
)

func testInit() {
	rw = NewQueueRW(true)

	// 写100个数字
	for i := 0; i < 100; i++ {
		rw.Push(i)
	}
}

func TestQueueImpl_Next(t *testing.T) {
	once.Do(testInit)

	v, err := rw.Next(false)
	if err != nil {
		t.Fatalf("Next not remove: %v", err)
	} else if v != 0 {
		t.Fatalf("Next not remove, the value is wrong")
	}

	res, err := rw.Seek(0, 0)
	if err != nil {
		t.Fatalf("Seek: %v", err)
	} else {
		t.Logf("Seek res %v", res)
	}

	v, err = rw.Next(true)
	if err != nil {
		t.Fatalf("Next not remove: %v", err)
	} else if v != 0 {
		t.Fatalf("Next not remove, the value is wrong")
	}

	res, err = rw.Seek(0, 0)
	if err != nil {
		t.Fatalf("Seek: %v", err)
	} else {
		t.Logf("Seek res %v", res)
	}
}

func TestQueueImpl_Remove(t *testing.T) {
	once.Do(testInit)

	if err := rw.Remove(-1, 0); err != nil {
		t.Logf("Remove: %v", err)
	} else {
		t.Fatalf("Remove not passed")
	}

	if err := rw.Remove(2, -1); err != nil {
		t.Logf("Remove: %v", err)
	} else {
		t.Fatalf("Remove not passed")
	}

	if err := rw.Remove(5, 1); err != nil {
		t.Logf("Remove: %v", err)
	} else {
		t.Fatalf("Remove not passed")
	}

	if err := rw.Remove(5, 10); err != nil {
		t.Fatalf("Remove not passed, Remove: %v", err)
	}

	res, err := rw.Seek(5, 5)
	if err != nil {
		t.Fatalf("Remove not passed, Seek: %v", err)
	}

	if len(res) > 0 {
		t.Fatalf("Remove not passed, Remove not successfully")
	}

	t.Logf("Rmove test passed")
}

func TestQueueImpl_Seek(t *testing.T) {
	once.Do(testInit)

	// seek 80 ~ 90
	res, err := rw.Seek(80, 10)
	if err != nil {
		t.Fatalf("Seek: %v", err)
	}

	for i, r := range res {
		t.Logf(" %d : %v", i, r)
	}

	// 重复seek 80~90
	res, err = rw.Seek(80, 10)
	if err != nil {
		t.Fatalf("Seek: %v", err)
	}

	for i, r := range res {
		t.Logf(" %d : %v", i, r)
	}
}

func TestQueueImpl_Single(t *testing.T) {
	once.Do(testInit)

	for i := 0; i < 100; i++ {
		<-rw.Single()
		if v, err := rw.Next(true); err != nil {
			t.Fatalf("Next: %v", err)
		} else {
			t.Logf("Value: %v", v)
		}
	}

	if res, err := rw.Seek(0, 99); err != nil {
		t.Fatalf("Seek: %v", err)
	} else if len(res) != 0 {
		t.Fatalf("Seek Result Number not nil")
	}
}
