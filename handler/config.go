package handler

// Config holds the configuration for the GitLab webhook handler
type Config struct {
	WebhookSecret string
	GitlabBaseURL string
	GitlabToken   string
}
