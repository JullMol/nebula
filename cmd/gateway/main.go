package main

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/gofiber/fiber/v2"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"

	pb "github.com/JullMol/nebula/api/pb"
)

func main() {
	workerAddr := "localhost:9090"
	conn, err := grpc.NewClient(workerAddr, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		log.Fatalf("❌ Gagal connect ke worker: %v", err)
	}
	defer conn.Close()

	workerClient := pb.NewWorkerServiceClient(conn)
	fmt.Printf("✅ Gateway terhubung ke Worker di %s\n", workerAddr)

	app := fiber.New()

	app.Post("/run", func(c *fiber.Ctx) error {
		type RequestPayload struct {
			Image   string `json:"image"`
			Command string `json:"command"`
		}

		var payload RequestPayload
		if err := c.BodyParser(&payload); err != nil {
			return c.Status(400).JSON(fiber.Map{"error": "Invalid JSON"})
		}

		if payload.Image == "" { payload.Image = "alpine" }
		if payload.Command == "" { payload.Command = "echo 'Hello from REST API!'" }

		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()

		fmt.Printf("🌍 HTTP POST /run: %s\n", payload.Image)

		resp, err := workerClient.StartContainer(ctx, &pb.StartContainerRequest{
			Image:   payload.Image,
			Command: payload.Command,
		})

		if err != nil {
			return c.Status(500).JSON(fiber.Map{"error": err.Error()})
		}

		return c.JSON(fiber.Map{
			"status":       "success",
			"container_id": resp.ContainerId,
			"message":      "Container berhasil dijalankan! Cek logs untuk outputnya.",
		})
	})

	app.Get("/logs/:id", func(c *fiber.Ctx) error {
		containerID := c.Params("id")

		fmt.Printf("🌍 HTTP GET /logs: %s\n", containerID)

		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()

		resp, err := workerClient.GetLogs(ctx, &pb.GetLogsRequest{
			ContainerId: containerID,
		})

		if err != nil {
			return c.Status(500).JSON(fiber.Map{"error": err.Error()})
		}

		return c.SendString(resp.Logs)
	})

	log.Fatal(app.Listen(":3000"))
}