package detect

import (
	"fmt"
	"os"

	"github.com/hczs/pxy/internal/config"
	"gopkg.in/yaml.v3"
)

type clashFile struct {
	MixedPort int `yaml:"mixed-port"`
	HTTPPort  int `yaml:"port"`
	SocksPort int `yaml:"socks-port"`
}

func ParseClash(path string) Result {
	data, err := os.ReadFile(path)
	if err != nil {
		return Result{Name: "Clash", Path: path, Priority: 10, Reason: fmt.Sprintf("read config: %v", err)}
	}
	var raw clashFile
	if err := yaml.Unmarshal(data, &raw); err != nil {
		return Result{Name: "Clash", Path: path, Found: true, Priority: 10, Reason: fmt.Sprintf("parse yaml: %v", err)}
	}
	cfg := config.Config{Proxy: config.Proxy{AutoDetect: true, Source: "自动检测(Clash)"}}
	if raw.MixedPort > 0 {
		cfg.Proxy.HTTP = config.Endpoint{Enabled: true, Host: "127.0.0.1", Port: raw.MixedPort}
		cfg.Proxy.HTTPS = config.Endpoint{Enabled: true, Host: "127.0.0.1", Port: raw.MixedPort}
		cfg.Proxy.SOCKS5 = config.Endpoint{Enabled: true, Host: "127.0.0.1", Port: raw.MixedPort}
	}
	if raw.HTTPPort > 0 {
		cfg.Proxy.HTTP = config.Endpoint{Enabled: true, Host: "127.0.0.1", Port: raw.HTTPPort}
		cfg.Proxy.HTTPS = config.Endpoint{Enabled: true, Host: "127.0.0.1", Port: raw.HTTPPort}
	}
	if raw.SocksPort > 0 {
		cfg.Proxy.SOCKS5 = config.Endpoint{Enabled: true, Host: "127.0.0.1", Port: raw.SocksPort}
	}
	return resultFromConfig("Clash", path, 10, cfg)
}
