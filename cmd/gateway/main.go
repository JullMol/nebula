package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
	"github.com/prometheus/client_golang/prometheus/promhttp"

	"github.com/JullMol/nebula/internal/gateway/proxy"
	"github.com/JullMol/nebula/internal/orchestrator/scheduler"
	"github.com/JullMol/nebula/internal/platform/queue"
	"github.com/JullMol/nebula/pkg/config"
)

func main() {
	cfg, _ := config.LoadConfig()
	lb := scheduler.NewRoundRobin()
	proxySvc := proxy.NewProxyService(lb, cfg.Server.Workers)

	q := queue.NewRedisQueue(cfg.Server.RedisAddr)
	fmt.Println("✅ Connected to Redis Queue")

	go func() {
		http.Handle("/metrics", promhttp.Handler())
		fmt.Println("📊 Metrics server running on :3001")
		if err := http.ListenAndServe(":3001", nil); err != nil {
			log.Printf("❌ Metrics server error: %v", err)
		}
	}()

	go func() {
		fmt.Println("🚜 Background Dispatcher Started...")
		for {
			ctx := context.Background()
			
			job, err := q.Dequeue(ctx)
			if err != nil {
				fmt.Printf("❌ Queue Error: %v\n", err)
				time.Sleep(1 * time.Second)
				continue
			}

			fmt.Printf("🚜 Processing Job ID: %s (Image: %s)\n", job.ID, job.Image)
			
			resp, err := proxySvc.ForwardRunRequest(ctx, job.Image, job.Command)
			
			result := ""
			if err != nil {
				result = fmt.Sprintf("Error: %v", err)
			} else {
				logs, _ := proxySvc.ForwardLogRequest(ctx, resp.ContainerId)
				result = logs
			}
			q.SetResult(ctx, job.ID, result)
			fmt.Printf("✅ Job %s Selesai & Disimpan.\n", job.ID)
		}
	}()

	app := fiber.New()

	app.Post("/submit", func(c *fiber.Ctx) error {
		type Req struct { Image string `json:"image"`; Command string `json:"command"` }
		var p Req
		if err := c.BodyParser(&p); err != nil { return c.Status(400).SendString("Bad Request") }

		jobID := uuid.New().String()

		err := q.Enqueue(context.Background(), queue.Job{
			ID:      jobID,
			Image:   p.Image,
			Command: p.Command,
		})

		if err != nil {
			return c.Status(500).JSON(fiber.Map{"error": "Failed to enqueue"})
		}

		return c.JSON(fiber.Map{
			"status": "queued",
			"job_id": jobID,
			"info":   "Gunakan GET /status/:job_id untuk melihat hasil",
		})
	})

	app.Get("/status/:job_id", func(c *fiber.Ctx) error {
		jobID := c.Params("job_id")
		
		res, err := q.GetResult(context.Background(), jobID)
		if err != nil {
			if err.Error() == "pending" {
				return c.JSON(fiber.Map{"status": "processing", "message": "Sabar ya bang, lagi dikerjain worker..."})
			}
			return c.Status(404).JSON(fiber.Map{"status": "not_found", "error": "Job ID salah atau expired"})
		}

		return c.JSON(fiber.Map{
			"status": "completed",
			"output": res,
		})
	})

	app.Static("/", "./cmd/gateway/index.html")

	log.Fatal(app.Listen(cfg.Server.Port))
}