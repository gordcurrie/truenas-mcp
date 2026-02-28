// Command truenas-mcp is an MCP server that exposes TrueNAS SCALE operations
// as MCP tools.
//
// Required environment variables:
//
//	TRUENAS_API_URL  Base URL of the TrueNAS SCALE API (e.g. https://truenas.local/api/v2.0)
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

	server := mcp.NewServer(&mcp.Implementation{
		Name:    "truenas-mcp",
		Version: "v0.1.0",
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
		handler := mcp.NewStreamableHTTPHandler(func(*http.Request) *mcp.Server {
			return server
		}, nil)
		httpServer := &http.Server{
			Addr:              *addr,
			Handler:           handler,
			ReadHeaderTimeout: 30 * time.Second,
		}
		slog.Info("truenas-mcp listening", "addr", *addr, "transport", "http")
		go func() {
			<-ctx.Done()
			if shutdownErr := httpServer.Shutdown(context.Background()); shutdownErr != nil {
				slog.Warn("HTTP server shutdown error", "err", shutdownErr)
			}
		}()
		if err := httpServer.ListenAndServe(); err != nil && err != http.ErrServerClosed {
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
