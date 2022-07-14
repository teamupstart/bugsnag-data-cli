package bugsnag

import (
	"context"
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestGet(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "/user/organizations", r.URL.Path)
		assert.Equal(t, url.Values{
			"per_page": []string{"1"},
		}, r.URL.Query())
		assert.Equal(t, "application/json", r.Header.Get("Content-Type"))

		w.WriteHeader(200)
	}))
	defer server.Close()

	client := NewClient(Config{APIEndpoint: server.URL}, WithTimeout(3*time.Second))
	resp, err := client.Get(context.Background(), "/user/organizations?per_page=1", Header{
		"Content-Type": "application/json",
	})

	assert.NoError(t, err)
	assert.Equal(t, 200, resp.StatusCode)

	_ = resp.Body.Close()
}