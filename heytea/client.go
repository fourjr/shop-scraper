package heytea

import (
	"encoding/json"
	"fmt"
	"io"
	"shops/http"
	"shops/models"
)

const baseUrl = "https://app-sg.heytea-co.com/api"

func request(method string, endpoint string, payload any) (io.ReadCloser, error) {
	url := baseUrl + endpoint

	resp, err := http.Do(method, url, payload, map[string]string{
		"Accept":                             "application/prs.heytea.v1+json",
		"Accept-Language":                    "en-US",
		"GMT-Zone":                           "+08:00",
		"Client":                             "2",
		"X-client":                           "app",
		"client-system":                      "android",
		"x-region-code":                      "SG",
		"x-region-id":                        "3",
		"Heytea-Secure-Transmission-Tenant":  "heyteago-android",
		"Heytea-Secure-Transmission-Version": "1",
		"User-Agent":                         "okhttp/4.12.0",
	})
	if err != nil {
		return nil, fmt.Errorf("failed to make API request: %v", err)
	}
	return resp.Body, nil
}

func getShops() ([]*shop, error) {
	// https://app-sg.heytea-co.com/api/service-smc/openapi/app/user/closest/shop-list?country_code=702&user_location=103.8292679%2C1.3044123&city_code=s702000001
	body, err := request("GET", "/service-smc/openapi/app/user/closest/shop-list", map[string]string{
		"country_code":  "702",
		"user_location": "103.8292679,1.3044123",
		"city_code":     "s702000001",
	})
	if err != nil {
		return nil, fmt.Errorf("failed to request shops: %v", err)
	}
	defer body.Close()

	return parseResponse[*shop](body)
}

func findShopById(shops []*shop, id int) *shop {
	for _, shop := range shops {
		if shop.Id == id {
			return shop
		}
	}
	return nil
}

func getShopTime(shops []*shop) ([]models.DBItem, error) {
	body, err := request("POST", "/service-ofc/openapi/agent/expect-time/shop/list", map[string]any{
		"expectTimeByShopIds": func() []map[string]int {
			ids := make([]map[string]int, 0)
			for _, shop := range shops {
				if !shop.IsOpen {
					continue
				}
				ids = append(ids, map[string]int{"shopId": shop.Id})
			}
			return ids
		}(),
		"isTakeaway": false,
		"showTime":   true,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to request shops: %v", err)
	}
	defer body.Close()

	shopTimes, err := parseResponse[*shopTime](body)
	if err != nil {
		return nil, fmt.Errorf("failed to parse shop times: %v", err)
	}

	dbItems := make([]models.DBItem, len(shopTimes))
	for i, shopTime := range shopTimes {
		shop := findShopById(shops, shopTime.ShopId)
		if shop == nil {
			return nil, fmt.Errorf("shop with ID %d not found", shopTime.ShopId)
		}
		dbItems[i] = models.DBItem{
			StoreId:     fmt.Sprintf("%d", shopTime.ShopId),
			StoreName:   shop.Name,
			Longitude:   shop.Longitude,
			Latitude:    shop.Latitude,
			WaitingCups: shopTime.MakingCups,
			WaitingTime: shopTime.ExpectTime,
			RawData:     shop.RawData,
		}
	}
	return dbItems, nil
}

func parseResponse[T rawDataItem](body io.ReadCloser) ([]T, error) {
	var response apiResponse
	if err := json.NewDecoder(body).Decode(&response); err != nil {
		return nil, fmt.Errorf("failed to decode response: %v", err)
	}
	if response.Code != 0 {
		return nil, fmt.Errorf("api request failed with error code %d: %s", response.Code, response.Message)
	}

	items := make([]T, 0, len(response.Data))
	for _, rawItem := range response.Data {
		var item T
		if err := json.Unmarshal(rawItem, &item); err != nil {
			return nil, fmt.Errorf("failed to decode item: %v", err)
		}

		// Keep the complete item object for insertion into the raw_data JSONB column.
		item.SetRawData(string(rawItem))
		items = append(items, item)
	}

	return items, nil
}

func RequestAll() ([]models.DBItem, []error) {
	shops, err := getShops()
	if err != nil {
		return nil, []error{fmt.Errorf("API (shops) request failed: %v", err)}
	}

	items, err := getShopTime(shops)
	if err != nil {
		return nil, []error{fmt.Errorf("API (shop times) request failed: %v", err)}
	}
	return items, nil
}
