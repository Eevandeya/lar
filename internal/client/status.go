package client

import (
	"encoding/json"
	"net/http"

	"github.com/eevandeya/lar/internal/api"
)

func (c *Client) Status(host string) (bool, error) {
	req, err := http.NewRequest("GET", c.url("/status"), nil)
	if err != nil {
		return false, err
	}

	q := req.URL.Query()
	q.Add("host", host)
	req.URL.RawQuery = q.Encode()
	c.setAuthHeader(req)

	resp, err := c.client.Do(req)
	if err != nil {
		return false, err
	}

	var response api.StatusResponse
	if err = json.NewDecoder(resp.Body).Decode(&response); err != nil {
		return false, err
	}

	return response.Online, nil
}
