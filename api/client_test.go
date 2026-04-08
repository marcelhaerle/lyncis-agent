package api

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestPollPendingTask(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Header.Get("Authorization") != "Bearer test-token" {
			t.Errorf("expected bearer token")
		}
		if r.URL.Path != "/api/v1/agent/tasks/pending" {
			t.Errorf("unexpected path: %s", r.URL.Path)
		}

		resp := TaskResponse{
			Task: &Task{
				ID:      "task-123",
				Command: "run_lynis",
			},
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(resp)
	}))
	defer ts.Close()

	client := NewClient(ts.URL)
	task, err := client.PollPendingTask("test-token")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if task == nil {
		t.Fatalf("expected task, got nil")
	}

	if task.ID != "task-123" || task.Command != "run_lynis" {
		t.Errorf("unexpected task: %+v", task)
	}
}

func TestPollPendingTask_NoTasks(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNoContent)
	}))
	defer ts.Close()

	client := NewClient(ts.URL)
	task, err := client.PollPendingTask("test-token")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if task != nil {
		t.Fatalf("expected nil task, got %+v", task)
	}
}
