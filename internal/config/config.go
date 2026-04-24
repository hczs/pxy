package config

import (
	"fmt"
	"net"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	"gopkg.in/yaml.v3"
)

// Config 是磁盘上的完整配置；结构体 tag 用来声明 YAML 字段名。
type Config struct {
	Proxy Proxy `yaml:"proxy"`
}

type Proxy struct {
	HTTP       Endpoint `yaml:"http"`
	HTTPS      Endpoint `yaml:"https"`
	SOCKS5     Endpoint `yaml:"socks5"`
	AutoDetect bool     `yaml:"auto_detect"`
	Source     string   `yaml:"source,omitempty"`
}

type Endpoint struct {
	Enabled bool   `yaml:"enabled"`
	Host    string `yaml:"host"`
	Port    int    `yaml:"port"`
}

func DefaultPath(home string) string {
	return filepath.Join(home, ".pxy", "config.yaml")
}

func Load(path string) (Config, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return Config{}, fmt.Errorf("read config: %w", err)
	}
	var cfg Config
	if err := yaml.Unmarshal(data, &cfg); err != nil {
		return Config{}, fmt.Errorf("parse config: %w", err)
	}
	if err := cfg.Validate(); err != nil {
		return Config{}, err
	}
	return cfg, nil
}

func Save(path string, cfg Config) error {
	if err := cfg.Validate(); err != nil {
		return err
	}
	data, err := yaml.Marshal(cfg)
	if err != nil {
		return fmt.Errorf("marshal config: %w", err)
	}
	if err := os.MkdirAll(filepath.Dir(path), 0o700); err != nil {
		return fmt.Errorf("create config dir: %w", err)
	}
	if err := os.WriteFile(path, data, 0o600); err != nil {
		return fmt.Errorf("write config: %w", err)
	}
	return nil
}

func (c Config) Validate() error {
	enabled := 0
	for name, endpoint := range map[string]Endpoint{
		"http":   c.Proxy.HTTP,
		"https":  c.Proxy.HTTPS,
		"socks5": c.Proxy.SOCKS5,
	} {
		if !endpoint.Enabled {
			continue
		}
		enabled++
		if err := validateEndpoint(name, endpoint); err != nil {
			return err
		}
	}
	if enabled == 0 {
		return fmt.Errorf("at least one proxy endpoint must be enabled")
	}
	return nil
}

func validateEndpoint(name string, endpoint Endpoint) error {
	if !validHost(endpoint.Host) {
		return fmt.Errorf("proxy.%s.host is invalid: %q", name, endpoint.Host)
	}
	if endpoint.Port < 1 || endpoint.Port > 65535 {
		return fmt.Errorf("proxy.%s.port must be 1-65535: %d", name, endpoint.Port)
	}
	return nil
}

func validHost(host string) bool {
	if host == "localhost" {
		return true
	}
	if ip := net.ParseIP(host); ip != nil {
		return true
	}
	labels := strings.Split(host, ".")
	for _, label := range labels {
		if label == "" || len(label) > 63 || strings.HasPrefix(label, "-") || strings.HasSuffix(label, "-") {
			return false
		}
		for _, r := range label {
			if (r >= 'a' && r <= 'z') || (r >= 'A' && r <= 'Z') || (r >= '0' && r <= '9') || r == '-' {
				continue
			}
			return false
		}
	}
	return true
}

func ParseAddress(value string) (Endpoint, error) {
	host, portText, err := net.SplitHostPort(value)
	if err != nil {
		return Endpoint{}, fmt.Errorf("parse address %q: %w", value, err)
	}
	port, err := strconv.Atoi(portText)
	if err != nil {
		return Endpoint{}, fmt.Errorf("parse port %q: %w", portText, err)
	}
	ep := Endpoint{Enabled: true, Host: host, Port: port}
	if err := validateEndpoint("address", ep); err != nil {
		return Endpoint{}, err
	}
	return ep, nil
}

func MustHTTPSFromHTTP(http Endpoint) Endpoint {
	if !http.Enabled {
		return Endpoint{}
	}
	return Endpoint{Enabled: true, Host: http.Host, Port: http.Port}
}
