package tools

import (
	"context"
	"slices"
	"strings"
	"testing"

	"github.com/modelcontextprotocol/go-sdk/mcp"
)

var (
	safeToolNames = []string{
		"add_vm_device", "create_dataset", "create_snapshot", "create_vm",
		"get_app", "get_dataset", "get_pool", "get_snapshot", "get_system_info", "get_vm",
		"install_app", "install_custom_app",
		"list_apps", "list_datasets", "list_directory", "list_images", "list_interfaces",
		"list_pools", "list_snapshots", "list_vm_devices", "list_vms",
		"restart_app", "restart_vm", "rollback_app", "start_app", "start_vm",
		"stop_app", "stop_vm", "update_vm", "upgrade_app", "upgrade_summary",
	}
	destructiveToolNames = []string{
		"delete_app", "delete_snapshot", "delete_vm", "delete_vm_device", "rollback_snapshot",
	}
	// additiveToolNames are mutating tools that only create or start things
	// and therefore advertise DestructiveHint=false. Every other mutating tool
	// must advertise DestructiveHint=true.
	additiveToolNames = []string{
		"add_vm_device", "create_dataset", "create_snapshot", "create_vm",
		"install_app", "install_custom_app", "start_app", "start_vm",
	}
)

// listTools returns every tool advertised by the session, sorted by name.
func listTools(t *testing.T, cs *mcp.ClientSession) []*mcp.Tool {
	t.Helper()
	var tools []*mcp.Tool
	for tool, err := range cs.Tools(context.Background(), nil) {
		if err != nil {
			t.Fatalf("listing tools: %v", err)
		}
		tools = append(tools, tool)
	}
	slices.SortFunc(tools, func(a, b *mcp.Tool) int { return strings.Compare(a.Name, b.Name) })
	return tools
}

func toolNames(tools []*mcp.Tool) []string {
	names := make([]string, len(tools))
	for i, tool := range tools {
		names[i] = tool.Name
	}
	return names
}

func TestRegisterAllToolSet(t *testing.T) {
	t.Run("destructive tools disabled", func(t *testing.T) {
		cs, cleanup := connectTestServerWithConfig(t, &mockTruenasClient{}, Config{})
		defer cleanup()

		got := toolNames(listTools(t, cs))
		want := slices.Sorted(slices.Values(safeToolNames))
		if !slices.Equal(got, want) {
			t.Errorf("tools =\n%v\nwant\n%v", got, want)
		}
	})

	t.Run("destructive tools enabled", func(t *testing.T) {
		cs, cleanup := connectTestServerWithConfig(t, &mockTruenasClient{}, Config{AllowDestructive: true})
		defer cleanup()

		got := toolNames(listTools(t, cs))
		want := slices.Sorted(slices.Values(slices.Concat(safeToolNames, destructiveToolNames)))
		if !slices.Equal(got, want) {
			t.Errorf("tools =\n%v\nwant\n%v", got, want)
		}
	})

	t.Run("destructive tool cannot be called when disabled", func(t *testing.T) {
		cs, cleanup := connectTestServerWithConfig(t, &mockTruenasClient{}, Config{})
		defer cleanup()

		_, err := cs.CallTool(context.Background(), &mcp.CallToolParams{
			Name:      "delete_vm",
			Arguments: map[string]any{"id": 1, "confirmed": true},
		})
		if err == nil || !strings.Contains(err.Error(), "unknown tool") {
			t.Errorf("CallTool delete_vm err = %v, want unknown tool error", err)
		}
	})
}

func TestToolMetadata(t *testing.T) {
	cs, cleanup := connectTestServer(t, &mockTruenasClient{})
	defer cleanup()

	for _, tool := range listTools(t, cs) {
		t.Run(tool.Name, func(t *testing.T) {
			if tool.Description == "" {
				t.Error("missing description")
			}
			schema, ok := tool.InputSchema.(map[string]any)
			if !ok {
				t.Fatalf("input schema is %T, want object", tool.InputSchema)
			}
			if schema["type"] != "object" {
				t.Errorf("input schema type = %v, want object", schema["type"])
			}
			if tool.Annotations == nil {
				t.Fatal("missing annotations")
			}

			readOnly := strings.HasPrefix(tool.Name, "get_") || strings.HasPrefix(tool.Name, "list_") ||
				tool.Name == "upgrade_summary"
			if tool.Annotations.ReadOnlyHint != readOnly {
				t.Errorf("ReadOnlyHint = %v, want %v", tool.Annotations.ReadOnlyHint, readOnly)
			}

			if readOnly {
				return
			}
			// Per the MCP spec an omitted destructiveHint defaults to true, so
			// every mutating tool must state it explicitly.
			if tool.Annotations.DestructiveHint == nil {
				t.Fatal("mutating tool must set DestructiveHint explicitly")
			}
			wantDestructive := !slices.Contains(additiveToolNames, tool.Name)
			if *tool.Annotations.DestructiveHint != wantDestructive {
				t.Errorf("DestructiveHint = %v, want %v", *tool.Annotations.DestructiveHint, wantDestructive)
			}
		})
	}
}

// TestInputSchemaValidation pins the SDK behaviour of rejecting arguments
// that do not match a tool's input schema with an isError result, before the
// handler runs.
func TestInputSchemaValidation(t *testing.T) {
	tests := []struct {
		name string
		args map[string]any
		want string
	}{
		{"missing required property", map[string]any{"confirmed": true}, `missing properties: ["id"]`},
		{"wrong type", map[string]any{"id": "abc", "confirmed": true}, `want "integer"`},
		{"unknown property", map[string]any{"id": 1, "confirmed": true, "bogus": 1}, `additional properties ["bogus"]`},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			called := false
			mock := &mockTruenasClient{deleteVMFn: func(context.Context, int) error {
				called = true
				return nil
			}}
			cs, cleanup := connectTestServer(t, mock)
			defer cleanup()

			res := callTool(t, cs, "delete_vm", tt.args)
			assertError(t, res, tt.want)
			if called {
				t.Error("handler called despite invalid arguments")
			}
		})
	}
}
