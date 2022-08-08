package main

import (
	"context"
	"fmt"
	"io"
	"net/http"
	_ "net/http/pprof"
	"time"

	"github.com/tcnksm/go-httpstat"
)

func doRequest(ctx context.Context, url string) (io.ReadCloser, error) {
	req, err := http.NewRequest(http.MethodGet, url, nil)
	if err != nil {
		return nil, err
	}
	var result httpstat.Result

	req = req.WithContext(ctx)
	req = req.WithContext(httpstat.WithHTTPStat(req.Context(), &result))

	client := &http.Client{
		Transport: &http.Transport{
			ResponseHeaderTimeout: 10 * time.Second,
		},
	}
	resp, err := client.Do(req)
	// TODO show network info
	if err != nil {
		return nil, err
	}
	if resp.StatusCode != 200 {
		return nil, fmt.Errorf("resp code is %d", resp.StatusCode)
	}
	return resp.Body, nil
}
