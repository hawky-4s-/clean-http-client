package http

import (
	"context"
	"fmt"
	"io"
	"log"
	"net"
	"net/http"
	"strings"
	"time"
)

const (
	jsonType              = "application/json"
	defaultRequestTimeOut = 30 * time.Second
)

// Client provides a high-level API for working with HTTP requests and constructing them.
type Client interface {
	GetFrom(path string) (*http.Response, error)
	PostTo(path string, body io.Reader) (*http.Response, error)
	PutTo(path string, body io.Reader) (*http.Response, error)
	DeleteFrom(path string) (*http.Response, error)

	GetRequest(path string) (*http.Request, error)
	PostRequest(path string, body io.Reader) (*http.Request, error)
	PutRequest(path string, body io.Reader) (*http.Request, error)
	DeleteRequest(path string) (*http.Request, error)
	ExecuteRequest(r *http.Request) (*http.Response, error)

	GetFromWithContext(ctx context.Context, path string) (*http.Response, error)
	PostToWithContext(ctx context.Context, path string, body io.Reader) (*http.Response, error)
	PutToWithContext(ctx context.Context, path string, body io.Reader) (*http.Response, error)
	DeleteFromWithContext(ctx context.Context, path string) (*http.Response, error)
}

// HttpConfig holds the base configuration for the HttpClient.
type HttpConfig struct {
	baseURL  string
	username string
	password string
	accept   string
}

// HttpClient wraps the underlying http.Client and its HttpConfig.
type HttpClient struct {
	client *http.Client
	config *HttpConfig
}

// NotFoundError allows to check for the not found url
type NotFoundError struct {
	Message string
	URL     string
}

func (e NotFoundError) Error() string {
	return e.Message
}

// UnauthorizedError contains the rejected URL and Status
type UnauthorizedError struct {
	Message string
	URL     string
	Status  int
}

func (e UnauthorizedError) Error() string {
	return e.Message
}

// RemoteError
type RemoteError struct {
	Host string
	err  error
}

func (e RemoteError) Error() string {
	return e.err.Error()
}

func NewHttpConfig(baseURL string, username string, password string, accept string) *HttpConfig {
	config := &HttpConfig{
		baseURL:  baseURL,
		username: username,
		password: password,
		accept:   jsonType,
	}

	if accept != "" {
		config.accept = accept
	}

	return config
}

func NewDefaultHttpConfig(baseURL string) *HttpConfig {
	return NewHttpConfig(baseURL, "", "", jsonType)
}

// Create a new default HttpClient with a custom transport for clean resource usage
func NewDefaultHttpClient(baseURL string) *HttpClient {
	client := &http.Client{
		Transport: &http.Transport{
			Proxy: http.ProxyFromEnvironment,
			DialContext: (&net.Dialer{
				Timeout:   defaultRequestTimeOut,
				KeepAlive: defaultRequestTimeOut,
			}).DialContext,
			MaxIdleConns:          100,
			IdleConnTimeout:       90 * time.Second,
			TLSHandshakeTimeout:   10 * time.Second,
			ExpectContinueTimeout: 1 * time.Second,
		},
		Timeout: defaultRequestTimeOut,
	}

	config := NewDefaultHttpConfig(baseURL)

	return &HttpClient{
		client: client,
		config: config,
	}
}

// NewHttpClientWithConfig creates a new HttpClient with given HttpConfig and a custom transport for clean resource usage
func NewHttpClientWithConfig(config *HttpConfig) *HttpClient {
	if config == nil {
		panic("config is nil")
	}

	client := &http.Client{
		Transport: &http.Transport{
			Proxy: http.ProxyFromEnvironment,
			DialContext: (&net.Dialer{
				Timeout:   defaultRequestTimeOut,
				KeepAlive: defaultRequestTimeOut,
			}).DialContext,
			MaxIdleConns:          100,
			IdleConnTimeout:       90 * time.Second,
			TLSHandshakeTimeout:   10 * time.Second,
			ExpectContinueTimeout: 1 * time.Second,
		},
		Timeout: defaultRequestTimeOut,
	}

	return &HttpClient{
		client: client,
		config: config,
	}
}

// NewHttpClientWithConfigAndClient creates a new HttpClient with given HttpConfig and a custom http.Client.
func NewHttpClientWithConfigAndClient(config *HttpConfig, client *http.Client) *HttpClient {
	if config == nil {
		panic("config is nil")
	}
	if client == nil {
		panic("client is nil")
	}

	return &HttpClient{
		client: client,
		config: config,
	}
}

//
// Interface implementations
//
func (h *HttpClient) GetFrom(path string) (*http.Response, error) {
	return h.GetFromWithContext(context.Background(), path)
}

func (h *HttpClient) PostTo(path string, body io.Reader) (*http.Response, error) {
	return h.PostToWithContext(context.Background(), path, body)
}

func (h *HttpClient) PutTo(path string, body io.Reader) (*http.Response, error) {
	return h.PutToWithContext(context.Background(), path, body)
}

func (h *HttpClient) DeleteFrom(path string) (*http.Response, error) {
	return h.DeleteFromWithContext(context.Background(), path)
}

func (h *HttpClient) GetFromWithContext(ctx context.Context, path string) (*http.Response, error) {
	request, err := createRequest(ctx, h.config.baseURL, path, http.MethodGet, nil, h.config.username, h.config.password)
	if err != nil {
		return nil, err
	}
	requestWithCtx := request.WithContext(ctx)
	return h.ExecuteRequest(requestWithCtx)
}

func (h *HttpClient) PostToWithContext(ctx context.Context, path string, body io.Reader) (*http.Response, error) {
	request, err := createRequest(ctx, h.config.baseURL, path, http.MethodPost, body, h.config.username, h.config.password)
	if err != nil {
		return nil, err
	}
	requestWithCtx := request.WithContext(ctx)
	return h.ExecuteRequest(requestWithCtx)
}

