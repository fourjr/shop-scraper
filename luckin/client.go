package luckin

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"os"
	"shops/http"
	"shops/models"
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

type LuckinAccounter interface {
	GetLuckinAccount(context.Context, int) (*models.LuckinAccount, error)
	UpdateLuckinToken(context.Context, models.LuckinAccount) (string, error)
}

func getConsideration(deptId int) int {
	currHr := time.Now().Hour() + time.Now().Day()*24
	return deptId + currHr
}

func request(ctx context.Context, am LuckinAccounter, shop Shop) (io.ReadCloser, error) {
	consideration := getConsideration(shop.DeptId)
	token, err := am.GetLuckinAccount(ctx, consideration)
	if err != nil {
		return nil, fmt.Errorf("failed to get luckin account: %v", err)
	}

	body := buildRequest(shop.DeptId)
	reader, err := json.Marshal(body)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request body: %v", err)
	}

	resp, err := http.DoPost(baseUrl+"/api/capi/resource/isalestradecapi/order/preview",
		reader, map[string]string{
			"Content-Type":    "application/json",
			"Accept-Language": "en-US",
			"Cookie":          fmt.Sprintf("LK_PROD_ILUCKYINWAP_SID=%s; lk_isLogin=true; ", *token.Token),
		})
	if err != nil {
		return nil, fmt.Errorf("failed to make API request: %v", err)
	}

	return resp.Body, nil
}

var ShopClosedError = errors.New("shop is closed")
var AccountError = errors.New("account error")

func parseResponse(shop Shop, body io.ReadCloser) (*models.DBItem, error) {
	rawContent, err := io.ReadAll(body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response body: %v", err)
	}
	var response apiResponse
	if err := json.Unmarshal(rawContent, &response); err != nil {
		return nil, fmt.Errorf("failed to decode response: %v", err)
	}
	if response.Code == 7 {
		return nil, ShopClosedError
	}
	if response.Code == 5 {
		return nil, AccountError
	}
	if response.Code != 1 {
		return nil, fmt.Errorf("api request failed with error code %d: %s - %s", response.Code, response.BusiCode, string(rawContent))
	}
	if response.BusiCode != "200" {
		return nil, fmt.Errorf("api request failed with busi code %s - %s", response.BusiCode, string(rawContent))
	}

	msWaitTime := response.Content.AboutTime - response.Content.ProductList[0].AddTime
	rawData, err := json.Marshal(shop)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal shop data: %v", err)
	}
	return &models.DBItem{
		StoreId:     fmt.Sprintf("%d-%s", shop.DeptId, shop.ShopId),
		StoreName:   shop.StoreName,
		Longitude:   fmt.Sprintf("%.6f", shop.Longitude),
		Latitude:    fmt.Sprintf("%.6f", shop.Latitude),
		WaitingTime: msWaitTime / 1000 / 60,
		RawData:     string(rawData),
	}, nil
}

func RequestAll(ctx context.Context, accounter LuckinAccounter) (allItems []models.DBItem, errors []error) {
	shops, err := getShops()
	if err != nil {
		return nil, []error{fmt.Errorf("failed to get shops: %v", err)}
	}
	noSkip := os.Getenv("LUCKIN_NOSKIP")

	for _, shop := range *shops {
		if noSkip == "" {
			if shop.DeptId != 977 {
				// TEMP ONLY GENEO
				continue
			}
		}
		if !shop.Open {
			continue
		}
		response, err := request(ctx, accounter, shop)
		defer time.Sleep(500 * time.Millisecond)
		if err != nil {
			errors = append(errors, fmt.Errorf("API request failed for shop %s: %v", shop.ShopId, err))
			continue
		}
		defer response.Close()

		items, err := parseResponse(shop, response)
		if err != nil {
			if err == ShopClosedError {
				continue
			}
			if err == AccountError {
				// Attempt to update the token and retry the request
				account, err := accounter.GetLuckinAccount(ctx, getConsideration(shop.DeptId))
				if err != nil {
					errors = append(errors, fmt.Errorf("failed to get luckin account for shop %s: %v", shop.ShopId, err))
					continue
				}
				_, err = accounter.UpdateLuckinToken(ctx, *account)
				if err != nil {
					errors = append(errors, fmt.Errorf("failed to update token for shop %s: %v", shop.ShopId, err))
					continue
				}
				// Retry the request with the new token
				response, err = request(ctx, accounter, shop)
				if err != nil {
					errors = append(errors, fmt.Errorf("API request failed after token update for shop %s: %v", shop.ShopId, err))
					continue
				}
				defer response.Close()
				items, err = parseResponse(shop, response)
				if err != nil {
					errors = append(errors, fmt.Errorf("failed to parse API response after token update for shop %s: %v", shop.ShopId, err))
					continue
				}
			} else {
				errors = append(errors, fmt.Errorf("failed to parse API response for shop %s: %v", shop.ShopId, err))
				continue
			}
		}

		allItems = append(allItems, *items)
	}

	return allItems, errors
}
