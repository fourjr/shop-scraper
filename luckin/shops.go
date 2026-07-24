package luckin

import (
	"encoding/json"
	"fmt"
	"io"
	"shops/http"
)

type cityPageRequest struct {
	Type  int `json:"type"`
	Limit int `json:"limit"`
}

type Shop struct {
	DeptId    int     `json:"deptId"`
	ShopId    string  `json:"shopNo"`
	StoreName string  `json:"shopName"`
	Longitude float64 `json:"locationLongitude"`
	Latitude  float64 `json:"locationLatitude"`
	Open      bool    `json:"beOpening"`
}

func getShops() (*[]Shop, error) {
	request := cityPageRequest{
		Type:  1,
		Limit: 1000,
	}
	reader, err := json.Marshal(request)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request body: %v", err)
	}

	resp, err := http.DoPost(baseUrl+"/api/capi/resource/locate/shop/inCityPage",
		reader, map[string]string{
			"Content-Type":    "application/json",
			"Accept-Language": "en-US",
			"X-LK-Tenant":     "LKSG",
		})
	if err != nil {
		return nil, fmt.Errorf("failed to make API request: %v", err)
	}
	defer resp.Body.Close()

	var response struct {
		BusiCode string `json:"busiCode"`
		Code     int    `json:"code"`
		Content  struct {
			ShopList []Shop `json:"shopList"`
		} `json:"content"`
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

	return &response.Content.ShopList, nil
}
