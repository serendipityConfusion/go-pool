package go_pool

import (
	"fmt"
	"sync"
	"testing"
	"time"
)

func TestGoPool(t *testing.T) {
	pool := NewPool(3, 20)
	sp := sync.WaitGroup{}
	for i := 0; i < 1000; i++ {
		val := i
		sp.Add(1)
		job := func() {
			defer sp.Done()
			fmt.Println("Hello, World!", val, "GID:", GetGID())
			time.Sleep(100 * time.Millisecond)
		}
		pool.Do(job)
	}
	// 模拟等待所有任务完成
	sp.Wait()
	return
}
