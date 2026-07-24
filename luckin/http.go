package luckin

import (
	"encoding/json"
	"fmt"
	"io"
	"shops/http"

	netHttp "net/http"
)

func request[T any](endpoint string, body any, token string, out *T) (*netHttp.Response, error) {
	headers := map[string]string{
		"X-LK-Tenant": "LKSG",
	}
	if token != "" {
		headers["Cookie"] = fmt.Sprintf("LK_PROD_ILUCKYINWAP_SID=%s; lk_isLogin=true; ", token)
	}
	resp, err := http.DoPost(baseUrl+endpoint, body, headers)
	if err != nil {
		return nil, fmt.Errorf("failed to make API request: %v", err)
	}
	defer resp.Body.Close()

	var response struct {
		BusiCode string `json:"busiCode"`
		Code     int    `json:"code"`
		Content  T      `json:"content"`
	}
	rawContent, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response body: %v", err)
	}
	if err := json.Unmarshal(rawContent, &response); err != nil {
		return nil, fmt.Errorf("failed to decode response: %v", err)
	}
	if response.Code != 1 {
		return nil, fmt.Errorf("api request failed with error code %d: %s - %s", response.Code, response.BusiCode, string(rawContent))
	}
	if response.BusiCode != "200" {
		return nil, fmt.Errorf("api request failed with busi code %s - %s", response.BusiCode, string(rawContent))
	}
	*out = response.Content
	return resp, nil
}
