package http

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"time"
)

const defaultTimeout = 10 * time.Second

func DoPost(url string, body any, headers map[string]string) (*http.Response, error) {
	client := &http.Client{
		Timeout: defaultTimeout,
	}
	reader, err := json.Marshal(body)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request body: %v", err)
	}
	req, err := http.NewRequest(http.MethodPost, url, bytes.NewBuffer(reader))
	if err != nil {
		return nil, err
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Accept-Language", "en-US")
	req.Header.Set("Language", "en-US")
	for key, value := range headers {
		req.Header.Set(key, value)
	}

	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	if resp.StatusCode < http.StatusOK || resp.StatusCode >= http.StatusMultipleChoices {
		return nil, fmt.Errorf("request failed with status: %s", resp.Status)
	}
	return resp, nil
}

func DoGet(url string, query url.Values, headers map[string]string) (*http.Response, error) {
	client := &http.Client{
		Timeout: defaultTimeout,
	}
	req, err := http.NewRequest(http.MethodGet, url, nil)
	if err != nil {
		return nil, err
	}
	req.URL.RawQuery = query.Encode()

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Accept-Language", "en-US")
	req.Header.Set("Language", "en-US")
	for key, value := range headers {
		req.Header.Set(key, value)
	}

	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	if resp.StatusCode < http.StatusOK || resp.StatusCode >= http.StatusMultipleChoices {
		return nil, fmt.Errorf("request failed with status: %s", resp.Status)
	}
	return resp, nil
}

func Do(method string, queryUrl string, body any, headers map[string]string) (*http.Response, error) {
	switch method {
	case http.MethodPost:
		return DoPost(queryUrl, body, headers)
	case http.MethodGet:
		query := url.Values{}
		bodyMap, ok := body.(map[string]string)
		if !ok {
			return nil, fmt.Errorf("body must be a map[string]string for GET requests")
		}
		for key, value := range bodyMap {
			query.Set(key, value)
		}
		return DoGet(queryUrl, query, headers)
	default:
		return nil, fmt.Errorf("unsupported HTTP method: %s", method)
	}
}
