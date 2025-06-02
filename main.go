package go_pool

import (
	"bytes"
	"fmt"
	"os"
	"os/signal"
	"runtime"
	"strconv"
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

func main() {
	pool := NewPool(3, 20)
	for i := 0; i < 1000; i++ {
		val := i
		job := func() {
			time.Sleep(2 * time.Second)
			fmt.Println("Hello, World!", val, "GID:", GetGID())
		}
		pool.Do(job)
	}
	// 模拟等待所有任务完成
	stop := make(chan os.Signal)
	signal.Notify(stop, os.Interrupt)
	<-stop
	return
}
