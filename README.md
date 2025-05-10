# MergeRequestNotifier

A lightweight Go service that listens for GitLab webhook events and posts comments to merge requests.

## Features

- Receives GitLab webhook events for merge requests
- Authenticates requests using webhook secrets
- Posts comments to merge requests via GitLab API
- Configurable via environment variables

## Requirements

- Go 1.16+
- GitLab Personal Access Token with API access

## Installation

1. Clone the repository:
   ```
   git clone https://github.com/yourusername/gitlab-request.git
   cd gitlab-request
   ```

2. Install dependencies:
   ```
   go mod download
   ```

3. Create a `.env` file based on the example:
   ```
   cp .env.example .env
   ```

4. Edit the `.env` file with your GitLab token and webhook secret.

## Configuration

Configure the application using the following environment variables:

| Variable | Description | Default |
|----------|-------------|---------|
| PORT | Port the server will listen on | 8080 |
| GITLAB_TOKEN | GitLab Personal Access Token | (required) |
| GITLAB_BASE_URL | Base URL for GitLab API | https://gitlab.com |
| GITLAB_WEBHOOK_SECRET | Secret for webhook authentication | (optional) |

## Usage

1. Quick setup:
   ```
   make setup
   ```

2. Start the server for development (with live reloading):
   ```
   make dev
   ```

3. Or build and run:
   ```
   make run
   ```

4. Configure a webhook in your GitLab project:
   - Go to your project settings > Webhooks
   - Set URL to `http://your-server:8080/webhook`
   - Select "Merge request events"
   - Set your secret token (must match GITLAB_WEBHOOK_SECRET)
   - Click "Add webhook"

## Build and Deploy

Build a binary:
```
make build
```

Run the binary:
```
./bin/webhook-server
```

## Project Structure

- `main.go` - Entry point and configuration loading
- `handler/webhook.go` - Webhook handling logic 
- `handler/gitlab.go` - GitLab API integration
- `handler/config.go` - Configuration type definitions
- `Makefile` - Build and development commands

## License

MIT