package main

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/gofiber/fiber/v2"

	"github.com/JullMol/nebula/internal/gateway/proxy"
	"github.com/JullMol/nebula/internal/orchestrator/scheduler"
	"github.com/JullMol/nebula/pkg/config"
)

func main() {
	cfg, _ := config.LoadConfig()
	lb := scheduler.NewRoundRobin()
	proxySvc := proxy.NewProxyService(lb, cfg.Server.Workers)

	app := fiber.New()

	app.Post("/run", func(c *fiber.Ctx) error {
		type Req struct { Image string `json:"image"`; Command string `json:"command"` }
		var p Req
		if err := c.BodyParser(&p); err != nil { return c.Status(400).SendString("Bad Request") }

		ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
		defer cancel()

		resp, err := proxySvc.ForwardRunRequest(ctx, p.Image, p.Command)
		if err != nil {
			return c.Status(500).JSON(fiber.Map{"error": err.Error()})
		}

		return c.JSON(resp)
	})

	app.Get("/logs/:id", func(c *fiber.Ctx) error {
		logs, err := proxySvc.ForwardLogRequest(context.Background(), c.Params("id"))
		if err != nil { return c.Status(500).SendString(err.Error()) }
		return c.SendString(logs)
	})

	fmt.Printf("🌍 Gateway running on %s with %d workers\n", cfg.Server.Port, len(cfg.Server.Workers))
	log.Fatal(app.Listen(cfg.Server.Port))
}