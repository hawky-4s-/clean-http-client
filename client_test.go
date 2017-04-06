package http

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

const (
	fixture_baseUrl   = "https://github.com/hawky-4s-"
	fixture_basicJson = "{ \"id\": 1 }"
)

func TestDefaultHttpConfig(t *testing.T) {
	defaultHttpConfig := DefaultHttpConfig(fixture_baseUrl)

	if defaultHttpConfig.baseUrl != fixture_baseUrl {
		t.Errorf("Expected %s but got %s", fixture_baseUrl, defaultHttpConfig.baseUrl)
	}
	if defaultHttpConfig.username != "" {
		t.Errorf("Expected \"\" but got %s", defaultHttpConfig.username)
	}
	if defaultHttpConfig.password != "" {
		t.Errorf("Expected \"\" but got %s", defaultHttpConfig.password)
	}
	if defaultHttpConfig.accept != "application/json" {
		t.Errorf("Expected \"application/json\" but got %s", defaultHttpConfig.accept)
	}

}

func TestNewDefaultHttpClient(t *testing.T) {
	client := NewDefaultHttpClient(fixture_baseUrl)

	httpConfig := client.config

	if httpConfig.baseUrl != fixture_baseUrl {
		t.Errorf("Expected %s but got %s", fixture_baseUrl, httpConfig.baseUrl)
	}
	if httpConfig.username != "" {
		t.Errorf("Expected \"\" but got %s", httpConfig.username)
	}
	if httpConfig.password != "" {
		t.Errorf("Expected \"\" but got %s", httpConfig.password)
	}
	if httpConfig.accept != "application/json" {
		t.Errorf("Expected \"application/json\" but got %s", httpConfig.accept)
	}
}

func TestNewHttpClient(t *testing.T) {
	customHttpConfig := NewHttpConfig(fixture_baseUrl, "", "", "application/json")
	client := NewHttpClient(customHttpConfig)

	httpConfig := client.config

	if httpConfig.baseUrl != fixture_baseUrl {
		t.Errorf("Expected %s but got %s", fixture_baseUrl, httpConfig.baseUrl)
	}
	if httpConfig.username != "" {
		t.Errorf("Expected \"\", got %s", httpConfig.username)
	}
	if httpConfig.password != "" {
		t.Errorf("Expected \"\", got %s", httpConfig.password)
	}
	if httpConfig.accept != "application/json" {
		t.Errorf("Expected \"application/json\", got %s", httpConfig.accept)
	}
}

func TestHttpClient_Get(t *testing.T) {
	server := mockServer(http.StatusOK, "application/json", fixture_basicJson)
	defer server.Close()

	client := createTestHttpClient(server.URL)
	resp := client.Get("")

	assertResponseHasStatus(resp, http.StatusOK, t)
	assertResponseBodyIs(resp, fixture_basicJson, t)
}

func TestHttpClient_Post(t *testing.T) {
	server := mockEchoServer(200)
	defer server.Close()

	client := createTestHttpClient(server.URL)
	resp := client.Post("", strings.NewReader(fixture_basicJson))

	assertResponseHasStatus(resp, http.StatusOK, t)
	assertResponseBodyIs(resp, fixture_basicJson, t)
}

func TestHttpClient_Put(t *testing.T) {
	server := mockEchoServer(200)
	defer server.Close()

	client := createTestHttpClient(server.URL)
	resp := client.Post("", strings.NewReader(fixture_basicJson))

	assertResponseHasStatus(resp, http.StatusOK, t)
	assertResponseBodyIs(resp, fixture_basicJson, t)
}

func TestHttpClient_Delete(t *testing.T) {
	f := func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(200)
		w.Header().Set("Content-Type", "application/json")
	}
	server := mockServerWith(http.HandlerFunc(f))
	defer server.Close()

	client := createTestHttpClient(server.URL)
	resp := client.Post("", strings.NewReader(fixture_basicJson))

	assertResponseHasStatus(resp, http.StatusOK, t)
	assertResponseBodyIs(resp, "", t)
}

func mockEchoServer(statusCode int) *httptest.Server {
	f := func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(statusCode)
		w.Header().Set("Content-Type", "application/json")

		body := ""
		if r.Body != nil {
			defer r.Body.Close()
			bodyAsBytes, _ := ioutil.ReadAll(r.Body)
			body = string(bodyAsBytes)
		}
		fmt.Fprint(w, body)
	}

	return httptest.NewServer(http.HandlerFunc(f))
}

// mockServer returns a pointer to a server to handle the get call.
func mockServer(statusCode int, contentType string, body string) *httptest.Server {
	f := func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(statusCode)
		w.Header().Set("Content-Type", contentType)
		fmt.Fprint(w, body)
	}

	return httptest.NewServer(http.HandlerFunc(f))
}

func mockServerWith(handlerFunc http.HandlerFunc) *httptest.Server {
	if handlerFunc == nil {
		handlerFunc = func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(200)
			w.Header().Set("Content-Type", "application/json")
			fmt.Fprint(w, "{}")
		}
	}

	return httptest.NewServer(handlerFunc)
}

func createTestHttpClient(baseUrl string) *HttpClient {
	config := NewHttpConfig(baseUrl, "", "", "application/json")
	return NewHttpClient(config)
}

func assertResponseHasStatus(resp *http.Response, statusCode int, t *testing.T) {
	if resp.StatusCode != statusCode {
		t.Errorf("Expected statuscode %d, got %d.", statusCode, resp.StatusCode)
	}
}

func assertResponseBodyIs(resp *http.Response, expectedValue string, t *testing.T) {
	defer resp.Body.Close()
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		t.Error("Unexpected error")
	}
	expectedValueAsBytes := []byte(expectedValue)
	if bytes.Compare(body, expectedValueAsBytes) != 0 {
		t.Errorf("Expected body %s, got %s.", expectedValueAsBytes, body)
	}
}
