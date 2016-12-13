package taskutil

import (
	"fmt"
	"testing"
	"time"
)

func Test(t *testing.T) {
	executor := NewTaskPoolExecutor(3, 0)
	executor.Start()

	for i := 0; i < 10; i++ {
		executor.Execute(func(p ...interface{}) {
			fmt.Printf("a test ...., %v\n", p...)
			time.Sleep(time.Millisecond * 100)
		}, i)
	}

	time.Sleep(time.Second * 1)

	for i := 0; i < 10; i++ {
		executor.Execute(func(p ...interface{}) {
			fmt.Printf("b test ...., %v\n", p...)
			time.Sleep(time.Millisecond * 100)
		}, i)
	}

	time.Sleep(time.Second * 1)

	executor.Shutdown()
}
