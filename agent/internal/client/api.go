package client

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"time"

	"github.com/alexGoLyceum/calculator-service/agent/internal/config"
	"github.com/alexGoLyceum/calculator-service/agent/internal/tasks"
)

var apiBaseURL = "http://localhost:8080/api/v1/internal/task"

func SetUpUrl(cfg config.OrchestratorConfig) {
	if cfg.Host != "" && cfg.Port != 0 {
		apiBaseURL = fmt.Sprintf("http://%s:%d/api/v1/internal/task", cfg.Host, cfg.Port)
	}
}

var HTTPClient = &http.Client{Timeout: 5 * time.Second}

type SetTaskResultRequest struct {
	Task   tasks.Task `json:"task"`
	Result float64    `json:"result"`
}

type SetTaskResultResponse struct {
	Error string `json:"error"`
}

type GetTaskResponse struct {
	Task tasks.Task `json:"task"`
}

func GetTask() (*tasks.Task, error) {
	parsedURL, err := url.Parse(apiBaseURL)
	if err != nil {
		return nil, fmt.Errorf("failed to parse URL: %w", err)
	}

	resp, err := HTTPClient.Get(parsedURL.String())
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusOK {
		var response GetTaskResponse
		if err := json.NewDecoder(resp.Body).Decode(&response); err != nil {
			return nil, fmt.Errorf("failed to decode response: %w", err)
		}
		return &response.Task, nil
	}
	if resp.StatusCode == http.StatusNotFound {
		return nil, nil
	}
	return nil, fmt.Errorf("unexpected status code: %d", resp.StatusCode)
}

func SetTaskResult(task tasks.Task, result float64) error {
	requestBody, err := json.Marshal(SetTaskResultRequest{Task: task, Result: result})
	if err != nil {
		return fmt.Errorf("failed to marshal request: %w", err)
	}

	req, err := http.NewRequest(http.MethodPost, apiBaseURL, bytes.NewBuffer(requestBody))
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := HTTPClient.Do(req)
	if err != nil {
		return fmt.Errorf("failed to send request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusOK {
		return nil
	}

	var response SetTaskResultResponse
	if err := json.NewDecoder(resp.Body).Decode(&response); err != nil {
		return fmt.Errorf("failed to decode response: %w", err)
	}

	if response.Error != "" {
		return fmt.Errorf("failed to set task result: %s", response.Error)
	}

	return nil
}
