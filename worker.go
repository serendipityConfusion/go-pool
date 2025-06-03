package go_pool

import "time"

type Worker interface {
	Send(job Job)
	Close()
	ActiveTime() time.Time // 活跃时间
}

type DefaultWorker struct {
	jobChan    chan Job // 工作协程通道
	pool       *Pool
	activeTime time.Time // 活跃时间
}

func (w *DefaultWorker) ActiveTime() time.Time {
	return w.activeTime
}

var _ Worker = (*DefaultWorker)(nil)

func (w *DefaultWorker) Send(job Job) {
	w.jobChan <- job // 将任务发送到工作协程
}

func (w *DefaultWorker) Run() {
	//ticker := time.NewTicker(5 * time.Second)
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
	}
	go worker.Run()
	return worker
}
