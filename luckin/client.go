package luckin

import (
	"encoding/json"
	"fmt"
	"io"
	"os"
	"shops/db"
	"shops/http"
	"strings"
	"time"
)

const baseUrl = "https://in.luckincoffee.com"

type apiRequest struct {
	DeliveryType int       `json:"deliveryType"`
	DeptId       string    `json:"deptId"`
	PromotionSel []int     `json:"promotionSel"`
	ProductList  []product `json:"productList"`
}

type product struct {
	SPUCode string `json:"spuCode"`
	SKUCode string `json:"skuCode"`
	Type    int    `json:"type"`
	Num     int    `json:"num"`
	Index   int    `json:"index"`
}

type apiResponse struct {
	BusiCode string `json:"busiCode"`
	Code     int    `json:"code"`
	Content  struct {
		AboutTime   int `json:"aboutTime"`
		ProductList []struct {
			AddTime int `json:"addTime"`
		} `json:"productList"`
	} `json:"content"`
}

func buildRequest(deptId int) apiRequest {
	skuCode := os.Getenv("LUCKIN_SKU_CODE")
	spuCode := os.Getenv("LUCKIN_SPU_CODE")
	if skuCode == "" || spuCode == "" {
		panic("LUCKIN_SKU_CODE and LUCKIN_SPU_CODE environment variables must be set")
	}
	return apiRequest{
		DeptId:       fmt.Sprintf("%d", deptId),
		PromotionSel: []int{},
		ProductList: []product{
			{SPUCode: spuCode, SKUCode: skuCode, Type: 0, Num: 1, Index: 0},
		},
	}
}

func request(shop Shop) (io.ReadCloser, error) {
	tokens := os.Getenv("LUCKIN_TOKENS")
	if tokens == "" {
		panic("LUCKIN_TOKENS environment variable must be set")
	}

	allTokens := strings.Split(tokens, ",")
	currHr := time.Now().Hour() + time.Now().Day()*24
	consideration := shop.DeptId + currHr
	token := allTokens[consideration%len(allTokens)]

	body := buildRequest(shop.DeptId)
	reader, err := json.Marshal(body)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request body: %v", err)
	}

	resp, err := http.DoPost(baseUrl+"/api/capi/resource/isalestradecapi/order/preview",
		reader, map[string]string{
			"Content-Type": "application/json",
			"Cookie":       fmt.Sprintf("LK_PROD_ILUCKYINWAP_SID=%s; lk_isLogin=true; ", token),
		})
	if err != nil {
		return nil, fmt.Errorf("failed to make API request: %v", err)
	}

	return resp, nil
}

func parseResponse(shop Shop, body io.ReadCloser) (*db.Item, error) {
	var response apiResponse
	if err := json.NewDecoder(body).Decode(&response); err != nil {
		return nil, fmt.Errorf("failed to decode response: %v", err)
	}
	if response.Code != 1 {
		return nil, fmt.Errorf("api request failed with error code %d: %s", response.Code, response.BusiCode)
	}
	if response.BusiCode != "200" {
		return nil, fmt.Errorf("api request failed with busi code %s", response.BusiCode)
	}

	msWaitTime := response.Content.AboutTime - response.Content.ProductList[0].AddTime
	rawData, err := json.Marshal(shop)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal shop data: %v", err)
	}
	return &db.Item{
		StoreId:     fmt.Sprintf("%d-%s", shop.DeptId, shop.ShopId),
		StoreName:   shop.StoreName,
		Longitude:   fmt.Sprintf("%.6f", shop.Longitude),
		Latitude:    fmt.Sprintf("%.6f", shop.Latitude),
		WaitingTime: msWaitTime / 1000 / 60,
		RawData:     string(rawData),
	}, nil
}

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
			"Content-Type": "application/json",
			"X-LK-Tenant":  "LKSG",
		})
	if err != nil {
		return nil, fmt.Errorf("failed to make API request: %v", err)
	}
	defer resp.Close()

	var response struct {
		BusiCode string `json:"busiCode"`
		Code     int    `json:"code"`
		Content  struct {
			ShopList []Shop `json:"shopList"`
		} `json:"content"`
	}
	rawContent, err := io.ReadAll(resp)
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

func RequestAll() (*[]db.Item, error) {
	shops, err := getShops()
	if err != nil {
		return nil, fmt.Errorf("failed to get shops: %v", err)
	}

	var allItems []db.Item
	for _, shop := range *shops {
		if shop.DeptId != 977 && shop.DeptId != 910 && shop.DeptId != 1180 && shop.DeptId != 286 && shop.DeptId != 950 && shop.DeptId != 309 {
			// TEMP ONLY GENEO
			continue
		}
		if !shop.Open {
			continue
		}
		response, err := request(shop)
		if err != nil {
			return nil, fmt.Errorf("API request failed for shop %s: %v", shop.ShopId, err)
		}
		defer response.Close()

		items, err := parseResponse(shop, response)
		if err != nil {
			return nil, fmt.Errorf("failed to parse API response for shop %s: %v", shop.ShopId, err)
		}

		allItems = append(allItems, *items)
		time.Sleep(500 * time.Millisecond)
	}

	return &allItems, nil
}
