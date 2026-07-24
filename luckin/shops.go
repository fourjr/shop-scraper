package luckin

import (
	"fmt"
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
	body := cityPageRequest{
		Type:  1,
		Limit: 1000,
	}

	var response struct {
		ShopList []Shop `json:"shopList"`
	}
	_, err := request("/api/capi/resource/locate/shop/inCityPage", body, "", &response)
	if err != nil {
		return nil, fmt.Errorf("failed to make API request: %v", err)
	}

	return &response.ShopList, nil
}
