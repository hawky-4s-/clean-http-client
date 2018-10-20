package http

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"
)

const (
	contentTypeJSON = "application/json"

	fixtureBaseURL   = "https://github.com/hawky-4s-"
	fixtureBasicJSON = "{ \"id\": 1 }"
)

func TestDefaultHttpConfig(t *testing.T) {
	defaultHttpConfig := NewDefaultHttpConfig(fixtureBaseURL)

	if defaultHttpConfig.baseURL != fixtureBaseURL {
		t.Errorf("Expected %s but got %s", fixtureBaseURL, defaultHttpConfig.baseURL)
	}
	if defaultHttpConfig.username != "" {
		t.Errorf("Expected \"\" but got %s", defaultHttpConfig.username)
	}
	if defaultHttpConfig.password != "" {
		t.Errorf("Expected \"\" but got %s", defaultHttpConfig.password)
	}
	if defaultHttpConfig.accept != contentTypeJSON {
		t.Errorf("Expected '%s' but got %s", contentTypeJSON, defaultHttpConfig.accept)
	}

}

func TestNewDefaultHttpClient(t *testing.T) {
	client := NewDefaultHttpClient(fixtureBaseURL)

	httpConfig := client.config

	if httpConfig.baseURL != fixtureBaseURL {
		t.Errorf("Expected %s but got %s", fixtureBaseURL, httpConfig.baseURL)
	}
	if httpConfig.username != "" {
		t.Errorf("Expected \"\" but got %s", httpConfig.username)
	}
	if httpConfig.password != "" {
		t.Errorf("Expected \"\" but got %s", httpConfig.password)
	}
	if httpConfig.accept != contentTypeJSON {
		t.Errorf("Expected '%s' but got %s", contentTypeJSON, httpConfig.accept)
	}
}

func TestNewHttpClient(t *testing.T) {
	customHTTPConfig := NewHttpConfig(fixtureBaseURL, "", "", contentTypeJSON)
	client := NewHttpClientWithConfig(customHTTPConfig)

	httpConfig := client.config

	if httpConfig.baseURL != fixtureBaseURL {
		t.Errorf("Expected %s but got %s", fixtureBaseURL, httpConfig.baseURL)
	}
	if httpConfig.username != "" {
		t.Errorf("Expected \"\", got %s", httpConfig.username)
	}
	if httpConfig.password != "" {
		t.Errorf("Expected \"\", got %s", httpConfig.password)
	}
	if httpConfig.accept != contentTypeJSON {
		t.Errorf("Expected '%s' but got %s", contentTypeJSON, httpConfig.accept)
	}
}

func TestHttpClient_GetFrom(t *testing.T) {
	server := mockServer(http.StatusOK, contentTypeJSON, fixtureBasicJSON)
	defer server.Close()

	client := createTestHTTPClient(server.URL)
	resp, _ := client.GetFrom("")

	assertResponseHasStatus(resp, http.StatusOK, t)
	assertResponseBodyIs(resp, fixtureBasicJSON, t)
}

func TestHttpClient_PostTo(t *testing.T) {
	server := mockEchoServer(200)
	defer server.Close()

	client := createTestHTTPClient(server.URL)
	resp, _ := client.PostTo("", strings.NewReader(fixtureBasicJSON))

	assertResponseHasStatus(resp, http.StatusOK, t)
	assertResponseBodyIs(resp, fixtureBasicJSON, t)
}

func TestHttpClient_PutTo(t *testing.T) {
	server := mockEchoServer(200)
	defer server.Close()

	client := createTestHTTPClient(server.URL)
	resp, _ := client.PutTo("", strings.NewReader(fixtureBasicJSON))

	assertResponseHasStatus(resp, http.StatusOK, t)
	assertResponseBodyIs(resp, fixtureBasicJSON, t)
}

func TestHttpClient_DeleteFrom(t *testing.T) {
	f := func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(200)
		w.Header().Set("Content-Type", contentTypeJSON)
	}
	server := mockServerWith(http.HandlerFunc(f))
	defer server.Close()

	client := createTestHTTPClient(server.URL)
	resp, _ := client.DeleteFrom("")

	assertResponseHasStatus(resp, http.StatusOK, t)
	assertResponseBodyIs(resp, "", t)
}

func TestHttpClient_GetRequest(t *testing.T) {
	client := createTestHTTPClient(fixtureBaseURL)
	req, _ := client.GetRequest("path")

	assertURLIs(req.URL, fixtureBaseURL+"/path", t)
}

func TestHttpClient_PostRequest(t *testing.T) {

}

func TestHttpClient_PutRequest(t *testing.T) {

}

func TestHttpClient_DeleteRequest(t *testing.T) {

}

func mockEchoServer(statusCode int) *httptest.Server {
	f := func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(statusCode)
		w.Header().Set("Content-Type", contentTypeJSON)

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
			w.Header().Set("Content-Type", contentTypeJSON)
			fmt.Fprint(w, "{}")
		}
	}

	return httptest.NewServer(handlerFunc)
}

func createTestHTTPClient(baseURL string) *HttpClient {
	config := NewHttpConfig(baseURL, "", "", contentTypeJSON)
	return NewHttpClientWithConfig(config)
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

func assertURLIs(actual *url.URL, expectedV interface{}, t *testing.T) {
	switch expected := expectedV.(type) {
	case string:
		if expected != actual.String() {
			t.Fail()
		}
	case *url.URL:
		if expected != actual {
			t.Fail()
		}
	default:
		t.Errorf("Type not supported %v", expected)
	}
}
