package go_pool

import "sync/atomic"

type Pool struct {
	minGoroutine int64       // 常驻的协程数量
	capacity     int64       // 最大协程数量
	WorkerChan   chan Worker // 工作协程通道
	jobChan      chan Job    // 任务通道
	curGoroutine int64       // 当前协程数量
}

func NewPool(minGoroutine, capacity int64) *Pool {
	pool := &Pool{
		minGoroutine: minGoroutine,
		capacity:     capacity,
		WorkerChan:   make(chan Worker, capacity),
		jobChan:      make(chan Job, 100), // 任务通道，缓冲区大小为100
	}
	for i := int64(0); i < minGoroutine; i++ {
		worker := NewDefaultWorker(pool)
		pool.WorkerChan <- worker // 将工作协程放入工作协程通道
	}
	pool.curGoroutine = minGoroutine // 初始化当前协程数量
	go pool.run()
	return pool
}

func (p *Pool) Do(job Job) {
	p.jobChan <- job // 将任务放入任务通道
}

func (p *Pool) run() {
	for job := range p.jobChan {
		if job == nil {
			continue
		}
		select {
		case worker := <-p.WorkerChan:
			worker.Send(job)
		default:
			if atomic.LoadInt64(&p.curGoroutine) < p.capacity {
				atomic.AddInt64(&p.curGoroutine, 1)
				worker := NewDefaultWorker(p)
				worker.Send(job)
			} else {
				// 阻塞直到有 worker 可用
				worker := <-p.WorkerChan
				worker.Send(job)
			}
		}
	}
}

func (p *Pool) Close() {

}
