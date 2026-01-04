package scheduler

type SchedulerPolicy interface {
	NextWorker(availableWorkers []string) string
}