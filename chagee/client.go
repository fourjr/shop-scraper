package chagee

import (
	"encoding/json"
	"fmt"
	"io"
	"shops/db"
	"shops/http"
)

const baseUrl = "https://api-sea.chagee.com/api/navigation/store/list"

type apiRequest struct {
	PageSize    int    `json:"pageSize"`
	ChannelCode string `json:"channelCode"`
}

type apiResponse struct {
	ErrCode string `json:"errcode"`
	ErrMsg  string `json:"errmsg"`
	Data    struct {
		PageList []json.RawMessage `json:"pageList"`
	} `json:"data"`
}

func request() (io.ReadCloser, error) {
	body := apiRequest{
		PageSize:    100,
		ChannelCode: "H5",
	}
	reader, err := json.Marshal(body)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request body: %v", err)
	}

	resp, err := http.DoPost(baseUrl, reader, map[string]string{
		"Content-Type":         "application/json",
		"Language":             "en-us",
		"Region":               "SG",
		"devicetimezoneregion": "Asia/Singapore",
		"Channel":              "H5",
	})
	if err != nil {
		return nil, fmt.Errorf("failed to make API request: %v", err)
	}
	return resp, nil
}

func parseResponse(body io.ReadCloser) (*[]db.Item, error) {
	var response apiResponse
	if err := json.NewDecoder(body).Decode(&response); err != nil {
		return nil, fmt.Errorf("failed to decode response: %v", err)
	}
	if response.ErrCode != "0" {
		return nil, fmt.Errorf("api request failed with error code %s: %s", response.ErrCode, response.ErrMsg)
	}

	items := make([]db.Item, 0, len(response.Data.PageList))
	for _, rawItem := range response.Data.PageList {
		var item db.Item
		if err := json.Unmarshal(rawItem, &item); err != nil {
			return nil, fmt.Errorf("failed to decode item: %v", err)
		}

		if item.RunningStatus == 1 {
			continue
		}
		// Keep the complete item object for insertion into the raw_data JSONB column.
		item.RawData = string(rawItem)
		items = append(items, item)
	}

	return &items, nil
}

func RequestAll() (*[]db.Item, error) {
	response, err := request()
	if err != nil {
		return nil, fmt.Errorf("API request failed: %v", err)
	}
	defer response.Close()
	items, err := parseResponse(response)
	if err != nil {
		return nil, fmt.Errorf("failed to parse API response: %v", err)
	}
	// log.Printf("Successfully retrieved %d items", len(*items))
	return items, nil
}
