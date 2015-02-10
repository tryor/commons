package event

import (
	//"bytes"
	"fmt"
	"testing"
	//"time"
)

type TestListener struct {
	A int
}

func (l *TestListener) HandleEvent(e IEvent) bool {
	println("handle event ***************** ", e.GetType())
	return false
}

func Test(t *testing.T) {
	d := NewDispatcher()
	d.RegisterListener(1, &TestListener{})
	fmt.Println(d.AllTypeListeners)
	d.RegisterListener(1, &TestListener{})
	fmt.Println(d.AllTypeListeners)
	d.RegisterListener(1, &TestListener{})
	fmt.Println(d.AllTypeListeners)
	d.RegisterListener(1, &TestListener{}, 0)
	fmt.Println(d.AllTypeListeners)
	d.RegisterListener(1, &TestListener{}, 10)
	fmt.Println(d.AllTypeListeners)
	d.RegisterListener(1, &TestListener{}, 5)
	fmt.Println(d.AllTypeListeners)
	d.RegisterListener(1, &TestListener{}, 5)
	fmt.Println(d.AllTypeListeners)
	d.RegisterListener(1, &TestListener{}, 1)
	fmt.Println(d.AllTypeListeners)

	d.RegisterListener(2, &TestListener{})
	fmt.Println(d.AllTypeListeners)
	l := &TestListener{}
	d.RegisterListener(2, l)
	fmt.Println(d.AllTypeListeners)
	d.RegisterListener(2, &TestListener{})
	fmt.Println(d.AllTypeListeners)
	d.RegisterListener(2, l)
	fmt.Println(d.AllTypeListeners)
	d.RegisterListener(2, &TestListener{})
	fmt.Println(d.AllTypeListeners)

	d.RemoveListener(l)
	fmt.Println(d.AllTypeListeners)

	d.RegisterListener(ZERO_TYPE, &TestListener{})
	d.RegisterListener(ZERO_TYPE, &TestListener{})

	d.FireEvent(NewEvent(2, nil))
}
