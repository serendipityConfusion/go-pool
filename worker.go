package go_pool

type Worker interface {
	Run()
	Send(job Job)
	Close()
}

type DefaultWorker struct {
	jobChan chan Job // 工作协程通道
	pool    *Pool
}

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
			// todo 后续增加超时功能
			if job != nil {
				job()
			}
			w.pool.WorkerChan <- w // 将工作协程放回池中
			//case <-ticker.C: // 长时间不使用回收部分协程
			//	w.Close()
		}
	}
}

func (w *DefaultWorker) Close() {
	close(w.jobChan) // 关闭工作协程通道
}

func NewDefaultWorker(pool *Pool) *DefaultWorker {
	worker := &DefaultWorker{
		jobChan: make(chan Job, 1), // 初始化工作协程通道
		pool:    pool,
	}
	go worker.Run()
	return worker
}
