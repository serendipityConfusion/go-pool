package go_pool

var (
	ErrNotAvailableWorker = &Error{Code: 1, Message: "No available worker to handle the job"}
	ErrPoolClosed         = &Error{Code: 2, Message: "Pool is closed and cannot accept new jobs"}
	ErrJobNil             = &Error{Code: 3, Message: "Job is nil"}
)

// pool state
const (
	Running = iota + 1 // Running state
	Closed             // Closed state
)
