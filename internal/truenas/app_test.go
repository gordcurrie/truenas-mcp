package truenas

import (
	"context"
	"encoding/json"
	"errors"
	"testing"
)

func TestListApps_success(t *testing.T) {
	t.Parallel()

	apps := []App{
		{Name: "nginx", State: "RUNNING"},
		{Name: "redis", State: "STOPPED"},
	}

	srv := wsTestServer(t, map[string]methodHandler{
		"app.query": func(_ json.RawMessage) (any, *rpcError) {
			return apps, nil
		},
	})
	defer srv.Close()

	c := newTestClient(t, srv.URL)
	got, err := c.ListApps(context.Background())
	if err != nil {
		t.Fatalf("ListApps: %v", err)
	}
	if len(got) != 2 {
		t.Errorf("len = %d, want 2", len(got))
	}
	if got[0].Name != "nginx" {
		t.Errorf("Name = %q, want %q", got[0].Name, "nginx")
	}
}

func TestGetApp_success(t *testing.T) {
	t.Parallel()

	app := App{Name: "nginx", State: "RUNNING"}

	srv := wsTestServer(t, map[string]methodHandler{
		"app.get_instance": func(_ json.RawMessage) (any, *rpcError) {
			return app, nil
		},
	})
	defer srv.Close()

	c := newTestClient(t, srv.URL)
	got, err := c.GetApp(context.Background(), "nginx")
	if err != nil {
		t.Fatalf("GetApp: %v", err)
	}
	if got.Name != "nginx" {
		t.Errorf("Name = %q, want %q", got.Name, "nginx")
	}
}

func TestGetApp_notFound(t *testing.T) {
	t.Parallel()

	srv := wsTestServer(t, map[string]methodHandler{
		"app.get_instance": func(_ json.RawMessage) (any, *rpcError) {
			return nil, &rpcError{Code: -32001, Message: "not found"}
		},
	})
	defer srv.Close()

	c := newTestClient(t, srv.URL)
	_, err := c.GetApp(context.Background(), "does-not-exist")
	if !errors.Is(err, ErrNotFound) {
		t.Fatalf("expected ErrNotFound, got %v", err)
	}
}

func TestStartApp_success(t *testing.T) {
	t.Parallel()

	srv := wsTestServer(t, map[string]methodHandler{
		"app.start": func(_ json.RawMessage) (any, *rpcError) {
			return 10, nil
		},
	})
	defer srv.Close()

	c := newTestClient(t, srv.URL)
	jobID, err := c.StartApp(context.Background(), "nginx")
	if err != nil {
		t.Fatalf("StartApp: %v", err)
	}
	if jobID != 10 {
		t.Errorf("jobID = %d, want 10", jobID)
	}
}

func TestStopApp_success(t *testing.T) {
	t.Parallel()

	srv := wsTestServer(t, map[string]methodHandler{
		"app.stop": func(_ json.RawMessage) (any, *rpcError) {
			return 11, nil
		},
	})
	defer srv.Close()

	c := newTestClient(t, srv.URL)
	jobID, err := c.StopApp(context.Background(), "nginx")
	if err != nil {
		t.Fatalf("StopApp: %v", err)
	}
	if jobID != 11 {
		t.Errorf("jobID = %d, want 11", jobID)
	}
}

