package proxytest

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestRunParsesIPInfo(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`{"ip":"1.2.3.4","country":"US","city":"San Francisco"}`))
	}))
	defer server.Close()

	got, err := Run(context.Background(), server.URL, server.Client())
	if err != nil {
		t.Fatalf("Run() error = %v", err)
	}
	if got.IP != "1.2.3.4" || got.Country != "US" || got.City != "San Francisco" {
		t.Fatalf("Run() = %+v", got)
	}
}
