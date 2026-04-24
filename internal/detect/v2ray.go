package detect

import (
	"encoding/json"
	"fmt"
	"os"

	"github.com/hczs/px/internal/config"
)

type v2rayConfig struct {
	HTTPPort  int `json:"httpPort"`
	SocksPort int `json:"socksPort"`
	LocalPort int `json:"localPort"`
}

func ParseV2Ray(name, path string, priority int) Result {
	data, err := os.ReadFile(path)
	if err != nil {
		return Result{Name: name, Path: path, Priority: priority, Reason: fmt.Sprintf("read config: %v", err)}
	}
	var raw v2rayConfig
	if err := json.Unmarshal(data, &raw); err != nil {
		return Result{Name: name, Path: path, Found: true, Priority: priority, Reason: fmt.Sprintf("parse json: %v", err)}
	}
	cfg := config.Config{Proxy: config.Proxy{AutoDetect: true, Source: "自动检测(" + name + ")"}}
	if raw.HTTPPort > 0 {
		cfg.Proxy.HTTP = config.Endpoint{Enabled: true, Host: "127.0.0.1", Port: raw.HTTPPort}
		cfg.Proxy.HTTPS = config.MustHTTPSFromHTTP(cfg.Proxy.HTTP)
	}
	if raw.LocalPort > 0 && !cfg.Proxy.HTTP.Enabled {
		cfg.Proxy.HTTP = config.Endpoint{Enabled: true, Host: "127.0.0.1", Port: raw.LocalPort}
		cfg.Proxy.HTTPS = config.MustHTTPSFromHTTP(cfg.Proxy.HTTP)
	}
	if raw.SocksPort > 0 {
		cfg.Proxy.SOCKS5 = config.Endpoint{Enabled: true, Host: "127.0.0.1", Port: raw.SocksPort}
	}
	return resultFromConfig(name, path, priority, cfg)
}
