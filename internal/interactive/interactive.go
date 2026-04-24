package interactive

import (
	"bufio"
	"fmt"
	"io"
	"strings"

	"github.com/hczs/px/internal/config"
)

func Confirm(in io.Reader, out io.Writer, prompt string) (bool, error) {
	fmt.Fprintf(out, "%s [Y/n] ", prompt)
	text, err := bufio.NewReader(in).ReadString('\n')
	if err != nil && err != io.EOF {
		return false, fmt.Errorf("read confirmation: %w", err)
	}
	answer := strings.ToLower(strings.TrimSpace(text))
	return answer == "" || answer == "y" || answer == "yes", nil
}

func ManualConfig(in io.Reader, out io.Writer) (config.Config, error) {
	reader := bufio.NewReader(in)
	http, err := askEndpoint(reader, out, "HTTP代理地址 [默认: 127.0.0.1:7890]: ", "127.0.0.1:7890")
	if err != nil {
		return config.Config{}, err
	}
	https, err := askEndpoint(reader, out, "HTTPS代理地址 [默认与HTTP相同，回车跳过]: ", fmt.Sprintf("%s:%d", http.Host, http.Port))
	if err != nil {
		return config.Config{}, err
	}
	socksEnabled, err := askYesNo(reader, out, "是否启用SOCKS5代理？ [y/n, 默认 y]: ", true)
	if err != nil {
		return config.Config{}, err
	}
	var socks config.Endpoint
	if socksEnabled {
		socks, err = askEndpoint(reader, out, "SOCKS5代理地址 [默认: 127.0.0.1:7891]: ", "127.0.0.1:7891")
		if err != nil {
			return config.Config{}, err
		}
	}
	cfg := config.Config{Proxy: config.Proxy{HTTP: http, HTTPS: https, SOCKS5: socks, AutoDetect: false, Source: "手动配置(~/.px/config.yaml)"}}
	if err := cfg.Validate(); err != nil {
		return config.Config{}, err
	}
	return cfg, nil
}

func askEndpoint(reader *bufio.Reader, out io.Writer, prompt, defaultValue string) (config.Endpoint, error) {
	fmt.Fprint(out, prompt)
	text, err := reader.ReadString('\n')
	if err != nil && err != io.EOF {
		return config.Endpoint{}, fmt.Errorf("read proxy address: %w", err)
	}
	value := strings.TrimSpace(text)
	if value == "" {
		value = defaultValue
	}
	ep, err := config.ParseAddress(value)
	if err != nil {
		return config.Endpoint{}, err
	}
	return ep, nil
}

func askYesNo(reader *bufio.Reader, out io.Writer, prompt string, defaultValue bool) (bool, error) {
	fmt.Fprint(out, prompt)
	text, err := reader.ReadString('\n')
	if err != nil && err != io.EOF {
		return false, fmt.Errorf("read yes/no: %w", err)
	}
	answer := strings.ToLower(strings.TrimSpace(text))
	if answer == "" {
		return defaultValue, nil
	}
	return answer == "y" || answer == "yes", nil
}
