package main

import (
	"flag"
	"fmt"
	"log"
	"net"

	"google.golang.org/grpc"

	pb "github.com/JullMol/nebula/api/pb"
	"github.com/JullMol/nebula/internal/platform/docker"
	"github.com/JullMol/nebula/internal/worker"
)

func main() {
	portPtr := flag.String("port", "9090", "Port untuk Worker")
	flag.Parse()

	port := fmt.Sprintf(":%s", *portPtr)

	fmt.Printf("⚡ Nebula Worker Node Starting on Port %s...\n", port)

	dockerCli, err := docker.NewClient()
	if err != nil {
		log.Fatalf("❌ Gagal konek Docker: %v", err)
	}
	
	lis, err := net.Listen("tcp", port)
	if err != nil {
		log.Fatalf("❌ Gagal listen port %s: %v", port, err)
	}

	grpcServer := grpc.NewServer()
	workerServer := worker.NewServer(dockerCli)
	pb.RegisterWorkerServiceServer(grpcServer, workerServer)

	fmt.Printf("🚀 Worker siap di %s\n", port)

	if err := grpcServer.Serve(lis); err != nil {
		log.Fatalf("❌ Gagal serve gRPC: %v", err)
	}
}