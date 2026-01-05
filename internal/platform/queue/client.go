package queue

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/redis/go-redis/v9"
)

type Job struct {
	ID      string `json:"id"`
	Image   string `json:"image"`
	Command string `json:"command"`
	Code    string `json:"code"`
}

type QueueSystem interface {
	Enqueue(ctx context.Context, job Job) error
	Dequeue(ctx context.Context) (*Job, error)
	SetResult(ctx context.Context, jobID string, result string) error
	GetResult(ctx context.Context, jobID string) (string, error)
}

type RedisQueue struct {
	client *redis.Client
	queueName string
}

func NewRedisQueue(addr string) *RedisQueue {
	rdb := redis.NewClient(&redis.Options{
		Addr: addr,
	})
	return &RedisQueue{
		client: rdb,
		queueName: "nebula_jobs",
	}
}

func (r *RedisQueue) Enqueue(ctx context.Context, job Job) error {
	data, _ := json.Marshal(job)
	return r.client.RPush(ctx, r.queueName, data).Err()
}

func (r *RedisQueue) Dequeue(ctx context.Context) (*Job, error) {
	val, err := r.client.BLPop(ctx, 0*time.Second, r.queueName).Result()
	if err != nil {
		return nil, err
	}

	var job Job
	if err := json.Unmarshal([]byte(val[1]), &job); err != nil {
		return nil, err
	}
	return &job, nil
}

func (r *RedisQueue) SetResult(ctx context.Context, jobID string, result string) error {
	key := fmt.Sprintf("result:%s", jobID)
	return r.client.Set(ctx, key, result, 10*time.Minute).Err()
}

func (r *RedisQueue) GetResult(ctx context.Context, jobID string) (string, error) {
	key := fmt.Sprintf("result:%s", jobID)
	val, err := r.client.Get(ctx, key).Result()
	if err == redis.Nil {
		return "", fmt.Errorf("pending")
	}
	if err != nil {
		return "", err
	}
	return val, nil
}