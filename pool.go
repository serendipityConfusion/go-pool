package go_pool

import (
	"sync/atomic"
	"time"
)

type Pool struct {
	minGoroutine int64       // 常驻的协程数量
	capacity     int64       // 最大协程数量
	workerChan   chan Worker // 工作协程通道
	jobChan      chan Job    // 任务通道, 默认缓冲区大小为100
	curGoroutine int64       // 当前协程数量
	expireTime   int32       // 关闭的超时时间，单位为秒,默认3秒,如果小于等于0则不超时
	state        int32       // 池的状态
}

type PoolOption func(*Pool)

func WithExpireTime(expireTime int32) PoolOption {
	return func(p *Pool) {
		p.expireTime = expireTime
	}
}

func WithJobChanCapacity(length int) PoolOption {
	return func(p *Pool) {
		if length < 0 {
			p.jobChan = make(chan Job, 100) // 默认缓冲区大小为100
			return
		}
		p.jobChan = make(chan Job, length) // 设置任务通道的缓冲区大小
	}
}

func NewPool(minGoroutine, capacity int64, options ...PoolOption) *Pool {
	pool := &Pool{
		minGoroutine: minGoroutine,
		capacity:     capacity,
		workerChan:   make(chan Worker, capacity),
		jobChan:      make(chan Job, 100), // 任务通道，缓冲区大小为100
		expireTime:   3,
		state:        Running,
	}
	// 应用配置选项
	for _, option := range options {
		option(pool)
	}
	for i := int64(0); i < minGoroutine; i++ {
		worker := NewDefaultWorker(pool)
		pool.workerChan <- worker // 将工作协程放入工作协程通道
	}
	pool.curGoroutine = minGoroutine // 初始化当前协程数量
	go pool.run()
	return pool
}

func (p *Pool) Do(job Job) error {
	if atomic.LoadInt32(&p.state) == Closed {
		return ErrPoolClosed // 如果池已关闭，返回错误
	}
	p.jobChan <- job // 将任务放入任务通道
	return nil
}

func (p *Pool) run() {
	for job := range p.jobChan {
		if job == nil {
			continue
		}
		select {
		case worker := <-p.workerChan:
			worker.Send(job)
		default:
			if atomic.LoadInt64(&p.curGoroutine) < p.capacity {
				atomic.AddInt64(&p.curGoroutine, 1)
				worker := NewDefaultWorker(p)
				worker.Send(job)
			} else {
				// 阻塞直到有 worker 可用
				worker := <-p.workerChan
				worker.Send(job)
			}
		}
	}
}

func (p *Pool) Close() {
	atomic.CompareAndSwapInt32(&p.state, Running, Closed) // 设置池的状态为 Closed
	//close(p.jobChan) // 不关闭，避免panic的问题，设计上可以容忍
	finishClose := make(chan struct{})
	go func() {
		for p.curGoroutine > 0 {
			worker := <-p.workerChan
			worker.Close()
			_ = atomic.AddInt64(&p.curGoroutine, -1)
		}
		close(finishClose)
	}()
	// 无超时时间
	if 0 >= p.expireTime {
		<-finishClose // 等待所有工作协程完成
		close(p.workerChan)
		return
	}
	select {
	case <-finishClose:
		close(p.workerChan) // 关闭工作协程通道
	case <-time.After(time.Duration(p.expireTime) * time.Second):
		// 超时后强制关闭
		return
	}
}
