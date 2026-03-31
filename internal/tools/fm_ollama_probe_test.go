package tools

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/davidl71/exarp-go/internal/config"
)

func TestOllamaReachableForFM_httptest(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/api/tags" {
			http.NotFound(w, r)
			return
		}
		w.WriteHeader(http.StatusOK)
	}))
	defer srv.Close()

	orig := config.GetGlobalConfig()
	cfg := *orig
	cfg.Tools.Ollama.DefaultHost = srv.URL
	config.SetGlobalConfig(&cfg)
	defer config.SetGlobalConfig(orig)

	resetOllamaReachabilityCacheForTest()
	defer resetOllamaReachabilityCacheForTest()

	if !OllamaReachableForFM() {
		t.Fatal("expected OllamaReachableForFM true when /api/tags returns 200")
	}
	if !OllamaReachableForFM() {
		t.Fatal("expected second call to use cache")
	}
}

func TestOllamaReachableForFM_hostChangeInvalidatesCache(t *testing.T) {
	srvOK := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/api/tags" {
			w.WriteHeader(http.StatusOK)
			return
		}
		http.NotFound(w, r)
	}))
	defer srvOK.Close()

	orig := config.GetGlobalConfig()
	cfg := *orig
	cfg.Tools.Ollama.DefaultHost = srvOK.URL
	config.SetGlobalConfig(&cfg)
	defer config.SetGlobalConfig(orig)

	resetOllamaReachabilityCacheForTest()
	defer resetOllamaReachabilityCacheForTest()

	if !OllamaReachableForFM() {
		t.Fatal("first probe should succeed")
	}

	cur := config.GetGlobalConfig()
	cfg2 := *cur
	cfg2.Tools.Ollama.DefaultHost = "http://127.0.0.1:9"
	config.SetGlobalConfig(&cfg2)

	if OllamaReachableForFM() {
		t.Fatal("after host change to unreachable port, expected false")
	}
}
