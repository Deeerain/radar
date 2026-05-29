package simple

import (
	"context"
	"io"
	"log/slog"
	"net/http"
	"regexp"
)

type SimpleResolver struct {
	url        string
	httpClient *http.Client
}

func New(url string, client http.Client) *SimpleResolver {
	return &SimpleResolver{
		url:        url,
		httpClient: &client,
	}
}

func (r *SimpleResolver) Resolve(ctx context.Context) (string, error) {
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

	resp, err := r.httpClient.Do(req)
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
