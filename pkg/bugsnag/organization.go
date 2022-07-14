package bugsnag

import (
	"context"
	"encoding/json"
	"net/http"
)

func (c *Client) Organization() ([]*Organization, error) {
	res, err := c.Get(context.Background(), "/user/organizations", nil)
	if err != nil {
		return nil, err
	}
	if res == nil {
		return nil, ErrEmptyResponse
	}
	defer func() { _ = res.Body.Close() }()

	if res.StatusCode != http.StatusOK {
		return nil, formatUnexpectedResponse(res)
	}

	var out []*Organization

	err = json.NewDecoder(res.Body).Decode(&out)

	return out, err
}
