package truenas

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestListContainers_success(t *testing.T) {
	t.Parallel()

	containers := []Container{
		{Name: "nginx", State: "RUNNING"},
		{Name: "redis", State: "STOPPED"},
	}

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet || r.URL.Path != "/app" {
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

	container := Container{Name: "nginx", State: "RUNNING"}

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet || r.URL.Path != "/app/id/nginx" {
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
	got, err := c.GetContainer(context.Background(), "nginx")
	if err != nil {
		t.Fatalf("GetContainer: %v", err)
	}
	if got.Name != "nginx" {
		t.Errorf("Name = %q, want %q", got.Name, "nginx")
	}
}

func TestGetContainer_notFound(t *testing.T) {
	t.Parallel()

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusNotFound)
	}))
	defer srv.Close()

	c := newTestClient(t, srv.URL)
	_, err := c.GetContainer(context.Background(), "does-not-exist")
	if !errors.Is(err, ErrNotFound) {
		t.Fatalf("expected ErrNotFound, got %v", err)
	}
}

func TestStartContainer_success(t *testing.T) {
	t.Parallel()

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost || r.URL.Path != "/app/start" {
			w.WriteHeader(http.StatusNotFound)
			return
		}
		var name string
		if err := json.NewDecoder(r.Body).Decode(&name); err != nil || name != "nginx" {
			w.WriteHeader(http.StatusBadRequest)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		if err := json.NewEncoder(w).Encode(10); err != nil {
			t.Errorf("encoding response: %v", err)
		}
	}))
	defer srv.Close()

	c := newTestClient(t, srv.URL)
	jobID, err := c.StartContainer(context.Background(), "nginx")
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
		if r.Method != http.MethodPost || r.URL.Path != "/app/stop" {
			w.WriteHeader(http.StatusNotFound)
			return
		}
		var name string
		if err := json.NewDecoder(r.Body).Decode(&name); err != nil || name != "nginx" {
			w.WriteHeader(http.StatusBadRequest)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		if err := json.NewEncoder(w).Encode(11); err != nil {
			t.Errorf("encoding response: %v", err)
		}
	}))
	defer srv.Close()

	c := newTestClient(t, srv.URL)
	jobID, err := c.StopContainer(context.Background(), "nginx")
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
		// TrueNAS SCALE uses /app/redeploy (not /restart) with the app name as the JSON body.
		if r.Method != http.MethodPost || r.URL.Path != "/app/redeploy" {
			w.WriteHeader(http.StatusNotFound)
			return
		}
		var name string
		if err := json.NewDecoder(r.Body).Decode(&name); err != nil || name != "nginx" {
			w.WriteHeader(http.StatusBadRequest)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		if err := json.NewEncoder(w).Encode(12); err != nil {
			t.Errorf("encoding response: %v", err)
		}
	}))
	defer srv.Close()

	c := newTestClient(t, srv.URL)
	jobID, err := c.RestartContainer(context.Background(), "nginx")
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
		{ID: "sha256:abc123", RepoTags: []string{"nginx:latest"}, Size: 142000000},
		{ID: "sha256:def456", RepoTags: []string{"redis:7"}, Size: 45000000},
	}

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet || r.URL.Path != "/app/image" {
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

func TestCreateContainer_catalogSuccess(t *testing.T) {
	t.Parallel()

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost || r.URL.Path != "/app" {
			w.WriteHeader(http.StatusNotFound)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		if err := json.NewEncoder(w).Encode(42); err != nil {
			t.Errorf("encoding response: %v", err)
		}
	}))
	defer srv.Close()

	c := newTestClient(t, srv.URL)
	jobID, err := c.CreateContainer(context.Background(), &CreateContainerParams{
		AppName:    "my-jellyfin",
		CatalogApp: "jellyfin",
		Train:      "stable",
		Version:    "latest",
	})
	if err != nil {
		t.Fatalf("CreateContainer: %v", err)
	}
	if jobID != 42 {
		t.Errorf("jobID = %d, want 42", jobID)
	}
}

func TestCreateContainer_customSuccess(t *testing.T) {
	t.Parallel()

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost || r.URL.Path != "/app" {
			w.WriteHeader(http.StatusNotFound)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		if err := json.NewEncoder(w).Encode(7); err != nil {
			t.Errorf("encoding response: %v", err)
		}
	}))
	defer srv.Close()

	c := newTestClient(t, srv.URL)
	jobID, err := c.CreateContainer(context.Background(), &CreateContainerParams{
		AppName:                   "my-custom-app",
		CustomApp:                 true,
		CustomComposeConfigString: "services:\n  web:\n    image: nginx\n",
	})
	if err != nil {
		t.Fatalf("CreateContainer: %v", err)
	}
	if jobID != 7 {
		t.Errorf("jobID = %d, want 7", jobID)
	}
}

func TestCreateContainer_validation(t *testing.T) {
	t.Parallel()

	c := newTestClient(t, "http://localhost") // no server needed — validation is local

	tests := []struct {
		name   string
		params *CreateContainerParams
	}{
		{"nil params", nil},
		{"empty app_name", &CreateContainerParams{CatalogApp: "jellyfin"}},
		{"app_name too long", &CreateContainerParams{AppName: "a123456789012345678901234567890123456789x", CatalogApp: "jellyfin"}},
		{"app_name invalid chars", &CreateContainerParams{AppName: "My_App", CatalogApp: "jellyfin"}},
		{"app_name starts with hyphen", &CreateContainerParams{AppName: "-bad", CatalogApp: "jellyfin"}},
		{"catalog_app missing", &CreateContainerParams{AppName: "myapp"}},
		{"custom missing compose", &CreateContainerParams{AppName: "myapp", CustomApp: true}},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			_, err := c.CreateContainer(context.Background(), tc.params)
			if err == nil {
				t.Fatal("expected error, got nil")
			}
		})
	}
}

func TestDeleteContainer_success(t *testing.T) {
	t.Parallel()

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodDelete || r.URL.Path != "/app/id/my-app" {
			w.WriteHeader(http.StatusNotFound)
			return
		}
		w.WriteHeader(http.StatusOK)
	}))
	defer srv.Close()

	c := newTestClient(t, srv.URL)
	if err := c.DeleteContainer(context.Background(), "my-app"); err != nil {
		t.Fatalf("DeleteContainer: %v", err)
	}
}

func TestDeleteContainer_notFound(t *testing.T) {
	t.Parallel()

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusNotFound)
	}))
	defer srv.Close()

	c := newTestClient(t, srv.URL)
	if err := c.DeleteContainer(context.Background(), "does-not-exist"); !errors.Is(err, ErrNotFound) {
		t.Fatalf("expected ErrNotFound, got %v", err)
	}
}
