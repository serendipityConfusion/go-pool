package go_pool

import (
	"bytes"
	"fmt"
	"runtime"
	"strconv"
	"sync"
	"testing"
	"time"
)

func GetGID() uint64 {
	// 分配一个足够大的缓冲区
	b := make([]byte, 64)
	b = b[:runtime.Stack(b, false)]
	// b 示例："goroutine 10 [running]:\n"
	b = bytes.TrimPrefix(b, []byte("goroutine "))
	i := bytes.IndexByte(b, ' ')
	if i < 0 {
		return 0
	}
	gid, err := strconv.ParseUint(string(b[:i]), 10, 64)
	if err != nil {
		return 0
	}
	return gid
}

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
		_ = pool.Do(job)
	}
	// 模拟等待所有任务完成
	sp.Wait()
	return
}

func TestGoPoolClose(t *testing.T) {
	pool := NewPool(3, 20, WithExpireTime(1))
	for i := 0; i < 1000; i++ {
		val := i
		job := func() {
			fmt.Println("Hello, World!", val, "GID:", GetGID())
			time.Sleep(100 * time.Millisecond)
		}
		_ = pool.Do(job)
	}
	fmt.Println("before close curWorkChan", len(pool.workerChan), "curGoroutine", pool.curGoroutine)
	pool.Close()
	fmt.Println("after close curWorkChan", len(pool.workerChan), "curGoroutine", pool.curGoroutine)
}
