package detect

import (
	"bufio"
	"fmt"
	"os"
	"strings"

	"github.com/hczs/pxy/internal/config"
)

func ParseSurge(path string) Result {
	file, err := os.Open(path)
	if err != nil {
		return Result{Name: "Surge", Path: path, Priority: 20, Reason: fmt.Sprintf("read config: %v", err)}
	}
	defer file.Close()

	cfg := config.Config{Proxy: config.Proxy{AutoDetect: true, Source: "自动检测(Surge)"}}
	inGeneral := false
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line == "" || strings.HasPrefix(line, "#") || strings.HasPrefix(line, ";") {
			continue
		}
		if strings.HasPrefix(line, "[") && strings.HasSuffix(line, "]") {
			inGeneral = strings.EqualFold(line, "[General]")
			continue
		}
		if !inGeneral {
			continue
		}
		key, value, ok := strings.Cut(line, "=")
		if !ok {
			continue
		}
		ep, err := config.ParseAddress(strings.TrimSpace(value))
		if err != nil {
			continue
		}
		switch strings.TrimSpace(key) {
		case "http-listen":
			cfg.Proxy.HTTP = ep
			cfg.Proxy.HTTPS = config.MustHTTPSFromHTTP(ep)
		case "socks5-listen":
			cfg.Proxy.SOCKS5 = ep
		}
	}
	if err := scanner.Err(); err != nil {
		return Result{Name: "Surge", Path: path, Found: true, Priority: 20, Reason: fmt.Sprintf("scan config: %v", err)}
	}
	return resultFromConfig("Surge", path, 20, cfg)
}
