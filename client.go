package http

import (
	"errors"
	"fmt"
	"io"
	"log"
	"net"
	"net/http"
	"strings"
	"time"
)

const (
	JSON_TYPE = "application/json"
)

type Client interface {
	GetFrom(path string) (*http.Response, error)
	PostTo(path string, body io.Reader) (*http.Response, error)
	PutTo(path string, body io.Reader) (*http.Response, error)
	DeleteFrom(path string) (*http.Response, error)

	GetRequest(path string) (*http.Request, error)
	PostRequest(path string, body io.Reader) (*http.Request, error)
	PutRequest(path string, body io.Reader) (*http.Request, error)
	DeleteRequest(path string) (*http.Request, error)
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
	Request() (*http.Request, error)
	Exec() (*http.Response, error)
}

type requestBuilder struct {
	method      string
	path        string
	queryParams []string
	body        io.Reader
	request     *http.Request
	accept      string
}

type HttpConfig struct {
	baseUrl  string
	username string
	password string
	accept   string
}

type HttpClient struct {
	requestBuilder
	client *http.Client
	config *HttpConfig
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
	rb.accept = JSON_TYPE
	return rb
}

func (rb *requestBuilder) QueryParam(key string, value string) RequestBuilder {
	rb.queryParams = append(rb.queryParams, key, "=", value)
	return rb
}

func (rb requestBuilder) Request() (*http.Request, error) {
	request, error := http.NewRequest(rb.method, rb.path, rb.body)

	if error != nil {
		return request, error
	}
	return request, nil
}

// TODO: implement correctly
func (rb requestBuilder) Exec() (*http.Response, error) {
	if rb.request == nil {
		return nil, errors.New("you must create a request using the request builder before calling Exec()")
	}

	return nil, nil
}

type NotFoundError struct {
	Message string
	Url     string
}

func (e NotFoundError) Error() string {
	return e.Message
}

type UnauthorizedError struct {
	Message string
	Url     string
}

func (e UnauthorizedError) Error() string {
	return e.Message
}

type RemoteError struct {
	Host string
	err  error
}

func (e RemoteError) Error() string {
	return e.err.Error()
}

func NewHttpConfig(baseUrl string, username string, password string, accept string) *HttpConfig {
	config := &HttpConfig{
		baseUrl:  baseUrl,
		username: username,
		password: password,
		accept:   JSON_TYPE,
	}

	if accept != "" {
		config.accept = accept
	}

	return config
}

func DefaultHttpConfig(baseUrl string) *HttpConfig {
	return NewHttpConfig(baseUrl, "", "", JSON_TYPE)
}

/**
 * Create a new HTTPClient with a custom transport for clean resource usage
 */
func NewHttpClient(config *HttpConfig) *HttpClient {
	client := &http.Client{
		Transport: &http.Transport{
			Proxy: http.ProxyFromEnvironment,
			DialContext: (&net.Dialer{
				Timeout:   30 * time.Second,
				KeepAlive: 30 * time.Second,
			}).DialContext,
			MaxIdleConns:          100,
			IdleConnTimeout:       90 * time.Second,
			TLSHandshakeTimeout:   10 * time.Second,
			ExpectContinueTimeout: 1 * time.Second,
		},
		Timeout: 30 * time.Second,
	}

	return &HttpClient{
		client: client,
		config: config,
	}
}

func NewDefaultHttpClient(baseUrl string) *HttpClient {
	client := &http.Client{
		Transport: &http.Transport{
			Proxy: http.ProxyFromEnvironment,
			DialContext: (&net.Dialer{
				Timeout:   30 * time.Second,
				KeepAlive: 30 * time.Second,
			}).DialContext,
			MaxIdleConns:          100,
			IdleConnTimeout:       90 * time.Second,
			TLSHandshakeTimeout:   10 * time.Second,
			ExpectContinueTimeout: 1 * time.Second,
		},
		Timeout: 30 * time.Second,
	}

	config := DefaultHttpConfig(baseUrl)

	return &HttpClient{
		client: client,
		config: config,
	}
}

func (h *HttpClient) GetFrom(path string) (*http.Response, error) {
	request, error := createRequest(h.config.baseUrl, path, http.MethodGet, nil, h.config.username, h.config.password)
	if error != nil {
		return nil, error
	}
	return h.executeRequest(request)
}

func (h *HttpClient) PostTo(path string, body io.Reader) (*http.Response, error) {
	request, error := createRequest(h.config.baseUrl, path, http.MethodPost, body, h.config.username, h.config.password)
	if error != nil {
		return nil, error
	}
	return h.executeRequest(request)
}

func (h *HttpClient) PutTo(path string, body io.Reader) (*http.Response, error) {
	request, error := createRequest(h.config.baseUrl, path, http.MethodPut, body, h.config.username, h.config.password)
	if error != nil {
		return nil, error
	}
	return h.executeRequest(request)
}

func (h *HttpClient) DeleteFrom(path string) (*http.Response, error) {
	request, error := createRequest(h.config.baseUrl, path, http.MethodDelete, nil, h.config.username, h.config.password)
	if error != nil {
		return nil, error
	}
	return h.executeRequest(request)
}

func (h *HttpClient) GetRequest(path string) (*http.Request, error) {
	return createRequest(h.config.baseUrl, path, http.MethodGet, nil, h.config.username, h.config.password)
}

func (h *HttpClient) PostRequest(path string, body io.Reader) (*http.Request, error) {
	return createRequest(h.config.baseUrl, path, http.MethodPost, body, h.config.username, h.config.password)
}

func (h *HttpClient) PutRequest(path string, body io.Reader) (*http.Request, error) {
	return createRequest(h.config.baseUrl, path, http.MethodPut, body, h.config.username, h.config.password)
}

func (h *HttpClient) DeleteRequest(path string) (*http.Request, error) {
	return createRequest(h.config.baseUrl, path, http.MethodDelete, nil, h.config.username, h.config.password)
}

func createRequest(baseUrl string, endpoint string, method string, body io.Reader, username string, password string) (*http.Request, error) {
	// construct url by appending endpoint to base url
	baseUrl = strings.TrimSuffix(baseUrl, "/")

	request, err := http.NewRequest(method, baseUrl+"/"+endpoint, body)
	if err != nil {
		return request, err
	}

	request.Header.Set("Content-Type", JSON_TYPE)
	request.Header.Set("Accept", JSON_TYPE)

	if username != "" && password != "" {
		request.SetBasicAuth(username, password)
	}

	return request, nil
}

func (h *HttpClient) executeRequest(r *http.Request) (*http.Response, error) {
	resp, error := h.client.Do(r)

	if error != nil {
		return handleError(resp, error)
	}

	return resp, nil
}

func handleError(resp *http.Response, error error) (*http.Response, error) {
	log.Fatal(error)

	if resp.StatusCode == http.StatusUnauthorized {
		return resp, &UnauthorizedError{Message: "Authentication required.", Url: resp.Request.URL.String()}
	}

	if resp.StatusCode == http.StatusNotFound {
		return resp, &NotFoundError{Message: "Resource not found.", Url: resp.Request.URL.String()}
	}

	return resp, &RemoteError{resp.Request.URL.Host, fmt.Errorf("%d: (%s)", resp.StatusCode, resp.Request.URL.String())}
}