func TestRestartApp_success(t *testing.T) {
	t.Parallel()

	srv := wsTestServer(t, map[string]methodHandler{
		"app.redeploy": func(_ json.RawMessage) (any, *rpcError) {
			return 12, nil
		},
	})
	defer srv.Close()

	c := newTestClient(t, srv.URL)
	jobID, err := c.RestartApp(context.Background(), "nginx")
	if err != nil {
		t.Fatalf("RestartApp: %v", err)
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

	srv := wsTestServer(t, map[string]methodHandler{
		"app.image.query": func(_ json.RawMessage) (any, *rpcError) {
			return images, nil
		},
	})
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

func TestCreateApp_catalogSuccess(t *testing.T) {
	t.Parallel()

	srv := wsTestServer(t, map[string]methodHandler{
		"app.create": func(_ json.RawMessage) (any, *rpcError) {
			return 42, nil
		},
	})
	defer srv.Close()

	c := newTestClient(t, srv.URL)
	jobID, err := c.CreateApp(context.Background(), &CreateAppParams{
		AppName:    "my-jellyfin",
		CatalogApp: "jellyfin",
		Train:      "stable",
		Version:    "latest",
	})
	if err != nil {
		t.Fatalf("CreateApp: %v", err)
	}
	if jobID != 42 {
		t.Errorf("jobID = %d, want 42", jobID)
	}
}

func TestCreateApp_customSuccess(t *testing.T) {
	t.Parallel()

	srv := wsTestServer(t, map[string]methodHandler{
		"app.create": func(_ json.RawMessage) (any, *rpcError) {
			return 7, nil
		},
	})
	defer srv.Close()

	c := newTestClient(t, srv.URL)
	jobID, err := c.CreateApp(context.Background(), &CreateAppParams{
		AppName:                   "my-custom-app",
		CustomApp:                 true,
		CustomComposeConfigString: "services:\n  web:\n    image: nginx\n",
	})
	if err != nil {
		t.Fatalf("CreateApp: %v", err)
	}
	if jobID != 7 {
		t.Errorf("jobID = %d, want 7", jobID)
	}
}

func TestCreateApp_validation(t *testing.T) {
	t.Parallel()

	// No server needed — validation fires before any network call.
	c, err := NewClient("http://localhost:19999", "test-api-key", false) //nolint:gosec // G101: fake placeholder
	if err != nil {
		t.Fatalf("NewClient: %v", err)
	}

	tests := []struct {
		name   string
		params *CreateAppParams
	}{
		{"nil params", nil},
		{"empty app_name", &CreateAppParams{CatalogApp: "jellyfin"}},
		{"app_name too long", &CreateAppParams{AppName: "a123456789012345678901234567890123456789x", CatalogApp: "jellyfin"}},
		{"app_name invalid chars", &CreateAppParams{AppName: "My_App", CatalogApp: "jellyfin"}},
		{"app_name starts with hyphen", &CreateAppParams{AppName: "-bad", CatalogApp: "jellyfin"}},
		{"catalog_app missing", &CreateAppParams{AppName: "myapp"}},
		{"custom missing compose", &CreateAppParams{AppName: "myapp", CustomApp: true}},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			_, err := c.CreateApp(context.Background(), tc.params)
			if err == nil {
				t.Fatal("expected error, got nil")
			}
		})
	}
}

func TestDeleteApp_validation(t *testing.T) {
	t.Parallel()

	// No server needed — validation fires before any network call.
	c, err := NewClient("http://localhost:19999", "test-api-key", false) //nolint:gosec // G101: fake placeholder
	if err != nil {
		t.Fatalf("NewClient: %v", err)
	}

	tests := []struct {
		name    string
		appName string
	}{
		{"empty name", ""},
		{"name too long", "a123456789012345678901234567890123456789x"},
		{"invalid chars", "My_App"},
		{"starts with hyphen", "-bad"},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			if err := c.DeleteApp(context.Background(), tc.appName); err == nil {
				t.Fatal("expected error, got nil")
			}
		})
	}
}

func TestDeleteApp_success(t *testing.T) {
	t.Parallel()

	srv := wsTestServer(t, map[string]methodHandler{
		"app.delete": func(_ json.RawMessage) (any, *rpcError) {
			return nil, nil
		},
	})
	defer srv.Close()

	c := newTestClient(t, srv.URL)
	if err := c.DeleteApp(context.Background(), "my-app"); err != nil {
		t.Fatalf("DeleteApp: %v", err)
	}
}

func TestDeleteApp_notFound(t *testing.T) {
	t.Parallel()

	srv := wsTestServer(t, map[string]methodHandler{
		"app.delete": func(_ json.RawMessage) (any, *rpcError) {
			return nil, &rpcError{Code: -32001, Message: "not found"}
		},
	})
	defer srv.Close()

	c := newTestClient(t, srv.URL)
	if err := c.DeleteApp(context.Background(), "does-not-exist"); !errors.Is(err, ErrNotFound) {
		t.Fatalf("expected ErrNotFound, got %v", err)
	}
}

func TestUpgradeApp_success(t *testing.T) {
	t.Parallel()

	srv := wsTestServer(t, map[string]methodHandler{
		"app.upgrade": func(_ json.RawMessage) (any, *rpcError) {
			return 20, nil
		},
	})
	defer srv.Close()

	c := newTestClient(t, srv.URL)
	jobID, err := c.UpgradeApp(context.Background(), "my-app", "")
	if err != nil {
		t.Fatalf("UpgradeApp: %v", err)
	}
	if jobID != 20 {
		t.Errorf("jobID = %d, want 20", jobID)
	}
}

