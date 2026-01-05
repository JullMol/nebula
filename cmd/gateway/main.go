package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"strings"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/limiter"
	"github.com/google/uuid"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"gorm.io/gorm"

	"github.com/JullMol/nebula/internal/gateway/proxy"
	"github.com/JullMol/nebula/internal/orchestrator/scheduler"
	"github.com/JullMol/nebula/internal/platform/database"
	"github.com/JullMol/nebula/internal/platform/queue"
	"github.com/JullMol/nebula/pkg/config"
)

var (
	jobsSubmitted = promauto.NewCounter(prometheus.CounterOpts{
		Name: "nebula_jobs_submitted_total",
		Help: "Total jumlah job yang disubmit user",
	})

	jobsProcessed = promauto.NewCounterVec(prometheus.CounterOpts{
		Name: "nebula_jobs_processed_total",
		Help: "Total job selesai berdasarkan status",
	}, []string{"status"})
)

func main() {
	cfg, _ := config.LoadConfig()

	dsn := "host=127.0.0.1 user=nebula password=nebula_password dbname=nebula_db port=5433 sslmode=disable TimeZone=Asia/Jakarta"
	db, err := database.NewConnection(dsn)
	if err != nil {
		log.Fatalf("❌ Gagal konek ke Database: %v", err)
	}
	fmt.Println("✅ Connected to PostgreSQL Database")

	lb := scheduler.NewRoundRobin()
	proxySvc := proxy.NewProxyService(lb, cfg.Server.Workers)
	q := queue.NewRedisQueue(cfg.Server.RedisAddr)
	fmt.Println("✅ Connected to Redis Queue")

	go func() {
		http.Handle("/metrics", promhttp.Handler())
		fmt.Println("📊 Metrics server running on :3001")
		http.ListenAndServe(":3001", nil)
	}()

	go func() {
		fmt.Println("🚜 Background Dispatcher Started...")
		for {
			ctx := context.Background()
			job, err := q.Dequeue(ctx)
			if err != nil {
				time.Sleep(1 * time.Second)
				continue
			}

			fmt.Printf("🚜 Processing Job ID: %s (Image: %s)\n", job.ID, job.Image)

			db.Model(&database.Job{}).Where("id = ?", job.ID).Update("status", "running")

			resp, err := proxySvc.ForwardRunRequest(ctx, job.Image, job.Command, job.Code)

			var resultLog string
			var finalStatus string

			if err != nil {
				fmt.Printf("❌ Job %s Gagal: %v\n", job.ID, err)
				resultLog = fmt.Sprintf("Error executing job: %v", err)
				finalStatus = "failed"
			} else {
				proxySvc.ForwardWaitRequest(ctx, resp.ContainerId)
				logs, _ := proxySvc.ForwardLogRequest(ctx, resp.ContainerId)
				resultLog = strings.ReplaceAll(logs, "\x00", "")
				finalStatus = "completed"
			}
			db.Model(&database.Job{}).Where("id = ?", job.ID).Updates(map[string]interface{}{
				"status": finalStatus,
				"result": resultLog,
				"updated_at": time.Now(),
			})

			jobsProcessed.WithLabelValues(finalStatus).Inc()

			fmt.Printf("✅ Job %s Selesai. Status: %s\n", job.ID, finalStatus)
		}
	}()

	app := fiber.New()

	app.Use(limiter.New(limiter.Config{
		Max:          10,
		Expiration:   30 * time.Second,
		KeyGenerator: func(c *fiber.Ctx) string { return c.IP() },
		LimitReached: func(c *fiber.Ctx) error {
			return c.Status(429).JSON(fiber.Map{"message": "Sabar bang! Jangan spam."})
		},
	}))

	app.Use(func(c *fiber.Ctx) error {
		if c.Path() == "/" || c.Path() == "/metrics" || (len(c.Path()) > 7 && c.Path()[:7] == "/status") {
			return c.Next()
		}
		apiKey := c.Get("X-API-KEY")
		if apiKey != "rahasia-negara" {
			return c.Status(401).JSON(fiber.Map{"message": "Siapa lu? Mana kuncinya? (Unauthorized)"})
		}
		return c.Next()
	})

	app.Post("/submit", func(c *fiber.Ctx) error {
		type Req struct {
			Image   string `json:"image"`
			Command string `json:"command"`
			Code    string `json:"code"`
		}
		var p Req
		if err := c.BodyParser(&p); err != nil {
			return c.Status(400).SendString("Bad Request")
		}

		jobID := uuid.New().String()

		newJob := database.Job{
			ID:        jobID,
			Image:     p.Image,
			Command:   p.Command,
			Status:    "queued",
			CreatedAt: time.Now(),
			UpdatedAt: time.Now(),
		}
		if err := db.Create(&newJob).Error; err != nil {
			return c.Status(500).JSON(fiber.Map{"error": "Gagal menyimpan ke database"})
		}

		err := q.Enqueue(context.Background(), queue.Job{
			ID:      jobID,
			Image:   p.Image,
			Command: p.Command,
			Code:    p.Code,
		})

		if err != nil {
			return c.Status(500).JSON(fiber.Map{"error": "Failed to enqueue"})
		}

		jobsSubmitted.Inc()

		return c.JSON(fiber.Map{
			"status": "queued",
			"job_id": jobID,
			"info":   "Job tersimpan di DB & Masuk Redis",
		})
	})

	app.Get("/status/:job_id", func(c *fiber.Ctx) error {
		jobID := c.Params("job_id")
		
		var job database.Job
		result := db.First(&job, "id = ?", jobID)
		
		if result.Error != nil {
			if result.Error == gorm.ErrRecordNotFound {
				return c.Status(404).JSON(fiber.Map{"error": "Job tidak ditemukan"})
			}
			return c.Status(500).JSON(fiber.Map{"error": "Database error"})
		}

		return c.JSON(fiber.Map{
			"job_id":     job.ID,
			"status":     job.Status,
			"result":     job.Result,
			"created_at": job.CreatedAt,
			"updated_at": job.UpdatedAt,
		})
	})

	app.Static("/", "./cmd/gateway/index.html")
	log.Fatal(app.Listen(cfg.Server.Port))
}