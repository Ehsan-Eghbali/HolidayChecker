package gate

import (
	"context"
	"sync"
)

func CheckSafe(ctx context.Context, date string, countries []string, hc HolidayClient) Result {
	var wg sync.WaitGroup
	hits := make(chan string, len(countries))
	errs := make(chan struct {
		code string
		err  error
	}, len(countries))

	for _, c := range countries {
		wg.Add(1)
		go func(code string) {
			defer wg.Done()
			ok, err := hc.IsPublic(ctx, code, date)
			if err != nil {
				errs <- struct {
					code string
					err  error
				}{code, err}
			}
			if ok {
				hits <- code
			}
		}(c)
	}

	wg.Wait()
	close(hits)
	close(errs)

	res := Result{Safe: true, Errors: map[string]error{}}
	for h := range hits {
		res.Safe = false
		res.HitCountries = append(res.HitCountries, h)
	}
	for e := range errs {
		res.Safe = false
		res.Errors[e.code] = e.err
	}
	return res
}
