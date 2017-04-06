package http

import (
	"io"
	"net"
	"net/http"
	"strings"
	"time"
)

const (
	JSON_TYPE = "application/json"
)

type Client interface {
	Get(path string) *http.Request
	Post(path string, body io.Reader) *http.Request
	Put(path string, body io.Reader) *http.Request
	Delete(path string) *http.Request
}

type HttpConfig struct {
	baseUrl  string
	username string
	password string
	accept   string
}

type HttpClient struct {
	client *http.Client
	config *HttpConfig
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

func (h *HttpClient) createRequest(endpoint string, method string, body io.Reader) *http.Request {
	// construct url by appending endpoint to base url
	baseUrl := strings.TrimSuffix(h.config.baseUrl, "/")

	request, err := http.NewRequest(method, baseUrl+"/"+endpoint, body)
	if err != nil {
		panic(err)
	}

	request.Header.Set("Content-Type", JSON_TYPE)
	request.Header.Set("Accept", JSON_TYPE)

	if h.config.username != "" && h.config.password != "" {
		request.SetBasicAuth(h.config.username, h.config.password)
	}

	return request
}

func (h *HttpClient) executeRequest(r *http.Request) *http.Response {
	resp, error := h.client.Do(r)

	if error != nil {
		panic(error)
	}

	return resp
}

func (h *HttpClient) Get(endpoint string) *http.Response {
	request := h.createRequest(endpoint, http.MethodGet, nil)
	return h.executeRequest(request)
}

func (h *HttpClient) Post(endpoint string, body io.Reader) *http.Response {
	request := h.createRequest(endpoint, http.MethodPost, body)
	return h.executeRequest(request)
}

func (h *HttpClient) Put(endpoint string, body io.Reader) *http.Response {
	request := h.createRequest(endpoint, http.MethodPut, body)
	return h.executeRequest(request)
}

func (h *HttpClient) Delete(endpoint string) *http.Response {
	request := h.createRequest(endpoint, http.MethodDelete, nil)
	return h.executeRequest(request)
}
