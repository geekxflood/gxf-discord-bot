package config

import (
	"os"
	"path/filepath"
	"testing"
)

func TestNewManager(t *testing.T) {
	schema := []byte(`
package schema

#Config: {
	bot: {
		prefix: string | *"!"
		token?: string
	}
	actions: [...{
		name: string
		type: string
	}]
}
`)

	tests := []struct {
		name        string
		config      string
		wantErr     bool
		description string
	}{
		{
			name: "valid config",
			config: `
bot:
  prefix: "!"
  token: "test-token"
actions:
  - name: "ping"
    type: "command"
`,
			wantErr:     false,
			description: "should load valid config successfully",
		},
		{
			name: "empty config",
			config: `
bot:
  prefix: "!"
actions: []
`,
			wantErr:     false,
			description: "should handle empty actions array",
		},
		{
			name:        "invalid yaml",
			config:      `invalid: yaml: syntax:`,
			wantErr:     true,
			description: "should fail on invalid YAML",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create temp config file
			tmpDir := t.TempDir()
			configPath := filepath.Join(tmpDir, "config.yaml")
			err := os.WriteFile(configPath, []byte(tt.config), 0644)
			if err != nil {
				t.Fatalf("failed to write test config: %v", err)
			}

			// Create manager
			mgr, err := NewManager(Options{
				SchemaContent: schema,
				ConfigPath:    configPath,
			})

			if tt.wantErr {
				if err == nil {
					t.Errorf("expected error but got none")
				}
				return
			}

			if err != nil {
				t.Errorf("unexpected error: %v", err)
				return
			}

			if mgr == nil {
				t.Error("expected manager but got nil")
			}
		})
	}
}

