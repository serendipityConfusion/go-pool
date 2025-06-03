# 简介
golang实现的简单协程池

## TODO
- 当worker 异常退出时，pool无法感知，导致close可能被阻塞(curGoroutine的数量可能又问题)
  - worker 和 pool 之间的状态同步问题
