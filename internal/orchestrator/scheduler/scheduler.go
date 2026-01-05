package scheduler

import "sync/atomic"

type LoadBalancer interface {
	NextWorker(workers []string) string
}
type RoundRobin struct {
	counter uint64
}

func NewRoundRobin() *RoundRobin {
	return &RoundRobin{
		counter: 0,
	}
}

func (r *RoundRobin) NextWorker(workers []string) string {
	if len(workers) == 0 {
		return ""
	}
	current := atomic.AddUint64(&r.counter, 1)
	index := (current - 1) % uint64(len(workers))
	return workers[index]
}