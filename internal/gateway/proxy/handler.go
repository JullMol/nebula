package proxy

import (
	"context"
	"fmt"
	"log"
	"time"

	pb "github.com/JullMol/nebula/api/pb"
	"github.com/JullMol/nebula/internal/orchestrator/scheduler"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

type ProxyService struct {
	scheduler scheduler.LoadBalancer
	workers   []string
}

func NewProxyService(lb scheduler.LoadBalancer, workers []string) *ProxyService {
	return &ProxyService{
		scheduler: lb,
		workers:   workers,
	}
}

func (s *ProxyService) ForwardRunRequest(ctx context.Context, image, command, code string) (*pb.StartContainerResponse, error) {
	workerAddress := s.scheduler.NextWorker(s.workers)
	fmt.Printf("🔀 [Proxy] Forwarding to: %s\n", workerAddress)

	conn, err := grpc.NewClient(workerAddress, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		log.Printf("❌ Gagal connect ke worker %s: %v", workerAddress, err)
		return nil, err
	}
	defer conn.Close()

	client := pb.NewWorkerServiceClient(conn)

	ctx, cancel := context.WithTimeout(ctx, 30*time.Second)
	defer cancel()

	return client.StartContainer(ctx, &pb.StartContainerRequest{
		Image:   image,
		Command: command,
		Code:    code,
	})
}

func (s *ProxyService) ForwardWaitRequest(ctx context.Context, containerID string) error {
	for _, w := range s.workers {
		conn, err := grpc.NewClient(w, grpc.WithTransportCredentials(insecure.NewCredentials()))
		if err == nil {
			client := pb.NewWorkerServiceClient(conn)
			ctx, cancel := context.WithTimeout(ctx, 60*time.Second)
			_, err := client.WaitContainer(ctx, &pb.WaitContainerRequest{ContainerId: containerID})
			cancel()
			conn.Close()
			if err == nil {
				return nil
			}
		}
	}
	return fmt.Errorf("wait failed on all workers")
}

func (s *ProxyService) ForwardLogRequest(ctx context.Context, containerID string) (string, error) {
	for _, w := range s.workers {
		conn, err := grpc.NewClient(w, grpc.WithTransportCredentials(insecure.NewCredentials()))
		if err == nil {
			client := pb.NewWorkerServiceClient(conn)
			resp, err := client.GetLogs(ctx, &pb.GetLogsRequest{ContainerId: containerID})
			conn.Close()
			if err == nil {
				return resp.Logs, nil
			}
		}
	}
	return "", fmt.Errorf("logs not found")
}