func TestUpgradeApp_validation(t *testing.T) {
	t.Parallel()

	c, err := NewClient("http://localhost:19999", "test-api-key", false) //nolint:gosec // G101: fake placeholder
	if err != nil {
		t.Fatalf("NewClient: %v", err)
	}

	tests := []struct {
		name    string
		appName string
	}{
		{"empty name", ""},
		{"name too long", "a123456789012345678901234567890123456789x"},
		{"invalid chars", "My_App"},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			_, err := c.UpgradeApp(context.Background(), tc.appName, "")
			if err == nil {
				t.Fatal("expected error, got nil")
			}
		})
	}
}

func TestGetUpgradeSummary_success(t *testing.T) {
	t.Parallel()

	changelog := "Bug fixes."
	summary := AppUpgradeSummary{
		LatestVersion:               "2.0.0",
		LatestHumanVersion:          "2.0.0_1.0.0",
		UpgradeVersion:              "2.0.0",
		UpgradeHumanVersion:         "2.0.0_1.0.0",
		AvailableVersionsForUpgrade: []AppVersionInfo{{Version: "2.0.0", HumanVersion: "2.0.0_1.0.0"}},
		Changelog:                   &changelog,
	}

	srv := wsTestServer(t, map[string]methodHandler{
		"app.upgrade_summary": func(_ json.RawMessage) (any, *rpcError) {
			return summary, nil
		},
	})
	defer srv.Close()

	c := newTestClient(t, srv.URL)
	got, err := c.GetUpgradeSummary(context.Background(), "my-app")
	if err != nil {
		t.Fatalf("GetUpgradeSummary: %v", err)
	}
	if !got.UpgradeAvailable {
		t.Errorf("UpgradeAvailable = false, want true")
	}
	if got.LatestVersion != "2.0.0" {
		t.Errorf("LatestVersion = %q, want %q", got.LatestVersion, "2.0.0")
	}
	if got.Changelog == nil || *got.Changelog != "Bug fixes." {
		t.Errorf("Changelog = %v, want %q", got.Changelog, "Bug fixes.")
	}
	if len(got.AvailableVersionsForUpgrade) != 1 {
		t.Errorf("AvailableVersionsForUpgrade len = %d, want 1", len(got.AvailableVersionsForUpgrade))
	}
}

func TestGetUpgradeSummary_noUpgrade(t *testing.T) {
	t.Parallel()

	srv := wsTestServer(t, map[string]methodHandler{
		"app.upgrade_summary": func(_ json.RawMessage) (any, *rpcError) {
			e := &rpcError{Code: -32001, Message: "no update available"}
			return nil, e
		},
	})
	defer srv.Close()

	c := newTestClient(t, srv.URL)
	got, err := c.GetUpgradeSummary(context.Background(), "my-app")
	if err != nil {
		t.Fatalf("expected no error for no-upgrade signal, got: %v", err)
	}
	if got.UpgradeAvailable {
		t.Errorf("UpgradeAvailable = true, want false")
	}
}

func TestRollbackApp_success(t *testing.T) {
	t.Parallel()

	srv := wsTestServer(t, map[string]methodHandler{
		"app.rollback": func(_ json.RawMessage) (any, *rpcError) {
			return 21, nil
		},
	})
	defer srv.Close()

	c := newTestClient(t, srv.URL)
	jobID, err := c.RollbackApp(context.Background(), "my-app", "1.9.0")
	if err != nil {
		t.Fatalf("RollbackApp: %v", err)
	}
	if jobID != 21 {
		t.Errorf("jobID = %d, want 21", jobID)
	}
}

func TestRollbackApp_emptyVersion(t *testing.T) {
	t.Parallel()

	c, err := NewClient("http://localhost:19999", "test-api-key", false) //nolint:gosec // G101: fake placeholder
	if err != nil {
		t.Fatalf("NewClient: %v", err)
	}
	_, err = c.RollbackApp(context.Background(), "my-app", "")
	if err == nil {
		t.Fatal("expected error for empty version, got nil")
	}
}
