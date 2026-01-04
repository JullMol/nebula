package main

import (
	"fmt"
	"log"
	"net"

	"google.golang.org/grpc"

	pb "github.com/JullMol/nebula/api/pb"
	"github.com/JullMol/nebula/internal/platform/docker"
	"github.com/JullMol/nebula/internal/worker"
)

func main() {
	fmt.Println("⚡ Nebula Worker Node Starting...")

	dockerCli, err := docker.NewClient()
	if err != nil {
		log.Fatalf("❌ Gagal konek Docker: %v", err)
	}
	fmt.Println("✅ Docker Connected")

	port := ":9090"
	lis, err := net.Listen("tcp", port)
	if err != nil {
		log.Fatalf("❌ Gagal listen port %s: %v", port, err)
	}

	grpcServer := grpc.NewServer()
	workerServer := worker.NewServer(dockerCli)
	pb.RegisterWorkerServiceServer(grpcServer, workerServer)

	fmt.Printf("🚀 Worker siap menerima perintah di port %s...\n", port)

	if err := grpcServer.Serve(lis); err != nil {
		log.Fatalf("❌ Gagal serve gRPC: %v", err)
	}
}