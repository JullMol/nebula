package main

import (
	"bytes"
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"net/http"
	"os"
	"time"
)

type RunRequest struct {
	Image   string `json:"image"`
	Command string `json:"command"`
}

type RunResponse struct {
	Status      string `json:"status"`
	ContainerID string `json:"container_id"`
	Message     string `json:"message"`
	Error       string `json:"error,omitempty"`
}

func main() {
	imagePtr := flag.String("image", "alpine", "Docker image to use (default: alpine)")
	cmdPtr := flag.String("cmd", "", "Command to run inside container")
	flag.Parse()

	if *cmdPtr == "" {
		fmt.Println("❌ Error: Command tidak boleh kosong.")
		fmt.Println("👉 Contoh: nebula-cli -cmd \"echo hello\"")
		os.Exit(1)
	}

	gatewayURL := "http://localhost:3000"
	fmt.Printf("🚀 Deploying function to Nebula... (Image: %s)\n", *imagePtr)

	reqBody, _ := json.Marshal(RunRequest{
		Image:   *imagePtr,
		Command: *cmdPtr,
	})

	resp, err := http.Post(gatewayURL+"/run", "application/json", bytes.NewBuffer(reqBody))
	if err != nil {
		fmt.Printf("❌ Gagal menghubungi Gateway: %v\n", err)
		os.Exit(1)
	}
	defer resp.Body.Close()

	var result RunResponse
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		fmt.Printf("❌ Gagal parsing response: %v\n", err)
		os.Exit(1)
	}

	if result.Error != "" {
		fmt.Printf("❌ Server Error: %s\n", result.Error)
		os.Exit(1)
	}

	fmt.Printf("✅ Success! Container ID: %s\n", result.ContainerID)
	fmt.Println("⏳ Waiting for logs...")

	time.Sleep(2 * time.Second)

	logResp, err := http.Get(fmt.Sprintf("%s/logs/%s", gatewayURL, result.ContainerID))
	if err != nil {
		fmt.Printf("⚠️ Gagal mengambil logs: %v\n", err)
		return
	}
	defer logResp.Body.Close()

	logs, _ := io.ReadAll(logResp.Body)

	fmt.Println("================ OUTPUT ================")
	fmt.Println(string(logs))
	fmt.Println("========================================")
}