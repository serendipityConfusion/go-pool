package go_pool

var (
	ErrNotAvailableWorker = &Error{Code: 1, Message: "No available worker to handle the job"}
)
