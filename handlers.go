package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
)

var (
	InfoLogger  *log.Logger
	ErrorLogger *log.Logger
	DebugLogger *log.Logger
)

func init() {
	InfoLogger = log.New(os.Stdout, "INFO: ", log.Ldate|log.Ltime|log.Lshortfile)
	ErrorLogger = log.New(os.Stderr, "ERROR: ", log.Ldate|log.Ltime|log.Lshortfile)
	DebugLogger = log.New(os.Stdout, "DEBUG: ", log.Ldate|log.Ltime|log.Lshortfile)
}

type WebhookPayload struct {
	ObjectKind       string           `json:"object_kind"`
	Project          ProjectInfo      `json:"project"`
	ObjectAttributes ObjectAttributes `json:"object_attributes"`
}

type ProjectInfo struct {
	ID int `json:"id"`
}

type ObjectAttributes struct {
	IID    int    `json:"iid"`
	State  string `json:"state"`
	Action string `json:"action"`
}

type GitlabNotePayload struct {
	Body string `json:"body"`
}

func handleWebhook(cfg Config) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		gitlabToken := r.Header.Get("X-Gitlab-Token")
		if cfg.WebhookSecret != "" {
			if gitlabToken == "" {
				http.Error(w, "Unauthorized: Missing X-Gitlab-Token header", http.StatusUnauthorized)
				ErrorLogger.Printf("Unauthorized: Missing X-Gitlab-Token header from %s", r.RemoteAddr)
				return
			}
			if gitlabToken != cfg.WebhookSecret {
				http.Error(w, "Unauthorized: Invalid X-Gitlab-Token", http.StatusUnauthorized)
				ErrorLogger.Printf("Unauthorized: Invalid X-Gitlab-Token from %s", r.RemoteAddr)
				return
			}
			DebugLogger.Printf("Webhook token validation successful for request from %s", r.RemoteAddr)
		} else {
			InfoLogger.Printf("Warning: Webhook secret not configured, skipping token validation")
		}

		if r.Method != http.MethodPost {
			http.Error(w, "Method Not Allowed", http.StatusMethodNotAllowed)
			ErrorLogger.Printf("Received non-POST request method: %s from %s", r.Method, r.RemoteAddr)
			return
		}

		body, err := io.ReadAll(r.Body)
		if err != nil {
			http.Error(w, "Failed to read request body", http.StatusInternalServerError)
			ErrorLogger.Printf("Failed reading request body: %v", err)
			return
		}
		defer r.Body.Close()

		var payload WebhookPayload
		err = json.Unmarshal(body, &payload)
		if err != nil {
			http.Error(w, "Bad Request: Failed to parse JSON payload", http.StatusBadRequest)
			ErrorLogger.Printf("Failed parsing JSON: %v. Body: %s", err, string(body))
			return
		}

		if payload.ObjectKind != "merge_request" {
			InfoLogger.Printf("Ignoring event kind '%s'", payload.ObjectKind)
			w.WriteHeader(http.StatusOK)
			fmt.Fprintf(w, "Event kind '%s' ignored", payload.ObjectKind)
			return
		}

		projectID := payload.Project.ID
		mergeRequestIID := payload.ObjectAttributes.IID
		mrAction := payload.ObjectAttributes.Action

		if projectID == 0 || mergeRequestIID == 0 {
			http.Error(w, "Bad Request: Missing project ID or merge request IID in payload", http.StatusBadRequest)
			ErrorLogger.Printf("Missing project_id (%d) or mr_iid (%d) in payload. Kind: %s. Body: %s",
				projectID, mergeRequestIID, payload.ObjectKind, string(body))
			return
		}

		InfoLogger.Printf("Received '%s' event for MR !%d in project %d (Action: %s)",
			payload.ObjectKind, mergeRequestIID, projectID, mrAction)

		// Pretty print the payload for debugging
		prettyPayload, err := json.MarshalIndent(payload, "", "  ")
		if err != nil {
			ErrorLogger.Printf("Failed to format payload for printing: %v", err)
		} else {
			DebugLogger.Printf("Full webhook payload:\n%s", string(prettyPayload))
		}

		commentBody := fmt.Sprintf("Received webhook event via direct API call. Merge Request IID is: `%d`", mergeRequestIID)
		go func(config Config, projID, mrIID int, body string) {
			InfoLogger.Printf("Goroutine started: Posting comment to MR !%d in project %d", mrIID, projID)
			err := postGitlabComment(config, projID, mrIID, body)
			if err != nil {
				ErrorLogger.Printf("Failed to post comment asynchronously to MR !%d in project %d: %v", mrIID, projID, err)
				return
			}
			InfoLogger.Printf("Successfully posted comment asynchronously to MR !%d in project %d", mrIID, projID)
		}(cfg, projectID, mergeRequestIID, commentBody)

		w.WriteHeader(http.StatusOK)
		fmt.Fprintf(w, "Webhook received and comment posted successfully for MR !%d.", mergeRequestIID)
	}
}

// postGitlabComment makes a direct HTTP request to the GitLab API to create an MR note
func postGitlabComment(cfg Config, projectID int, mergeRequestIID int, commentBody string) error {
	baseURL := cfg.GitlabBaseURL
	if baseURL == "" {
		baseURL = "https://gitlab.com"
	}

	apiURL := fmt.Sprintf("%s/api/v4/projects/%d/merge_requests/%d/notes",
		baseURL,
		projectID,
		mergeRequestIID,
	)

	notePayload := GitlabNotePayload{Body: commentBody}
	jsonBody, err := json.Marshal(notePayload)
	if err != nil {
		ErrorLogger.Printf("Failed to marshal comment JSON payload: %v", err)
		return fmt.Errorf("failed to marshal comment JSON: %w", err)
	}

	req, err := http.NewRequest(http.MethodPost, apiURL, bytes.NewBuffer(jsonBody))
	if err != nil {
		ErrorLogger.Printf("Failed to create new HTTP request: %v", err)
		return fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("PRIVATE-TOKEN", cfg.GitlabToken)

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		ErrorLogger.Printf("Failed to execute request to GitLab API: %v", err)
		return fmt.Errorf("failed to execute request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		respBodyBytes, readErr := io.ReadAll(resp.Body)
		respBody := "Could not read response body."
		if readErr == nil {
			respBody = string(respBodyBytes)
		}
		ErrorLogger.Printf("GitLab API returned non-success status code: %d. URL: %s. Response: %s",
			resp.StatusCode, apiURL, respBody)
		return fmt.Errorf("gitlab API error: status code %d", resp.StatusCode)
	}

	InfoLogger.Printf("GitLab API call successful (Status Code: %d)", resp.StatusCode)
	return nil
}
