package main

import (
	"context"
	"fmt"
	"log"
	"time"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"

	pb "github.com/JullMol/nebula/api/pb"
)

func main() {
	fmt.Println("🧠 Nebula Orchestrator Starting...")

	target := "localhost:9090"
	conn, err := grpc.NewClient(target, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		log.Fatalf("❌ Gagal connect ke worker: %v", err)
	}
	defer conn.Close()

	client := pb.NewWorkerServiceClient(conn)
	fmt.Printf("✅ Terhubung ke Worker di %s\n", target)

	fmt.Println("📤 Mengirim perintah StartContainer...")

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	resp, err := client.StartContainer(ctx, &pb.StartContainerRequest{
		Image:   "alpine",
		Command: "echo 'Halo Bos! Perintah diterima via gRPC!'",
	})

	if err != nil {
		log.Fatalf("❌ Gagal memanggil RPC: %v", err)
	}

	fmt.Printf("✅ Worker merespon: ContainerID=%s\n", resp.ContainerId)
}