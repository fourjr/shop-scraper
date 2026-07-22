package db

import "context"

type Item struct {
	StoreNo       string `json:"storeNo"`
	StoreName     string `json:"storeName"`
	WaitingCups   int    `json:"waitingCups"`
	WaitingTime   int    `json:"waitingTime"`
	RunningStatus int    `json:"runningStatus"`
	Longitude     string `json:"longitude"`
	Latitude      string `json:"latitude"`
	RawData       string `json:"-"`
}

type Client interface {
	AddItems(ctx context.Context, items *[]Item) error
}
