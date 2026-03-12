package factory

import (
	"testing"

	"github.com/davidl71/mcp-go-core/pkg/mcp/config"
)

func TestNewServer(t *testing.T) {
	tests := []struct {
		name          string
		frameworkType config.FrameworkType
		serverName    string
		version       string
		wantErr       bool
	}{
		{
			name:          "valid go-sdk",
			frameworkType: config.FrameworkGoSDK,
			serverName:    "test-server",
			version:       "1.0.0",
			wantErr:       false,
		},
		{
			name:          "invalid framework",
			frameworkType: config.FrameworkType("invalid"),
			serverName:    "test-server",
			version:       "1.0.0",
			wantErr:       true,
		},
		{
			name:          "empty server name",
			frameworkType: config.FrameworkGoSDK,
			serverName:    "",
			version:       "1.0.0",
			wantErr:       false, // Empty name is allowed (will use default)
		},
		{
			name:          "empty version",
			frameworkType: config.FrameworkGoSDK,
			serverName:    "test-server",
			version:       "",
			wantErr:       false, // Empty version is allowed
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			server, err := NewServer(tt.frameworkType, tt.serverName, tt.version)
			if (err != nil) != tt.wantErr {
				t.Errorf("NewServer() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !tt.wantErr {
				if server == nil {
					t.Error("NewServer() returned nil server")
					return
				}
				// Verify server name
				if server.GetName() != tt.serverName {
					t.Errorf("server.GetName() = %q, want %q", server.GetName(), tt.serverName)
				}
			}
		})
	}
}

func TestNewServerFromConfig(t *testing.T) {
	tests := []struct {
		name    string
		cfg     *config.BaseConfig
		wantErr bool
	}{
		{
			name: "valid config with go-sdk",
			cfg: &config.BaseConfig{
				Framework: config.FrameworkGoSDK,
				Name:      "test-server",
				Version:   "1.0.0",
			},
			wantErr: false,
		},
		{
			name: "invalid framework",
			cfg: &config.BaseConfig{
				Framework: config.FrameworkType("invalid"),
				Name:      "test-server",
				Version:   "1.0.0",
			},
			wantErr: true,
		},
		{
			name: "nil config",
			cfg:  nil,
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			server, err := NewServerFromConfig(tt.cfg)
			if (err != nil) != tt.wantErr {
				t.Errorf("NewServerFromConfig() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !tt.wantErr {
				if server == nil {
					t.Error("NewServerFromConfig() returned nil server")
					return
				}
				if tt.cfg != nil && server.GetName() != tt.cfg.Name {
					t.Errorf("server.GetName() = %q, want %q", server.GetName(), tt.cfg.Name)
				}
			}
		})
	}
}

func TestNewServerFromConfig_ServerName(t *testing.T) {
	cfg := &config.BaseConfig{
		Framework: config.FrameworkGoSDK,
		Name:      "custom-server",
		Version:   "2.0.0",
	}

	server, err := NewServerFromConfig(cfg)
	if err != nil {
		t.Fatalf("NewServerFromConfig() error = %v", err)
	}
	if server == nil {
		t.Fatal("NewServerFromConfig() returned nil server")
	}
	if server.GetName() != "custom-server" {
		t.Errorf("server.GetName() = %q, want %q", server.GetName(), "custom-server")
	}
}
