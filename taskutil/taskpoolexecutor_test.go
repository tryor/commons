package taskutil

import (
	"fmt"
	"testing"
	"time"
)

func Test(t *testing.T) {
	executor := NewTaskPoolExecutor(3, 100)
	executor.Start()

	for i := 0; i < 10; i++ {
		executor.Execute(func(p ...interface{}) {
			fmt.Printf("test ...., %v\n", p...)
			time.Sleep(time.Millisecond)
		}, i)
	}

	time.Sleep(time.Second)

	executor.CloseAndWait()
}
