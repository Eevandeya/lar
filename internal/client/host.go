package client

import (
	"encoding/json"
	"fmt"
	"log/slog"
	"net/http"

	"github.com/eevandeya/lar/internal/api"
)

func (c *Client) requestHost(method, target, host string) (*http.Response, error) {
	req, err := http.NewRequest(method, c.url(target), nil)
	if err != nil {
		return nil, err
	}

	q := req.URL.Query()
	q.Add("host", host)
	req.URL.RawQuery = q.Encode()
	c.setAuthHeader(req)

	slog.Debug("sending request to the gateway", "method", req.Method, "url", req.URL)
	resp, err := c.client.Do(req)
	if err != nil {
		return nil, err
	}

	if resp.StatusCode >= 400 {
		if resp.Header.Get("Content-Type") == "Application/json" {
			var errorResponse api.ErrorResponse
			if err = json.NewDecoder(resp.Body).Decode(&errorResponse); err != nil {
				return nil, fmt.Errorf("failed to parse error response body with code %d: %w", resp.StatusCode, err)
			}
			return nil, &HostError{
				HostName: host,
				Err: APIError{
					StatusCode: resp.StatusCode,
					Code:       errorResponse.Error.Code,
					Message:    errorResponse.Error.Message,
				},
			}
		}
		err = resp.Body.Close()
		if err != nil {
			slog.Debug("failed to close response body", "err", err)
		}
		return nil, UnexpectedAPIError(resp.StatusCode)
	}

	return resp, nil
}

func (c *Client) Shutdown(host string) error {
	resp, err := c.requestHost("POST", "/shutdown", host)
	if err != nil {
		return err
	}
	err = resp.Body.Close()
	if err != nil {
		slog.Debug("failed to close response body", "err", err)
		return err
	}
	return nil
}

func (c *Client) Status(host string) (bool, error) {
	resp, err := c.requestHost("GET", "/status", host)
	if err != nil {
		return false, err
	}

	var response api.StatusResponse
	if err = json.NewDecoder(resp.Body).Decode(&response); err != nil {
		return false, err
	}

	return response.Online, nil
}

func (c *Client) Wake(host string) error {
	resp, err := c.requestHost("POST", "/wake", host)
	if err != nil {
		return err
	}

	err = resp.Body.Close()
	if err != nil {
		slog.Debug("failed to close response body", "err", err)
		return err
	}
	return nil
}
