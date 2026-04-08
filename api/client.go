package api

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"
)

type RegisterRequest struct {
	Hostname string `json:"hostname"`
	OSInfo   string `json:"os_info"`
}

type RegisterResponse struct {
	AgentID string `json:"agent_id"`
	Token   string `json:"token"`
}

type Task struct {
	ID      string `json:"id"`
	Command string `json:"command"`
}

type TaskResponse struct {
	Task *Task `json:"task"`
}

type CompleteTaskRequest struct {
	Status string  `json:"status"`
	Error  *string `json:"error"`
}

type Client struct {
	BackendURL string
	HTTPClient *http.Client
}

func NewClient(backendURL string) *Client {
	return &Client{
		BackendURL: backendURL,
		HTTPClient: &http.Client{Timeout: 10 * time.Second},
	}
}

func (c *Client) Register(req RegisterRequest) (*RegisterResponse, error) {
	jsonData, err := json.Marshal(req)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request: %w", err)
	}

	url := fmt.Sprintf("%s/api/v1/agent/register", c.BackendURL)
	resp, err := c.HTTPClient.Post(url, "application/json", bytes.NewBuffer(jsonData))
	if err != nil {
		return nil, fmt.Errorf("failed to register agent: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusCreated {
		bodyInfo, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("bad status: %d, response: %s", resp.StatusCode, string(bodyInfo))
	}

	var regResp RegisterResponse
	if err := json.NewDecoder(resp.Body).Decode(&regResp); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}
	return &regResp, nil
}

func (c *Client) PollPendingTask(token string) (*Task, error) {
	reqURL := fmt.Sprintf("%s/api/v1/agent/tasks/pending", c.BackendURL)
	req, err := http.NewRequest(http.MethodGet, reqURL, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}
	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", token))

	resp, err := c.HTTPClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to execute request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusNoContent {
		return nil, nil // No pending tasks
	}

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("failed to poll tasks, status: %d", resp.StatusCode)
	}

	var taskResp TaskResponse
	if err := json.NewDecoder(resp.Body).Decode(&taskResp); err != nil {
		return nil, fmt.Errorf("error decoding task response: %w", err)
	}

	return taskResp.Task, nil
}

func (c *Client) CompleteTask(token, taskID string) error {
	completeURL := fmt.Sprintf("%s/api/v1/agent/tasks/%s/complete", c.BackendURL, taskID)
	reqBody := CompleteTaskRequest{
		Status: "completed",
		Error:  nil,
	}

	jsonData, err := json.Marshal(reqBody)
	if err != nil {
		return fmt.Errorf("failed to marshal request: %w", err)
	}

	req, err := http.NewRequest(http.MethodPost, completeURL, bytes.NewBuffer(jsonData))
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}
	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", token))
	req.Header.Set("Content-Type", "application/json")

	resp, err := c.HTTPClient.Do(req)
	if err != nil {
		return fmt.Errorf("failed to execute request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusNoContent {
		return fmt.Errorf("failed to complete task, status: %d", resp.StatusCode)
	}

	return nil
}
