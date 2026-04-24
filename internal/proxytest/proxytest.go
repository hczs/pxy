package proxytest

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"time"
)

type Result struct {
	IP      string `json:"ip"`
	Country string `json:"country"`
	City    string `json:"city"`
}

func Run(ctx context.Context, url string, client *http.Client) (Result, error) {
	if client == nil {
		client = &http.Client{Timeout: 5 * time.Second}
	}
	// context 提供超时和取消能力，类似 Java CompletableFuture/HTTP 调用里的 timeout 控制。
	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return Result{}, fmt.Errorf("create request: %w", err)
	}
	resp, err := client.Do(req)
	if err != nil {
		return Result{}, fmt.Errorf("test proxy: %w", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode < 200 || resp.StatusCode > 299 {
		return Result{}, fmt.Errorf("test proxy: unexpected status %s", resp.Status)
	}
	var result Result
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return Result{}, fmt.Errorf("decode ipinfo response: %w", err)
	}
	return result, nil
}
