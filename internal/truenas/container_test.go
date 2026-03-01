package truenas

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestListContainers_success(t *testing.T) {
	t.Parallel()

	containers := []Container{
		{ID: 1, Name: "nginx", Image: "nginx:latest", Status: ContainerStatus{State: "RUNNING"}},
		{ID: 2, Name: "redis", Image: "redis:7", Status: ContainerStatus{State: "STOPPED"}},
	}

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet || r.URL.Path != "/container/container" {
			w.WriteHeader(http.StatusNotFound)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		if err := json.NewEncoder(w).Encode(containers); err != nil {
			t.Errorf("encoding response: %v", err)
		}
	}))
	defer srv.Close()

	c := newTestClient(t, srv.URL)
	got, err := c.ListContainers(context.Background())
	if err != nil {
		t.Fatalf("ListContainers: %v", err)
	}
	if len(got) != 2 {
		t.Errorf("len = %d, want 2", len(got))
	}
	if got[0].Name != "nginx" {
		t.Errorf("Name = %q, want %q", got[0].Name, "nginx")
	}
}

func TestGetContainer_success(t *testing.T) {
	t.Parallel()

	container := Container{ID: 1, Name: "nginx", Image: "nginx:latest", Status: ContainerStatus{State: "RUNNING"}}

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet || r.URL.Path != "/container/container/id/1" {
			w.WriteHeader(http.StatusNotFound)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		if err := json.NewEncoder(w).Encode(container); err != nil {
			t.Errorf("encoding response: %v", err)
		}
	}))
	defer srv.Close()

	c := newTestClient(t, srv.URL)
	got, err := c.GetContainer(context.Background(), 1)
	if err != nil {
		t.Fatalf("GetContainer: %v", err)
	}
	if got.ID != 1 {
		t.Errorf("ID = %d, want 1", got.ID)
	}
}

func TestGetContainer_notFound(t *testing.T) {
	t.Parallel()

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusNotFound)
	}))
	defer srv.Close()

	c := newTestClient(t, srv.URL)
	_, err := c.GetContainer(context.Background(), 99)
	if err == nil {
		t.Fatal("expected error, got nil")
	}
}

func TestStartContainer_success(t *testing.T) {
	t.Parallel()

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost || r.URL.Path != "/container/container/id/1/start" {
			w.WriteHeader(http.StatusNotFound)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		if err := json.NewEncoder(w).Encode(10); err != nil {
			t.Errorf("encoding response: %v", err)
		}
	}))
	defer srv.Close()

	c := newTestClient(t, srv.URL)
	jobID, err := c.StartContainer(context.Background(), 1)
	if err != nil {
		t.Fatalf("StartContainer: %v", err)
	}
	if jobID != 10 {
		t.Errorf("jobID = %d, want 10", jobID)
	}
}

func TestStopContainer_success(t *testing.T) {
	t.Parallel()

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost || r.URL.Path != "/container/container/id/1/stop" {
			w.WriteHeader(http.StatusNotFound)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		if err := json.NewEncoder(w).Encode(11); err != nil {
			t.Errorf("encoding response: %v", err)
		}
	}))
	defer srv.Close()

	c := newTestClient(t, srv.URL)
	jobID, err := c.StopContainer(context.Background(), 1)
	if err != nil {
		t.Fatalf("StopContainer: %v", err)
	}
	if jobID != 11 {
		t.Errorf("jobID = %d, want 11", jobID)
	}
}

func TestRestartContainer_success(t *testing.T) {
	t.Parallel()

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost || r.URL.Path != "/container/container/id/1/restart" {
			w.WriteHeader(http.StatusNotFound)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		if err := json.NewEncoder(w).Encode(12); err != nil {
			t.Errorf("encoding response: %v", err)
		}
	}))
	defer srv.Close()

	c := newTestClient(t, srv.URL)
	jobID, err := c.RestartContainer(context.Background(), 1)
	if err != nil {
		t.Fatalf("RestartContainer: %v", err)
	}
	if jobID != 12 {
		t.Errorf("jobID = %d, want 12", jobID)
	}
}

func TestListImages_success(t *testing.T) {
	t.Parallel()

	images := []Image{
		{ID: 1, RepoTags: []string{"nginx:latest"}, Size: 142000000},
		{ID: 2, RepoTags: []string{"redis:7"}, Size: 45000000},
	}

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet || r.URL.Path != "/container/image" {
			w.WriteHeader(http.StatusNotFound)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		if err := json.NewEncoder(w).Encode(images); err != nil {
			t.Errorf("encoding response: %v", err)
		}
	}))
	defer srv.Close()

	c := newTestClient(t, srv.URL)
	got, err := c.ListImages(context.Background())
	if err != nil {
		t.Fatalf("ListImages: %v", err)
	}
	if len(got) != 2 {
		t.Errorf("len = %d, want 2", len(got))
	}
	if got[0].RepoTags[0] != "nginx:latest" {
		t.Errorf("RepoTags[0] = %q, want %q", got[0].RepoTags[0], "nginx:latest")
	}
}
