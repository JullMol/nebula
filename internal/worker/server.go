package worker

import (
	"context"
	"fmt"
	pb "github.com/JullMol/nebula/api/pb"
	"github.com/JullMol/nebula/internal/platform/docker"
)

type Server struct {
	pb.UnimplementedWorkerServiceServer
	dockerClient *docker.Client
}

func NewServer(dockerCli *docker.Client) *Server {
	return &Server{
		dockerClient: dockerCli,
	}
}

func (s *Server) StartContainer(ctx context.Context, req *pb.StartContainerRequest) (*pb.StartContainerResponse, error) {
	fmt.Printf("🔔 Request Masuk: Start Container image=%s cmd=%s\n", req.Image, req.Command)

	cmd := []string{req.Command} 
    if req.Command == "" {
        cmd = []string{"echo", "Hello from gRPC!"}
    } else {
        cmd = []string{"sh", "-c", req.Command}
    }

	containerID, err := s.dockerClient.RunContainer(ctx, req.Image, cmd, nil)
	if err != nil {
		fmt.Printf("❌ Gagal start: %v\n", err)
		return &pb.StartContainerResponse{
			Success:      false,
			ErrorMessage: err.Error(),
		}, nil
	}

	fmt.Printf("✅ Sukses! ID: %s\n", containerID)

	return &pb.StartContainerResponse{
		Success:     true,
		ContainerId: containerID,
	}, nil
}

func (s *Server) StopContainer(ctx context.Context, req *pb.StopContainerRequest) (*pb.StopContainerResponse, error) {
	fmt.Printf("🔔 Request Masuk: Stop Container ID=%s\n", req.ContainerId)

	err := s.dockerClient.StopContainer(ctx, req.ContainerId)
	if err != nil {
		return &pb.StopContainerResponse{
			Success:      false,
			ErrorMessage: err.Error(),
		}, nil
	}

	return &pb.StopContainerResponse{
		Success: true,
	}, nil
}

func (s *Server) GetLogs(ctx context.Context, req *pb.GetLogsRequest) (*pb.GetLogsResponse, error) {
	logs, err := s.dockerClient.GetLogs(ctx, req.ContainerId)
	if err != nil {
		return &pb.GetLogsResponse{
			ErrorMessage: err.Error(),
		}, nil
	}

	return &pb.GetLogsResponse{
		Logs: logs,
	}, nil
}