func (h *HttpClient) PutToWithContext(ctx context.Context, path string, body io.Reader) (*http.Response, error) {
	request, err := createRequest(ctx, h.config.baseURL, path, http.MethodPut, body, h.config.username, h.config.password)
	if err != nil {
		return nil, err
	}
	requestWithCtx := request.WithContext(ctx)
	return h.ExecuteRequest(requestWithCtx)
}

func (h *HttpClient) DeleteFromWithContext(ctx context.Context, path string) (*http.Response, error) {
	request, err := createRequest(ctx, h.config.baseURL, path, http.MethodDelete, nil, h.config.username, h.config.password)
	if err != nil {
		return nil, err
	}
	requestWithCtx := request.WithContext(ctx)
	return h.ExecuteRequest(requestWithCtx)
}

func (h *HttpClient) GetRequest(path string) (*http.Request, error) {
	return createRequest(nil, h.config.baseURL, path, http.MethodGet, nil, h.config.username, h.config.password)
}

func (h *HttpClient) PostRequest(path string, body io.Reader) (*http.Request, error) {
	return createRequest(nil, h.config.baseURL, path, http.MethodPost, body, h.config.username, h.config.password)
}

func (h *HttpClient) PutRequest(path string, body io.Reader) (*http.Request, error) {
	return createRequest(nil, h.config.baseURL, path, http.MethodPut, body, h.config.username, h.config.password)
}

func (h *HttpClient) DeleteRequest(path string) (*http.Request, error) {
	return createRequest(nil, h.config.baseURL, path, http.MethodDelete, nil, h.config.username, h.config.password)
}

//
// Internal functions
//
func createDefaultContext(ctx context.Context) (context.Context, context.CancelFunc) {
	if ctx != nil {
		return context.WithTimeout(ctx, defaultRequestTimeOut)
	}
	return context.WithTimeout(context.Background(), defaultRequestTimeOut)
}

func createRequest(ctx context.Context, baseURL string, endpoint string, method string, body io.Reader, username string, password string) (*http.Request, error) {
	// construct url by appending endpoint to base url
	baseURL = strings.TrimSuffix(baseURL, "/")
	endpoint = strings.TrimPrefix(endpoint, "/")

	request, err := http.NewRequest(method, baseURL+"/"+endpoint, body)
	if err != nil {
		return request, err
	}

	request.Header.Set("Content-Type", jsonType)
	request.Header.Set("Accept", jsonType)

	if username != "" && password != "" {
		request.SetBasicAuth(username, password)
	}

	return request, nil
}

func (h *HttpClient) ExecuteRequest(r *http.Request) (*http.Response, error) {
	if r.Context() == context.Background() {
		// TODO: handle Context's cancel function
		ctx, _ := createDefaultContext(r.Context())
		r = r.WithContext(ctx)
	}

	resp, err := h.client.Do(r)

	if err != nil {
		return handleError(resp, err)
	}

	return resp, nil
}

func handleError(resp *http.Response, error error) (*http.Response, error) {
	log.Fatal(error)

	if resp.StatusCode == http.StatusUnauthorized {
		return resp, &UnauthorizedError{Message: "Authentication required.", URL: resp.Request.URL.String()}
	}

	if resp.StatusCode == http.StatusNotFound {
		return resp, &NotFoundError{Message: "Resource not found.", URL: resp.Request.URL.String()}
	}

	return resp, &RemoteError{resp.Request.URL.Host, fmt.Errorf("%d: (%s)", resp.StatusCode, resp.Request.URL.String())}
}

type RequestBuilder interface {
	Get() RequestBuilder
	Post() RequestBuilder
	Put() RequestBuilder
	Delete() RequestBuilder
	Path(path string) RequestBuilder
	QueryParam(key string, value string) RequestBuilder
	WithContent(body io.Reader) RequestBuilder
	AsJson() RequestBuilder
	Build() (*http.Request, error)
}

type requestBuilder struct {
	method      string
	path        string
	queryParams map[string]interface{}
	body        io.Reader
	request     *http.Request
	accept      string
}

func NewRequestBuilder() RequestBuilder {
	return &requestBuilder{}
}

func (rb *requestBuilder) Get() RequestBuilder {
	rb.method = http.MethodGet
	return rb
}

func (rb *requestBuilder) Post() RequestBuilder {
	rb.method = http.MethodPost
	return rb
}

func (rb *requestBuilder) Put() RequestBuilder {
	rb.method = http.MethodPut
	return rb
}

func (rb *requestBuilder) Delete() RequestBuilder {
	rb.method = http.MethodDelete
	return rb
}

func (rb *requestBuilder) Path(path string) RequestBuilder {
	rb.path = path
	return rb
}

func (rb *requestBuilder) WithContent(body io.Reader) RequestBuilder {
	rb.body = body
	return rb
}

func (rb *requestBuilder) AsJson() RequestBuilder {
	rb.accept = jsonType
	return rb
}

func (rb *requestBuilder) QueryParam(key string, value string) RequestBuilder {
	if rb.queryParams == nil {
		rb.queryParams = make(map[string]interface{})
	}
	rb.queryParams[key] = value
	return rb
}

func (rb *requestBuilder) Build() (*http.Request, error) {
	request, err := http.NewRequest(rb.method, rb.path, rb.body)

	if rb.queryParams != nil {
		queryValues := request.URL.Query()

		for key, value := range rb.queryParams {
			queryValues.Add(key, value.(string))
		}

		request.URL.RawQuery = queryValues.Encode()
	}

	if err != nil {
		return request, err
	}
	return request, nil
}
