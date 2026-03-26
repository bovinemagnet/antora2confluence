package confluence

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"strconv"
	"time"
)

type Client struct {
	baseURL    string
	username   string
	token      string
	httpClient *http.Client
	maxRetries int
}

func NewClient(baseURL, username, token string) *Client {
	return &Client{
		baseURL:    baseURL,
		username:   username,
		token:      token,
		httpClient: &http.Client{Timeout: 30 * time.Second},
		maxRetries: 3,
	}
}

func (c *Client) ValidateAuth(spaceKey string) (string, error) {
	url := fmt.Sprintf("%s/api/v2/spaces?key=%s", c.baseURL, spaceKey)
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return "", err
	}

	resp, err := c.do(req)
	if err != nil {
		return "", fmt.Errorf("validating auth: %w", err)
	}
	defer resp.Body.Close()

	var result SpaceList
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return "", fmt.Errorf("decoding space response: %w", err)
	}

	if len(result.Results) == 0 {
		return "", fmt.Errorf("space %q not found or not accessible", spaceKey)
	}

	return result.Results[0].ID, nil
}

func (c *Client) do(req *http.Request) (*http.Response, error) {
	req.SetBasicAuth(c.username, c.token)
	if req.Header.Get("Content-Type") == "" {
		req.Header.Set("Content-Type", "application/json")
	}
	req.Header.Set("Accept", "application/json")

	// Buffer the request body so it can be replayed on retries.
	var bodyBytes []byte
	if req.Body != nil {
		var err error
		bodyBytes, err = io.ReadAll(req.Body)
		if err != nil {
			return nil, fmt.Errorf("reading request body: %w", err)
		}
		req.Body.Close()
	}

	var lastErr error
	for attempt := 0; attempt <= c.maxRetries; attempt++ {
		if attempt > 0 {
			slog.Debug("Retrying request", "attempt", attempt, "url", req.URL.String())
		}

		// Reset body for each attempt
		if bodyBytes != nil {
			req.Body = io.NopCloser(bytes.NewReader(bodyBytes))
			req.ContentLength = int64(len(bodyBytes))
		}

		resp, err := c.httpClient.Do(req)
		if err != nil {
			lastErr = err
			backoff(attempt)
			continue
		}

		if resp.StatusCode == http.StatusTooManyRequests {
			resp.Body.Close()
			retryAfter := parseRetryAfter(resp.Header.Get("Retry-After"))
			if retryAfter > 0 {
				time.Sleep(retryAfter)
			} else {
				backoff(attempt)
			}
			lastErr = fmt.Errorf("rate limited (429) on %s", req.URL.Path)
			continue
		}

		if resp.StatusCode >= 500 {
			resp.Body.Close()
			lastErr = fmt.Errorf("server error (%d) on %s", resp.StatusCode, req.URL.Path)
			backoff(attempt)
			continue
		}

		return resp, nil
	}

	return nil, fmt.Errorf("request failed after %d retries: %w", c.maxRetries, lastErr)
}

func (c *Client) decodeResponse(resp *http.Response, target interface{}) error {
	defer resp.Body.Close()

	if resp.StatusCode >= 400 {
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("API error %d: %s", resp.StatusCode, string(body))
	}

	return json.NewDecoder(resp.Body).Decode(target)
}

func backoff(attempt int) {
	d := time.Duration(1<<uint(attempt)) * 100 * time.Millisecond
	if d > 5*time.Second {
		d = 5 * time.Second
	}
	time.Sleep(d)
}

func parseRetryAfter(val string) time.Duration {
	if val == "" {
		return 0
	}
	secs, err := strconv.Atoi(val)
	if err != nil {
		return 0
	}
	return time.Duration(secs) * time.Second
}
