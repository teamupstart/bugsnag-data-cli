package bugsnag

import (
	"context"
	"encoding/json"
	"net/http"
)

// Me struct holds response from /user endpoint.
type Me struct {
	Name  string `json:"displayName"`
	Login string `json:"email"`
}

// Me fetches response from /user endpoint.
func (c *Client) Me() (*Me, error) {
	res, err := c.Get(context.Background(), "/user", nil)
	if err != nil {
		return nil, err
	}
	if res != nil {
		defer func() { _ = res.Body.Close() }()
	}
	if res.StatusCode != http.StatusOK {
		return nil, formatUnexpectedResponse(res)
	}

	var me Me

	err = json.NewDecoder(res.Body).Decode(&me)

	return &me, err
}
