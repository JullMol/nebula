package proxy

import (
	"context"
	"fmt"

	pb "github.com/JullMol/nebula/api/pb"
	"github.com/JullMol/nebula/internal/orchestrator/scheduler"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

type ProxyService struct {
	scheduler *scheduler.RoundRobin
	workers   []string
}

func NewProxyService(sched *scheduler.RoundRobin, workers []string) *ProxyService {
	return &ProxyService{
		scheduler: sched,
		workers:   workers,
	}
}

func (p *ProxyService) ForwardRunRequest(ctx context.Context, image, command string) (*pb.StartContainerResponse, error) {
	target := p.scheduler.NextWorker(p.workers)
	if target == "" {
		return nil, fmt.Errorf("no workers available")
	}

	fmt.Printf("🔀 [Proxy] Forwarding to: %s\n", target)

	conn, err := grpc.NewClient(target, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		return nil, fmt.Errorf("failed to connect to worker %s: %w", target, err)
	}
	defer conn.Close()

	client := pb.NewWorkerServiceClient(conn)

	return client.StartContainer(ctx, &pb.StartContainerRequest{
		Image:   image,
		Command: command,
	})
}

func (p *ProxyService) ForwardLogRequest(ctx context.Context, containerID string) (string, error) {
	target := p.workers[0]
	conn, err := grpc.NewClient(target, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		return "", err
	}
	defer conn.Close()

	client := pb.NewWorkerServiceClient(conn)
	resp, err := client.GetLogs(ctx, &pb.GetLogsRequest{ContainerId: containerID})
	if err != nil {
		return "", err
	}
	return resp.Logs, nil
}