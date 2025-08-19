package main

import (
	"context"
	"flag"
	"fmt"
	"github.com/Ehsan-Eghbali/HolidayChecker/internal/gate"
	"log"
	"strings"
	"time"
)

func main() {
	dateArg := flag.String("date", "", "YYYY/MM/DD")
	countriesArg := flag.String("countries", "", "comma-separated e.g. ES,FR,IT")
	base := flag.String("base", "https://openholidaysapi.org/PublicHolidays", "holidays API base")
	timeout := flag.Duration("timeout", 1500*time.Millisecond, "per-request timeout")
	retries := flag.Int("retries", 2, "retries per country")
	flag.Parse()

	if *dateArg == "" || *countriesArg == "" {
		log.Fatal("usage: holidayChecker -date=YYYY/MM/DD -countries=ES,FR,IT")
	}

	t, err := time.Parse("2006/01/02", *dateArg)
	if err != nil {
		log.Fatalf("invalid date: %v", err)
	}
	date := t.Format("2006-01-02")

	var codes []string
	for _, p := range strings.Split(*countriesArg, ",") {
		p = strings.ToUpper(strings.TrimSpace(p))
		if p != "" {
			codes = append(codes, p)
		}
	}

	hc := gate.HolidayClient{
		BaseURL:    *base,
		Timeout:    *timeout,
		MaxRetries: *retries,
	}
	res := gate.CheckSafe(context.Background(), date, codes, hc)

	if res.Safe {
		fmt.Printf("SAFE: %s (no holidays in %v)\n", date, codes)
		return
	}
	fmt.Printf("UNSAFE: %s\n", date)
	if len(res.HitCountries) > 0 {
		fmt.Printf("Holidays in: %v\n", res.HitCountries)
	}
	if len(res.Errors) > 0 {
		fmt.Printf("Errors: %v\n", res.Errors)
	}
}
