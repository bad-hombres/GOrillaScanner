package scanner

import (
  "sync"
  "sync/atomic"
  "runtime"
  "time"
  //"GOrillaScanner/log"
)
type Scheduler struct {
  threads int
  threadCount uint64
  wg sync.WaitGroup
}

func (s *Scheduler) StartThread(job func()) {
  atomic.AddUint64(&s.threadCount, 1)

  s.wg.Add(1)
  go func() {
    defer s.StopThread()
    job()
  }()
  s.WaitForThreads(false)
}

func (s *Scheduler) WaitForThreads(finish bool) {
  if atomic.LoadUint64(&s.threadCount) >= uint64(s.threads) {
    for atomic.LoadUint64(&s.threadCount) >= uint64(s.threads)  {
      time.Sleep(10 * time.Millisecond)
    }
  }
  if finish {
    s.wg.Wait()
  }
}

func (s *Scheduler) StopThread() {
  atomic.AddUint64(&s.threadCount, ^uint64(0))
  s.wg.Done()
}

func (s *Scheduler) GetThreadCount() uint64 {
  return atomic.LoadUint64(&s.threadCount)
}

func NewScheduler(threads int) *Scheduler {
  runtime.GOMAXPROCS(threads)
  return &Scheduler{threads: threads, wg: sync.WaitGroup{}, threadCount: 0}
}