func TestManager_GetString(t *testing.T) {
	schema := []byte(`
package schema

#Config: {
	bot: {
		prefix: string
		token: string
	}
}
`)

	config := `
bot:
  prefix: "!"
  token: "test-token"
`

	tmpDir := t.TempDir()
	configPath := filepath.Join(tmpDir, "config.yaml")
	os.WriteFile(configPath, []byte(config), 0644)

	mgr, err := NewManager(Options{
		SchemaContent: schema,
		ConfigPath:    configPath,
	})
	if err != nil {
		t.Fatalf("failed to create manager: %v", err)
	}

	tests := []struct {
		name         string
		key          string
		defaultValue string
		want         string
		wantErr      bool
	}{
		{
			name:         "existing key",
			key:          "bot.prefix",
			defaultValue: "?",
			want:         "!",
			wantErr:      false,
		},
		{
			name:         "missing key with default",
			key:          "bot.missing",
			defaultValue: "default",
			want:         "default",
			wantErr:      false,
		},
		{
			name:         "nested key",
			key:          "bot.token",
			defaultValue: "",
			want:         "test-token",
			wantErr:      false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := mgr.GetString(tt.key, tt.defaultValue)
			if (err != nil) != tt.wantErr {
				t.Errorf("GetString() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != tt.want {
				t.Errorf("GetString() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestManager_GetInt(t *testing.T) {
	schema := []byte(`
package schema

#Config: {
	server: {
		port: int
	}
}
`)

	config := `
server:
  port: 8080
`

	tmpDir := t.TempDir()
	configPath := filepath.Join(tmpDir, "config.yaml")
	os.WriteFile(configPath, []byte(config), 0644)

	mgr, err := NewManager(Options{
		SchemaContent: schema,
		ConfigPath:    configPath,
	})
	if err != nil {
		t.Fatalf("failed to create manager: %v", err)
	}

	tests := []struct {
		name         string
		key          string
		defaultValue int
		want         int
	}{
		{
			name:         "existing int",
			key:          "server.port",
			defaultValue: 3000,
			want:         8080,
		},
		{
			name:         "missing with default",
			key:          "server.timeout",
			defaultValue: 30,
			want:         30,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := mgr.GetInt(tt.key, tt.defaultValue)
			if err != nil {
				t.Errorf("GetInt() error = %v", err)
				return
			}
			if got != tt.want {
				t.Errorf("GetInt() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestManager_GetBool(t *testing.T) {
	schema := []byte(`
package schema

#Config: {
	features: {
		enabled: bool
	}
}
`)

	config := `
features:
  enabled: true
`

	tmpDir := t.TempDir()
	configPath := filepath.Join(tmpDir, "config.yaml")
	os.WriteFile(configPath, []byte(config), 0644)

	mgr, err := NewManager(Options{
		SchemaContent: schema,
		ConfigPath:    configPath,
	})
	if err != nil {
		t.Fatalf("failed to create manager: %v", err)
	}

	tests := []struct {
		name         string
		key          string
		defaultValue bool
		want         bool
	}{
		{
			name:         "existing bool true",
			key:          "features.enabled",
			defaultValue: false,
			want:         true,
		},
		{
			name:         "missing with default",
			key:          "features.debug",
			defaultValue: false,
			want:         false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := mgr.GetBool(tt.key, tt.defaultValue)
			if err != nil {
				t.Errorf("GetBool() error = %v", err)
				return
			}
			if got != tt.want {
				t.Errorf("GetBool() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestManager_GetStringSlice(t *testing.T) {
	schema := []byte(`
package schema

#Config: {
	servers: [...string]
}
`)

	config := `
servers:
  - "server1"
  - "server2"
  - "server3"
`

	tmpDir := t.TempDir()
	configPath := filepath.Join(tmpDir, "config.yaml")
	os.WriteFile(configPath, []byte(config), 0644)

	mgr, err := NewManager(Options{
		SchemaContent: schema,
		ConfigPath:    configPath,
	})
	if err != nil {
		t.Fatalf("failed to create manager: %v", err)
	}

	got, err := mgr.GetStringSlice("servers")
	if err != nil {
		t.Errorf("GetStringSlice() error = %v", err)
		return
	}

	want := []string{"server1", "server2", "server3"}
	if len(got) != len(want) {
		t.Errorf("GetStringSlice() length = %v, want %v", len(got), len(want))
		return
	}

	for i, v := range got {
		if v != want[i] {
			t.Errorf("GetStringSlice()[%d] = %v, want %v", i, v, want[i])
		}
	}
}

func TestManager_Exists(t *testing.T) {
	schema := []byte(`
package schema

#Config: {
	bot: {
		prefix: string
	}
}
`)

	config := `
bot:
  prefix: "!"
`

	tmpDir := t.TempDir()
	configPath := filepath.Join(tmpDir, "config.yaml")
	os.WriteFile(configPath, []byte(config), 0644)

	mgr, err := NewManager(Options{
		SchemaContent: schema,
		ConfigPath:    configPath,
	})
	if err != nil {
		t.Fatalf("failed to create manager: %v", err)
	}

	tests := []struct {
		name string
		key  string
		want bool
	}{
		{
			name: "existing key",
			key:  "bot.prefix",
			want: true,
		},
		{
			name: "missing key",
			key:  "bot.token",
			want: false,
		},
		{
			name: "existing parent",
			key:  "bot",
			want: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := mgr.Exists(tt.key)
			if got != tt.want {
				t.Errorf("Exists() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestManager_ArrayAccess(t *testing.T) {
	schema := []byte(`
package schema

#Config: {
	actions: [...{
		name: string
		type: string
	}]
}
`)

	config := `
actions:
  - name: "ping"
    type: "command"
  - name: "help"
    type: "command"
`

	tmpDir := t.TempDir()
	configPath := filepath.Join(tmpDir, "config.yaml")
	os.WriteFile(configPath, []byte(config), 0644)

	mgr, err := NewManager(Options{
		SchemaContent: schema,
		ConfigPath:    configPath,
	})
	if err != nil {
		t.Fatalf("failed to create manager: %v", err)
	}

	tests := []struct {
		name    string
		key     string
		want    string
		wantErr bool
	}{
		{
			name:    "first action name",
			key:     "actions[0].name",
			want:    "ping",
			wantErr: false,
		},
		{
			name:    "second action name",
			key:     "actions[1].name",
			want:    "help",
			wantErr: false,
		},
		{
			name:    "first action type",
			key:     "actions[0].type",
			want:    "command",
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := mgr.GetString(tt.key, "")
			if (err != nil) != tt.wantErr {
				t.Errorf("GetString() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != tt.want {
				t.Errorf("GetString() = %v, want %v", got, tt.want)
			}
		})
	}
}
