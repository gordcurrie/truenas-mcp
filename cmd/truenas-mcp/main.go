// Command truenas-mcp is an MCP server that exposes TrueNAS SCALE operations
// as MCP tools.
//
// Required environment variables:
//
//	TRUENAS_API_URL  Base URL of the TrueNAS SCALE host (e.g. https://truenas.local)
//	TRUENAS_API_KEY  API key created in TrueNAS UI under Credentials -> API Keys
//
// Optional environment variables:
//
//	TRUENAS_INSECURE          Set to "true" to skip TLS certificate verification
//	TRUENAS_ALLOW_DESTRUCTIVE Set to "true" to enable destructive tools (default: disabled)
//
// Flags:
//
//	--transport   Transport to use: "stdio" (default) or "http"
//	--addr        Listen address when --transport=http (default: localhost:8081)
package main

import (
	"context"
	"errors"
	"flag"
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/gordcurrie/truenas-mcp/internal/truenas"
	"github.com/gordcurrie/truenas-mcp/tools"
	"github.com/modelcontextprotocol/go-sdk/mcp"
)

// version is injected at build time via -ldflags "-X main.version=<tag>".
// It defaults to "dev" so that local builds have a recognisable identifier.
var version = "dev"

func main() {
	if err := run(); err != nil {
		slog.Error("fatal", "err", err)
		os.Exit(1)
	}
}

// run contains all application logic and returns an error on failure.
// Keeping it separate from main allows defers to execute normally and
// makes the entrypoint independently testable.
func run() error {
	transport := flag.String("transport", "stdio", "Transport: stdio or http")
	addr := flag.String("addr", "localhost:8081", "Listen address (http transport only)")
	flag.Parse()

	apiURL, err := requireEnv("TRUENAS_API_URL")
	if err != nil {
		return err
	}
	apiKey, err := requireEnv("TRUENAS_API_KEY")
	if err != nil {
		return err
	}
	insecure := os.Getenv("TRUENAS_INSECURE") == "true"
	allowDestructive := os.Getenv("TRUENAS_ALLOW_DESTRUCTIVE") == "true"

	client, err := truenas.NewClient(apiURL, apiKey, insecure)
	if err != nil {
		return fmt.Errorf("creating TrueNAS client: %w", err)
	}

	connectCtx, connectCancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer connectCancel()
	if err := client.Connect(connectCtx); err != nil {
		return fmt.Errorf("connecting to TrueNAS: %w", err)
	}
	defer func() {
		if closeErr := client.Close(); closeErr != nil {
			slog.Warn("closing TrueNAS WebSocket connection", "err", closeErr)
		}
	}()

	server := mcp.NewServer(&mcp.Implementation{
		Name:    "truenas-mcp",
		Version: version,
	}, nil)

	tools.RegisterAll(server, client, tools.Config{AllowDestructive: allowDestructive})

	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	switch *transport {
	case "stdio":
		if err := server.Run(ctx, &mcp.StdioTransport{}); err != nil {
			return fmt.Errorf("stdio server: %w", err)
		}
	case "http":
		// The factory always returns the same server instance: this is intentional for a
		// stateless tool server with no per-session resources. If per-session state is
		// ever needed, the factory must create a new server per request instead.
		handler := mcp.NewStreamableHTTPHandler(func(*http.Request) *mcp.Server {
			return server
		}, nil)
		httpServer := &http.Server{
			Addr:    *addr,
			Handler: http.MaxBytesHandler(handler, 4<<20),
			// ReadTimeout must leave room for slow clients sending large bodies.
			ReadTimeout: 60 * time.Second,
			// ReadHeaderTimeout is a tighter guard against Slowloris attacks.
			ReadHeaderTimeout: 30 * time.Second,
			// WriteTimeout must exceed the TrueNAS API client timeout (30s) plus
			// any job-polling time so that in-flight responses are never cut short.
			WriteTimeout:   90 * time.Second,
			IdleTimeout:    120 * time.Second,
			MaxHeaderBytes: 1 << 20,
		}
		slog.Info("truenas-mcp listening", "addr", *addr, "transport", "http")
		go func() {
			<-ctx.Done()
			shutdownCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
			defer cancel()
			if shutdownErr := httpServer.Shutdown(shutdownCtx); shutdownErr != nil {
				slog.Warn("HTTP server shutdown error", "err", shutdownErr)
			}
		}()
		if err := httpServer.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			return fmt.Errorf("http server: %w", err)
		}
	default:
		return fmt.Errorf("unknown transport %q: must be 'stdio' or 'http'", *transport)
	}

	return nil
}

// requireEnv returns the value of the named environment variable or an error
// if it is unset or empty.
func requireEnv(name string) (string, error) {
	v := os.Getenv(name)
	if v == "" {
		return "", fmt.Errorf("required environment variable %s is not set", name)
	}
	return v, nil
}
