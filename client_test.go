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
	CONTENT_TYPE_JSON = "application/json"

	fixture_baseUrl   = "https://github.com/camunda"
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
	if defaultHttpConfig.accept != CONTENT_TYPE_JSON {
		t.Errorf("Expected '%s' but got %s", CONTENT_TYPE_JSON, defaultHttpConfig.accept)
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
	if httpConfig.accept != CONTENT_TYPE_JSON {
		t.Errorf("Expected '%s' but got %s", CONTENT_TYPE_JSON, httpConfig.accept)
	}
}

func TestNewHttpClient(t *testing.T) {
	customHttpConfig := NewHttpConfig(fixture_baseUrl, "", "", CONTENT_TYPE_JSON)
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
	if httpConfig.accept != CONTENT_TYPE_JSON {
		t.Errorf("Expected '%s' but got %s", CONTENT_TYPE_JSON, httpConfig.accept)
	}
}

func TestHttpClient_GetFrom(t *testing.T) {
	server := mockServer(http.StatusOK, CONTENT_TYPE_JSON, fixture_basicJson)
	defer server.Close()

	client := createTestHttpClient(server.URL)
	resp, _ := client.GetFrom("")

	assertResponseHasStatus(resp, http.StatusOK, t)
	assertResponseBodyIs(resp, fixture_basicJson, t)
}

func TestHttpClient_PostTo(t *testing.T) {
	server := mockEchoServer(200)
	defer server.Close()

	client := createTestHttpClient(server.URL)
	resp, _ := client.PostTo("", strings.NewReader(fixture_basicJson))

	assertResponseHasStatus(resp, http.StatusOK, t)
	assertResponseBodyIs(resp, fixture_basicJson, t)
}

func TestHttpClient_PutTo(t *testing.T) {
	server := mockEchoServer(200)
	defer server.Close()

	client := createTestHttpClient(server.URL)
	resp, _ := client.PutTo("", strings.NewReader(fixture_basicJson))

	assertResponseHasStatus(resp, http.StatusOK, t)
	assertResponseBodyIs(resp, fixture_basicJson, t)
}

func TestHttpClient_DeleteFrom(t *testing.T) {
	f := func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(200)
		w.Header().Set("Content-Type", CONTENT_TYPE_JSON)
	}
	server := mockServerWith(http.HandlerFunc(f))
	defer server.Close()

	client := createTestHttpClient(server.URL)
	resp, _ := client.DeleteFrom("")

	assertResponseHasStatus(resp, http.StatusOK, t)
	assertResponseBodyIs(resp, "", t)
}

func TestHttpClient_GetRequest(t *testing.T) {
	client := createTestHttpClient(fixture_baseUrl)
	req, _ := client.GetRequest("path")

	assertUrlIs(req.URL, fixture_baseUrl+"/path", t)
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
		w.Header().Set("Content-Type", CONTENT_TYPE_JSON)

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
			w.Header().Set("Content-Type", CONTENT_TYPE_JSON)
			fmt.Fprint(w, "{}")
		}
	}

	return httptest.NewServer(handlerFunc)
}

func createTestHttpClient(baseUrl string) *HttpClient {
	config := NewHttpConfig(baseUrl, "", "", CONTENT_TYPE_JSON)
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

func assertUrlIs(actual *url.URL, expectedV interface{}, t *testing.T) {
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
