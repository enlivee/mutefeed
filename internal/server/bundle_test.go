package server

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync"
	"testing"

	"github.com/enlivee/mutefeed/internal/feed"
)

func TestAPIBundles(t *testing.T) {
	srv, _ := setupTestServer(t)
	handler := srv.Routes()

	req := httptest.NewRequest("GET", "/api/bundles", nil)
	w := httptest.NewRecorder()
	handler.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", w.Code)
	}

	contentType := w.Header().Get("Content-Type")
	if !strings.Contains(contentType, "application/json") {
		t.Errorf("expected application/json, got %q", contentType)
	}

	var bundles []feed.Bundle
	if err := json.Unmarshal(w.Body.Bytes(), &bundles); err != nil {
		t.Fatalf("failed to decode bundles JSON: %v", err)
	}

	if len(bundles) < 2 {
		t.Fatalf("expected at least 2 bundles, got %d", len(bundles))
	}

	foundTech := false
	foundRuTech := false
	for _, b := range bundles {
		if b.ID == "tech" {
			foundTech = true
			if len(b.URLs) != 5 {
				t.Errorf("expected 5 URLs in tech bundle, got %d", len(b.URLs))
			}
		}
		if b.ID == "ru-tech" {
			foundRuTech = true
			if len(b.URLs) != 3 {
				t.Errorf("expected 3 URLs in ru-tech bundle, got %d", len(b.URLs))
			}
		}
	}

	if !foundTech {
		t.Error("tech bundle not found in /api/bundles output")
	}
	if !foundRuTech {
		t.Error("ru-tech bundle not found in /api/bundles output")
	}
}

func TestInstallScript(t *testing.T) {
	srv, _ := setupTestServer(t)
	handler := srv.Routes()

	req := httptest.NewRequest("GET", "/install.sh", nil)
	w := httptest.NewRecorder()
	handler.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", w.Code)
	}

	contentType := w.Header().Get("Content-Type")
	if contentType != "text/plain; charset=utf-8" {
		t.Errorf("expected 'text/plain; charset=utf-8', got %q", contentType)
	}

	body := w.Body.String()
	if !strings.Contains(body, "#!/usr/bin/env bash") {
		t.Errorf("expected bash shebang in install script")
	}
	if !strings.Contains(body, `BINARY="mutefeed"`) {
		t.Errorf("expected BINARY=\"mutefeed\" in install script")
	}
}

func TestFeedBundle(t *testing.T) {
	srv, upstream := setupTestServer(t)
	t.Cleanup(feed.ResetBundles)

	// Register bundle pointing to test upstream
	feed.RegisterBundle(feed.Bundle{
		ID:          "tech",
		Name:        "Test Tech",
		Description: "Test tech bundle",
		URLs:        []string{upstream.URL},
	})

	handler := srv.Routes()

	// 1. Initial request (MISS)
	req := httptest.NewRequest("GET", "/feed?bundle=tech", nil)
	w := httptest.NewRecorder()
	handler.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d: %s", w.Code, w.Body.String())
	}

	if cacheHeader := w.Header().Get("X-Cache"); cacheHeader != "MISS" {
		t.Errorf("expected X-Cache: MISS on first request, got %q", cacheHeader)
	}

	body := w.Body.String()
	if !strings.Contains(body, "[Clean] MuteFeed · Clean Tech Digest") {
		t.Errorf("expected clean digest title in RSS output, got:\n%s", body)
	}

	// 2. Second request (HIT)
	w2 := httptest.NewRecorder()
	handler.ServeHTTP(w2, req)

	if w2.Code != http.StatusOK {
		t.Fatalf("expected 200 on cache hit, got %d", w2.Code)
	}
	if cacheHeader := w2.Header().Get("X-Cache"); cacheHeader != "HIT" {
		t.Errorf("expected X-Cache: HIT on second request, got %q", cacheHeader)
	}
}

func TestFeedBundleNotFound(t *testing.T) {
	srv, _ := setupTestServer(t)
	handler := srv.Routes()

	req := httptest.NewRequest("GET", "/feed?bundle=nonexistent-bundle", nil)
	w := httptest.NewRecorder()
	handler.ServeHTTP(w, req)

	if w.Code != http.StatusBadRequest {
		t.Fatalf("expected 400 Bad Request, got %d", w.Code)
	}
}

