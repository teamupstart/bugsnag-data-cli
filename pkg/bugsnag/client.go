package bugsnag

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net"
	"net/http"
	"net/http/httputil"
	"strings"
	"time"
)

const (
	ISO8601 = "2006-01-02T15:04:05Z"
)

var (
	// ErrNoResult denotes no results.
	ErrNoResult = fmt.Errorf("bugsnag: no result")
	// ErrEmptyResponse denotes empty response from the server.
	ErrEmptyResponse = fmt.Errorf("bugsnag: empty response from server")
)

// ErrUnexpectedResponse denotes response code other than the expected one.
type ErrUnexpectedResponse struct {
	Body       Errors
	Status     string
	StatusCode int
}

func (e *ErrUnexpectedResponse) Error() string {
	return e.Body.String()
}

// ErrMultipleFailed represents a grouped error, usually when
// multiple request fails when running them in a loop.
type ErrMultipleFailed struct {
	Msg string
}

func (e *ErrMultipleFailed) Error() string {
	return e.Msg
}

// Errors is a bugsnag error type.
type Errors struct {
	Errors          map[string]string
	ErrorMessages   []string
	WarningMessages []string
}

func (e Errors) String() string {
	var out strings.Builder

	if len(e.ErrorMessages) > 0 || len(e.Errors) > 0 {
		out.WriteString("\nError:\n")
		for _, v := range e.ErrorMessages {
			out.WriteString(fmt.Sprintf("  - %s\n", v))
		}
		for k, v := range e.Errors {
			out.WriteString(fmt.Sprintf("  - %s: %s\n", k, v))
		}
	}

	if len(e.WarningMessages) > 0 {
		out.WriteString("\nWarning:\n")
		for _, v := range e.WarningMessages {
			out.WriteString(fmt.Sprintf("  - %s\n", v))
		}
	}

	return out.String()
}

// Header is a key, value pair for request headers.
type Header map[string]string

// Config is a bugsnag config.
type Config struct {
	APIEndpoint string
	Login       string
	APIToken    string
	AuthType    AuthType
	Insecure    bool
	Debug       bool
}

// Client is a bugsnag client.
type Client struct {
	transport    http.RoundTripper
	api_endpoint string
	login        string
	api_token    string
	authType     AuthType
	timeout      time.Duration
	debug        bool
}

// ClientFunc decorates option for client.
type ClientFunc func(*Client)

// NewClient instantiates new bugsnag client.
func NewClient(c Config, opts ...ClientFunc) *Client {
	client := Client{
		api_endpoint: strings.TrimSuffix(c.APIEndpoint, "/"),
		login:        c.Login,
		api_token:    c.APIToken,
		authType:     c.AuthType,
		debug:        c.Debug,
	}

	for _, opt := range opts {
		opt(&client)
	}

	client.transport = &http.Transport{
		Proxy: http.ProxyFromEnvironment,
		DialContext: (&net.Dialer{
			Timeout: client.timeout,
		}).DialContext,
	}

	return &client
}

// WithTimeout is a functional opt to attach timeout to the client.
func WithTimeout(to time.Duration) ClientFunc {
	return func(c *Client) {
		c.timeout = to
	}
}

// Get sends GET request to v3 version of the bugsnag api.
func (c *Client) Get(ctx context.Context, path string, headers Header) (*http.Response, error) {
	return c.request(ctx, http.MethodGet, c.api_endpoint+path, nil, headers)
}

func (c *Client) request(ctx context.Context, method, endpoint string, body []byte, headers Header) (*http.Response, error) {
	var (
		req *http.Request
		res *http.Response
		err error
	)

	req, err = http.NewRequest(method, endpoint, bytes.NewReader(body))
	if err != nil {
		return nil, err
	}

	defer func() {
		if c.debug {
			dump(req, res)
		}
	}()

	for k, v := range headers {
		req.Header.Set(k, v)
	}

	if c.authType == AuthTypeBasic {
		req.SetBasicAuth(c.login, c.api_token)
	} else {
		req.Header.Add("Authorization", "token "+c.api_token)
	}

	res, err = c.transport.RoundTrip(req.WithContext(ctx))

	return res, err
}

func dump(req *http.Request, res *http.Response) {
	reqDump, _ := httputil.DumpRequest(req, true)
	respDump, _ := httputil.DumpResponse(res, false)

	prettyPrintDump("Request Details", reqDump)
	prettyPrintDump("Response Details", respDump)
}

func prettyPrintDump(heading string, data []byte) {
	const separatorWidth = 60

	fmt.Printf("\n\n%s", strings.ToUpper(heading))
	fmt.Printf("\n%s\n\n", strings.Repeat("-", separatorWidth))
	fmt.Print(string(data))
}

func formatUnexpectedResponse(res *http.Response) *ErrUnexpectedResponse {
	var b Errors

	// We don't care about decoding error here.
	_ = json.NewDecoder(res.Body).Decode(&b)

	return &ErrUnexpectedResponse{
		Body:       b,
		Status:     res.Status,
		StatusCode: res.StatusCode,
	}
}
