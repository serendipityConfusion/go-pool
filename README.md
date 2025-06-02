# 简介
golang实现的简单协程池

## TODO
- 当worker 异常退出时，pool无法感知，导致close可能被阻塞
  - woker 和 pool 之间的状态同步问题