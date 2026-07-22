package http

import (
	"bytes"
	"fmt"
	"io"
	"net/http"
	"time"
)

const defaultTimeout = 10 * time.Second

func DoPost(url string, body []byte, headers map[string]string) (io.ReadCloser, error) {
	client := &http.Client{
		Timeout: defaultTimeout,
	}
	req, err := http.NewRequest(http.MethodPost, url, bytes.NewBuffer(body))
	if err != nil {
		return nil, err
	}

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
	return resp.Body, nil
}
