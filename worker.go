package go_pool

import (
	"log"
	"os"
	"runtime/debug"
	"time"
)

type Worker interface {
	Send(job Job)
	Close()
	ActiveTime() time.Time // 活跃时间
}

type DefaultWorker struct {
	jobChan    chan Job // 工作协程通道
	pool       *Pool
	activeTime time.Time // 活跃时间
	logger     *log.Logger
}

func (w *DefaultWorker) ActiveTime() time.Time {
	return w.activeTime
}

var _ Worker = (*DefaultWorker)(nil)

func (w *DefaultWorker) Send(job Job) {
	w.jobChan <- job // 将任务发送到工作协程
}

func (w *DefaultWorker) run() {
	defer func() {
		if p := recover(); p != nil {
			// 捕获异常，防止工作协程崩溃
			w.logger.Println("worker exits from panic: %v\n%s\n", p, debug.Stack())
			go w.run() // 重新启动工作协程
		}
	}()
	for {
		select {
		case job, ok := <-w.jobChan:
			if !ok {
				return
			}
			if job != nil {
				job()
			}
			w.activeTime = time.Now() // 更新活跃时间
			w.pool.workerChan <- w    // 将工作协程放回池中
		}
	}
}

func (w *DefaultWorker) Close() {
	close(w.jobChan) // 关闭工作协程通道
}

func newDefaultWorker(pool *Pool) Worker {
	worker := &DefaultWorker{
		jobChan:    make(chan Job, 1), // 初始化工作协程通道
		pool:       pool,
		activeTime: time.Now(),
		logger:     log.New(os.Stderr, "[worker] ", log.LstdFlags),
	}
	go worker.run()
	return worker
}