func TestAPITestBundle(t *testing.T) {
	srv, upstream := setupTestServer(t)
	t.Cleanup(feed.ResetBundles)

	feed.RegisterBundle(feed.Bundle{
		ID:   "tech",
		URLs: []string{upstream.URL},
	})

	handler := srv.Routes()

	req := httptest.NewRequest("GET", "/api/test?bundle=tech", nil)
	w := httptest.NewRecorder()
	handler.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d: %s", w.Code, w.Body.String())
	}

	var resp map[string]interface{}
	if err := json.Unmarshal(w.Body.Bytes(), &resp); err != nil {
		t.Fatalf("failed to decode JSON: %v", err)
	}

	if title, ok := resp["title"].(string); !ok || title != "MuteFeed · Clean Tech Digest" {
		t.Errorf("expected title 'MuteFeed · Clean Tech Digest', got %v", resp["title"])
	}
	if _, ok := resp["decisions"]; !ok {
		t.Error("expected decisions in test response")
	}
}

func TestBadgeBundle(t *testing.T) {
	srv, upstream := setupTestServer(t)
	t.Cleanup(feed.ResetBundles)

	feed.RegisterBundle(feed.Bundle{
		ID:   "tech",
		URLs: []string{upstream.URL},
	})

	handler := srv.Routes()

	req := httptest.NewRequest("GET", "/badge?bundle=tech", nil)
	w := httptest.NewRecorder()
	handler.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", w.Code)
	}
	if ct := w.Header().Get("Content-Type"); !strings.Contains(ct, "image/svg+xml") {
		t.Errorf("expected SVG content type, got %q", ct)
	}
	if !strings.Contains(w.Body.String(), "<svg") {
		t.Errorf("expected SVG output, got %s", w.Body.String())
	}
}

func TestBundleServerConcurrencyWithRace(t *testing.T) {
	srv, upstream := setupTestServer(t)
	srv.limiter = NewIPRateLimiter(10000, 1000)
	t.Cleanup(feed.ResetBundles)

	feed.RegisterBundle(feed.Bundle{
		ID:   "tech",
		URLs: []string{upstream.URL},
	})

	handler := srv.Routes()

	var wg sync.WaitGroup
	for i := 0; i < 15; i++ {
		wg.Add(4)
		clientIP := fmt.Sprintf("10.10.1.%d:5000", i)

		go func(ip string) {
			defer wg.Done()
			req := httptest.NewRequest("GET", "/api/bundles", nil)
			req.RemoteAddr = ip
			w := httptest.NewRecorder()
			handler.ServeHTTP(w, req)
			if w.Code != http.StatusOK {
				t.Errorf("bundles handler failed: %d", w.Code)
			}
		}(clientIP)

		go func(ip string) {
			defer wg.Done()
			req := httptest.NewRequest("GET", "/feed?bundle=tech", nil)
			req.RemoteAddr = ip
			w := httptest.NewRecorder()
			handler.ServeHTTP(w, req)
			if w.Code != http.StatusOK {
				t.Errorf("feed bundle handler failed: %d", w.Code)
			}
		}(clientIP)

		go func(ip string) {
			defer wg.Done()
			req := httptest.NewRequest("GET", "/api/test?bundle=tech", nil)
			req.RemoteAddr = ip
			w := httptest.NewRecorder()
			handler.ServeHTTP(w, req)
			if w.Code != http.StatusOK {
				t.Errorf("api test bundle handler failed: %d", w.Code)
			}
		}(clientIP)

		go func(ip string) {
			defer wg.Done()
			req := httptest.NewRequest("GET", "/install.sh", nil)
			req.RemoteAddr = ip
			w := httptest.NewRecorder()
			handler.ServeHTTP(w, req)
			if w.Code != http.StatusOK {
				t.Errorf("install.sh handler failed: %d", w.Code)
			}
		}(clientIP)
	}
	wg.Wait()
}
