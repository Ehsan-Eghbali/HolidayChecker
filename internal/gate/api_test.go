package gate

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

func TestIsPublic_PublicTrue(t *testing.T) {
	s := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_ = json.NewEncoder(w).Encode([]map[string]any{{"type": "Public"}})
	}))
	defer s.Close()

	hc := HolidayClient{BaseURL: s.URL, Timeout: 100 * time.Millisecond, MaxRetries: 0}
	ok, err := hc.IsPublic(context.Background(), "ES", "2025-01-01")
	if err != nil || !ok {
		t.Fatalf("want public, got ok=%v err=%v", ok, err)
	}
}

func TestIsPublic_NotHoliday(t *testing.T) {
	s := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_ = json.NewEncoder(w).Encode([]any{}) // empty -> not a holiday
	}))
	defer s.Close()

	hc := HolidayClient{BaseURL: s.URL, Timeout: 100 * time.Millisecond, MaxRetries: 0}
	ok, err := hc.IsPublic(context.Background(), "ES", "2025-01-02")
	if err != nil || ok {
		t.Fatalf("want not-holiday, got ok=%v err=%v", ok, err)
	}
}

func TestCheckSafe_MixedCountries(t *testing.T) {
	pub := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_ = json.NewEncoder(w).Encode([]map[string]any{{"type": "Public"}})
	}))
	defer pub.Close()

	non := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_ = json.NewEncoder(w).Encode([]any{})
	}))
	defer non.Close()

	hc1 := HolidayClient{BaseURL: pub.URL, Timeout: 100 * time.Millisecond, MaxRetries: 1}
	hc2 := HolidayClient{BaseURL: non.URL, Timeout: 100 * time.Millisecond, MaxRetries: 1}

	r1 := CheckSafe(context.Background(), "2025-01-01", []string{"ES"}, hc1)
	r2 := CheckSafe(context.Background(), "2025-01-01", []string{"FR"}, hc2)

	if r1.Safe {
		t.Fatalf("expected unsafe for ES")
	}
	if !r2.Safe {
		t.Fatalf("expected safe for FR")
	}
}
