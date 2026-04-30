package config

import (
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"testing"
)

func TestValidateRejectsInvalidValues(t *testing.T) {
	tests := []struct {
		name string
		cfg  Config
		want string
	}{
		{
			name: "bad host",
			cfg:  Config{Proxy: Proxy{HTTP: Endpoint{Enabled: true, Host: "bad host", Port: 7890}}},
			want: "proxy.http.host",
		},
		{
			name: "bad port",
			cfg:  Config{Proxy: Proxy{HTTP: Endpoint{Enabled: true, Host: "127.0.0.1", Port: 70000}}},
			want: "proxy.http.port",
		},
		{
			name: "no enabled proxy",
			cfg:  Config{Proxy: Proxy{}},
			want: "at least one proxy endpoint",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.cfg.Validate()
			if err == nil || !strings.Contains(err.Error(), tt.want) {
				t.Fatalf("Validate() = %v, want containing %q", err, tt.want)
			}
		})
	}
}

func TestSaveLoadRoundTrip(t *testing.T) {
	path := filepath.Join(t.TempDir(), "config.yaml")
	cfg := Config{Proxy: Proxy{
		HTTP:       Endpoint{Enabled: true, Host: "127.0.0.1", Port: 7890},
		HTTPS:      Endpoint{Enabled: true, Host: "127.0.0.1", Port: 7890},
		SOCKS5:     Endpoint{Enabled: true, Host: "localhost", Port: 7891},
		AutoDetect: true,
		Source:     "manual",
	}}

	if err := Save(path, cfg); err != nil {
		t.Fatalf("Save() error = %v", err)
	}
	got, err := Load(path)
	if err != nil {
		t.Fatalf("Load() error = %v", err)
	}
	if got.Proxy.HTTP.Port != 7890 || got.Proxy.SOCKS5.Host != "localhost" || !got.Proxy.AutoDetect {
		t.Fatalf("Load() = %+v", got)
	}
	if runtime.GOOS != "windows" {
		if info, err := os.Stat(path); err != nil || info.Mode().Perm() != 0o600 {
			t.Fatalf("config mode = %v, err=%v; want 0600", info.Mode().Perm(), err)
		}
	}
}

func TestParseAddress(t *testing.T) {
	ep, err := ParseAddress("127.0.0.1:7890")
	if err != nil {
		t.Fatalf("ParseAddress() error = %v", err)
	}
	if ep.Host != "127.0.0.1" || ep.Port != 7890 || !ep.Enabled {
		t.Fatalf("endpoint = %+v", ep)
	}
}
