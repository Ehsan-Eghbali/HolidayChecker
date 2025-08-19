package gate

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"
)

type HolidayClient struct {
	BaseURL    string       // e.g. https://openholidaysapi.org/PublicHolidays
	HTTP       *http.Client // injectable for tests
	Timeout    time.Duration
	MaxRetries int
}

type holiday struct {
	Type string `json:"type"` // "Public"
}

func (hc HolidayClient) IsPublic(ctx context.Context, countryISO, date string) (bool, error) {
	client := hc.HTTP
	if client == nil {
		client = http.DefaultClient
	}

	var last error
	for attempt := 0; attempt <= hc.MaxRetries; attempt++ {
		ok, err := func() (bool, error) {
			c, cancel := context.WithTimeout(ctx, hc.Timeout)
			defer cancel()
			url := fmt.Sprintf("%s?countryIsoCode=%s&languageIsoCode=EN&validFrom=%s&validTo=%s",
				hc.BaseURL, countryISO, date, date)
			req, reqErr := http.NewRequestWithContext(c, http.MethodGet, url, nil)
			if reqErr != nil {
				return false, fmt.Errorf("build request: %w", reqErr)
			}
			resp, doErr := client.Do(req)
			if doErr != nil {
				return false, doErr
			}
			defer resp.Body.Close()
			// retryable server errors
			if resp.StatusCode >= 500 {
				return false, fmt.Errorf("server %d", resp.StatusCode)
			}

			// non-retryable client errors: return immediately (but we still closed body via defer)
			if resp.StatusCode != http.StatusOK {
				b, _ := io.ReadAll(resp.Body)
				return false, fmt.Errorf("status %d: %s", resp.StatusCode, string(b))
			}

			var arr []holiday
			if err := json.NewDecoder(resp.Body).Decode(&arr); err != nil {
				return false, fmt.Errorf("decode: %w", err)
			}
			for _, h := range arr {
				if h.Type == "Public" {
					return true, nil
				}
			}
			return false, nil
		}()

		if err == nil {
			return ok, nil
		}
		last = err

		if attempt < hc.MaxRetries {
			time.Sleep(backoff(attempt))
		}
	}
	return false, last
}
func backoff(attempt int) time.Duration {
	base := 150 * time.Millisecond
	duration := 2 * time.Second
	d := base << attempt
	if d > duration {
		d = duration
	}
	return d
}
