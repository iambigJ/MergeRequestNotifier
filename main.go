package main

import (
	"log"
	"net/http"
	"os"

	"gitlab-request/handler"

	"github.com/joho/godotenv"
)

type Config struct {
	Port          string
	GitlabToken   string
	WebhookSecret string
	GitlabBaseURL string
}

func main() {
	err := godotenv.Load()
	if err != nil {
		log.Printf("Warning: Error loading .env file: %v", err)
	}

	cfg := loadConfig()
	if cfg.GitlabToken == "" || cfg.GitlabBaseURL == "" || cfg.Port == "" {
		log.Fatal("FATAL: Required environment variables are not set. Please check GITLAB_TOKEN, GITLAB_BASE_URL, and PORT.")
	}

	log.Println("Configuration loaded.")

	// Convert main.Config to handler.Config
	webhookCfg := handler.Config{
		GitlabToken:   cfg.GitlabToken,
		WebhookSecret: cfg.WebhookSecret,
		GitlabBaseURL: cfg.GitlabBaseURL,
	}

	http.HandleFunc("/webhook", handler.HandleWebhook(webhookCfg))

	listenAddr := ":" + cfg.Port
	log.Printf("INFO: Starting server on %s\n", listenAddr)
	err = http.ListenAndServe(listenAddr, nil)
	if err != nil {
		log.Fatalf("FATAL: Failed to start server: %v", err)
	}
}

func loadConfig() Config {
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080" // Default port
	}
	baseURL := os.Getenv("GITLAB_BASE_URL")
	if baseURL == "" {
		baseURL = "https://gitlab.com" // Default GitLab URL
	}

	return Config{
		Port:          port,
		GitlabToken:   os.Getenv("GITLAB_TOKEN"),
		WebhookSecret: os.Getenv("GITLAB_WEBHOOK_SECRET"),
		GitlabBaseURL: baseURL,
	}
}
