# clean-http-client

This package provides a biased golang clean http-client implementation using json.
Meaning it will instantiate always a clean

## Installation

```
go get github.com/hawky-4s-/clean-http-client
```

## Usage

```
// instantiate a http-client with default parameters

baseUrl := "https://api.github.com"
client := http.DefaultNewHttpClient(baseUrl)


// instantiate a http-client with custom parameters

baseUrl := "https://api.github.com"
// basic auth
username := "foo"
password := "bar"
// accept header
accept := "application/json"

config := http.NewHttpConfig(baseUrl, username, password, accept)
client := http.NewHttpClient


client.
```
