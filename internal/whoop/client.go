package whoop

import (
	"fmt"
	"net/http"
	"strings"
)

const DefaultBaseURL = "https://api.prod.whoop.com/developer"

type Client struct {
	BaseURL    string
	HTTPClient *http.Client
}

type APIError struct {
	StatusCode int
	Status     string
	Message    string
}

func (e *APIError) Error() string {
	if e.Message == "" {
		return fmt.Sprintf("whoop api error: %s", e.Status)
	}
	return fmt.Sprintf("whoop api error: %s - %s", e.Status, e.Message)
}

func (c *Client) endpoint(path string) string {
	baseURL := c.BaseURL
	if baseURL == "" {
		baseURL = DefaultBaseURL
	}
	return strings.TrimRight(baseURL, "/") + path
}

func (c *Client) do(req *http.Request) (*http.Response, error) {
	client := c.HTTPClient
	if client == nil {
		client = http.DefaultClient
	}
	return client.Do(req)
}

func apiError(resp *http.Response) error {
	return &APIError{
		StatusCode: resp.StatusCode,
		Status:     resp.Status,
		Message:    statusMessage(resp.StatusCode),
	}
}

func statusMessage(statusCode int) string {
	switch statusCode {
	case http.StatusBadRequest:
		return "client error constructing request"
	case http.StatusUnauthorized:
		return "invalid authorization"
	case http.StatusNotFound:
		return "resource not found"
	case http.StatusTooManyRequests:
		return "request rejected due to rate limiting"
	case http.StatusInternalServerError:
		return "server error occurred while making request"
	default:
		return ""
	}
}
