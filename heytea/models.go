package heytea

import (
	"encoding/json"
)

type apiRequest struct {
	PageSize    int    `json:"pageSize"`
	ChannelCode string `json:"channelCode"`
}

type apiResponse struct {
	Code    int               `json:"code"`
	Message string            `json:"message"`
	Data    []json.RawMessage `json:"data"`
}

type rawDataItem interface {
	SetRawData(string)
}

type shop struct {
	Id        int    `json:"id"`
	Name      string `json:"store"`
	Longitude string `json:"longitude"`
	Latitude  string `json:"latitude"`
	IsOpen    bool   `json:"is_open"`
	RawData   string `json:"-"`
}

func (s *shop) SetRawData(raw string) {
	s.RawData = raw
}

type shopTime struct {
	ShopId       int    `json:"shopId"`
	MakingCups   int    `json:"makingCups"`
	MakingOrder  int    `json:"makingOrder"`
	ExpectTime   int    `json:"expectTime"`
	TakeawayTime int    `json:"takeawayTime"`
	RawData      string `json:"-"`
}

func (s *shopTime) SetRawData(raw string) {
	s.RawData = raw
}
