package client

import (
	"encoding/json"
	"log/slog"
	"net/http"

	"github.com/eevandeya/lar/internal/api"
)

func (c *Client) Wake(host string) error {
	req, err := http.NewRequest("POST", c.url("/wake"), nil)
	if err != nil {
		return err
	}

	q := req.URL.Query()
	q.Add("host", host)
	req.URL.RawQuery = q.Encode()
	c.setAuthHeader(req)

	slog.Debug("sending request to the gateway", "method", req.Method, "url", req.URL)
	resp, err := c.client.Do(req)
	if err != nil {
		return err
	}

	if resp.StatusCode >= 400 {
		var errorResponse api.ErrorResponse
		if err = json.NewDecoder(resp.Body).Decode(&errorResponse); err != nil {
			return err
		}
		return &HostError{
			HostName: host,
			Err: APIError{
				StatusCode: resp.StatusCode,
				Code:       errorResponse.Error.Code,
				Message:    errorResponse.Error.Message,
			},
		}
	}

	var response api.StatusResponse
	if err = json.NewDecoder(resp.Body).Decode(&response); err != nil {
		return err
	}

	return nil
}
