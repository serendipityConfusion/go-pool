package go_pool

import (
	"sync/atomic"
	"time"
)

type Pool struct {
	minGoroutine     int64              // 常驻的协程数量
	capacity         int64              // 最大协程数量
	workerChan       chan Worker        // 工作协程通道
	curGoroutine     int64              // 当前协程数量
	closeExpireTime  time.Duration      // 关闭的超时时间，单位为秒,默认3秒,如果小于等于0则不超时
	state            int32              // 池的状态
	createWorkerFunc func(*Pool) Worker // 创建工作协程的函数
	idleTimeout      time.Duration      // 空闲协程的超时时间,默认3秒
	releaseCount     int                // 单次检测协程数量
	reclaimerTime    time.Duration      // 回收协程的超时时间，默认10秒
}

type PoolOption func(*Pool)

// WithCloseExpireTime 设置关闭池的超时时间
func WithCloseExpireTime(expireTime time.Duration) PoolOption {
	return func(p *Pool) {
		p.closeExpireTime = expireTime
	}
}

// WithCreateWorkerFunc 设置创建工作协程的函数
func WithCreateWorkerFunc(createWorkerFunc func(*Pool) Worker) PoolOption {
	return func(p *Pool) {
		p.createWorkerFunc = createWorkerFunc // 设置自定义创建工作协程的函数
	}
}

// WithIdleTimeout 设置空闲协程超时时间
func WithIdleTimeout(idleTimeout time.Duration) PoolOption {
	return func(p *Pool) {
		p.idleTimeout = idleTimeout
	}
}

// WithReclaimerTime 设置回收协程的超时时间
func WithReclaimerTime(reclaimerTime time.Duration) PoolOption {
	return func(p *Pool) {
		p.reclaimerTime = reclaimerTime
	}
}

func WithReleaseContinueCount(releaseCount int) PoolOption {
	return func(p *Pool) {
		p.releaseCount = releaseCount
	}
}

func NewPool(minGoroutine, capacity int64, options ...PoolOption) *Pool {
	pool := &Pool{
		minGoroutine:     minGoroutine,
		capacity:         capacity,
		workerChan:       make(chan Worker, capacity),
		closeExpireTime:  3,
		state:            Running,
		createWorkerFunc: newDefaultWorker,
		idleTimeout:      3 * time.Second,  // 默认空闲协程超时时间为3秒
		reclaimerTime:    10 * time.Second, // 默认回收协程的超时时间为10秒
		releaseCount:     int(capacity-minGoroutine) / 3,
	}
	// 应用配置选项
	for _, option := range options {
		option(pool)
	}
	for i := int64(0); i < minGoroutine; i++ {
		worker := pool.createWorkerFunc(pool)
		pool.workerChan <- worker // 将工作协程放入工作协程通道
	}
	pool.curGoroutine = minGoroutine // 初始化当前协程数量
	go pool.reclaimer()
	return pool
}

func (p *Pool) Do(job Job) error {
	if p.IsClosed() {
		return ErrPoolClosed // 如果池已关闭，返回错误
	}
	if job == nil {
		return ErrJobNil
	}
	select {
	case worker := <-p.workerChan:
		worker.Send(job)
	default:
		if atomic.LoadInt64(&p.curGoroutine) < p.capacity {
			atomic.AddInt64(&p.curGoroutine, 1)
			worker := newDefaultWorker(p)
			worker.Send(job)
		} else {
			// 阻塞直到有 worker 可用
			worker := <-p.workerChan
			worker.Send(job)
		}
	}
	return nil
}

func (p *Pool) Close() {
	if !atomic.CompareAndSwapInt32(&p.state, Running, Closed) { // 设置池的状态为 Closed
		return
	}
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
	if 0 >= p.closeExpireTime {
		<-finishClose // 等待所有工作协程完成
		close(p.workerChan)
		return
	}
	select {
	case <-finishClose:
		close(p.workerChan) // 关闭工作协程通道
	case <-time.After(p.closeExpireTime):
		// 超时后强制关闭
		return
	}
}

func (p *Pool) IsClosed() bool {
	return atomic.LoadInt32(&p.state) == Closed
}

// 回收长时间空闲的worker
func (p *Pool) reclaimer() {
	// 不回收
	if p.reclaimerTime <= 0 || p.releaseCount <= 0 {
		return
	}
	ticker := time.NewTicker(p.reclaimerTime)
	defer ticker.Stop()
	for {
		<-ticker.C
		if p.IsClosed() {
			return
		}
		if len(p.workerChan) == 0 {
			continue
		}
	Next:
		count := 0
		for i := 0; i < min(p.releaseCount, len(p.workerChan)); i++ {
			worker := <-p.workerChan
			if p.curGoroutine > p.minGoroutine && time.Now().Sub(worker.ActiveTime()) > p.idleTimeout {
				worker.Close()
				atomic.AddInt64(&p.curGoroutine, -1)
				count++
				continue
			}
			p.workerChan <- worker
		}
		if count > p.releaseCount/2 {
			goto Next
		}
		ticker.Reset(p.reclaimerTime)
	}
}
