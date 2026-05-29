package simple

import (
	"context"
	"io"
	"log/slog"
	"net"
	"net/http"
	"regexp"
	"time"
)

type SimpleResolver struct {
	url string
}

func New(url string) *SimpleResolver {
	return &SimpleResolver{
		url: url,
	}
}

func (r *SimpleResolver) Resolve(ctx context.Context) (string, error) {
	client := http.Client{
		Transport: &http.Transport{
			Proxy: nil,
			DialContext: (&net.Dialer{
				Timeout:   5 * time.Second,
				KeepAlive: 30 * time.Second,
				Resolver: &net.Resolver{
					PreferGo: false,
				},
			}).DialContext,
		},
	}

	var req *http.Request
	var err error

	if ctx == nil {
		req, err = http.NewRequest("GET", r.url, nil)
		if err != nil {
			return "", err
		}
	} else {
		req, err = http.NewRequestWithContext(ctx, "GET", r.url, nil)
		if err != nil {
			return "", err
		}
	}

	req.Header.Set("User-Agent", "curl/8.4.0")

	resp, err := client.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}

	slog.Debug("Request", "to", r.url, "response", string(body))

	ipRegex := regexp.MustCompile(`\b(?:[0-9]{1,3}\.){3}[0-9]{1,3}\b`)

	ipAddress := ipRegex.FindString(string(body))
	return ipAddress, nil
}
