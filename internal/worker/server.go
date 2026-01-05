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

func NewServer(dockerClient *docker.Client) *Server {
	return &Server{dockerClient: dockerClient}
}

func (s *Server) StartContainer(ctx context.Context, req *pb.StartContainerRequest) (*pb.StartContainerResponse, error) {
	fmt.Printf("🚀 Request Masuk: Image=%s | CodeLength=%d\n", req.Image, len(req.Code))

	containerID, err := s.dockerClient.RunContainer(ctx, req.Image, req.Command, req.Code)
	
	if err != nil {
		return nil, err
	}

	return &pb.StartContainerResponse{
		ContainerId: containerID,
	}, nil
}

func (s *Server) StopContainer(ctx context.Context, req *pb.StopContainerRequest) (*pb.StopContainerResponse, error) {
	err := s.dockerClient.StopContainer(ctx, req.ContainerId)
	if err != nil {
		return &pb.StopContainerResponse{Success: false}, err
	}
	return &pb.StopContainerResponse{Success: true}, nil
}

func (s *Server) WaitContainer(ctx context.Context, req *pb.WaitContainerRequest) (*pb.WaitContainerResponse, error) {
	err := s.dockerClient.WaitContainer(ctx, req.ContainerId)
	if err != nil {
		return &pb.WaitContainerResponse{Success: false}, err
	}
	return &pb.WaitContainerResponse{Success: true}, nil
}

func (s *Server) GetLogs(ctx context.Context, req *pb.GetLogsRequest) (*pb.GetLogsResponse, error) {
	logs, err := s.dockerClient.GetLogs(ctx, req.ContainerId)
	if err != nil {
		return nil, err
	}
	return &pb.GetLogsResponse{Logs: logs}, nil
}