package bugsnag

import (
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestMe(t *testing.T) {
	var unexpectedStatusCode bool

	api_endpoint := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "/user", r.URL.Path)

		if unexpectedStatusCode {
			w.WriteHeader(400)
		} else {
			resp, err := ioutil.ReadFile("./testdata/user_me.json")
			assert.NoError(t, err)

			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(200)
			_, _ = w.Write(resp)
		}
	}))
	defer api_endpoint.Close()

	client := NewClient(Config{APIEndpoint: api_endpoint.URL}, WithTimeout(3*time.Second))

	actual, err := client.Me()
	assert.NoError(t, err)

	expected := &Me{
		Name:  "User A",
		Login: "user@bugsnag.com",
	}
	assert.Equal(t, expected, actual)

	unexpectedStatusCode = true

	_, err = client.Me()
	assert.Error(t, &ErrUnexpectedResponse{}, err)
}
