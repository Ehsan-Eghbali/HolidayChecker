package gate

import (
	"context"
	"encoding/json"
	"fmt"
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
		c, cancel := context.WithTimeout(ctx, hc.Timeout)
		url := fmt.Sprintf("%s?countryIsoCode=%s&languageIsoCode=EN&validFrom=%s&validTo=%s",
			hc.BaseURL, countryISO, date, date)
		req, _ := http.NewRequestWithContext(c, http.MethodGet, url, nil)

		resp, err := client.Do(req)
		cancel()
		if err != nil {
			last = err
			time.Sleep(backoff(attempt))
			continue
		}
		defer resp.Body.Close()

		if resp.StatusCode >= 500 {
			last = fmt.Errorf("server %d", resp.StatusCode)
			time.Sleep(backoff(attempt))
			continue
		}
		if resp.StatusCode != http.StatusOK {
			return false, fmt.Errorf("status %d", resp.StatusCode)
		}

		var arr []holiday
		if err := json.NewDecoder(resp.Body).Decode(&arr); err != nil {
			return false, err
		}
		for _, h := range arr {
			if h.Type == "Public" {
				return true, nil
			}
		}
		return false, nil
	}
	return false, last
}

func backoff(attempt int) time.Duration {
	base := 150 * time.Millisecond
	max := 2 * time.Second
	d := base << attempt
	if d > max {
		d = max
	}
	return d
}
