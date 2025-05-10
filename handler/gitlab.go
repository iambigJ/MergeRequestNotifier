package handler

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
)

type GitlabNotePayload struct {
	Body string `json:"body"`
}

// PostGitlabComment makes a direct HTTP request to the GitLab API to create an MR note
func PostGitlabComment(cfg Config, projectID int, mergeRequestIID int, commentBody string) error {
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
