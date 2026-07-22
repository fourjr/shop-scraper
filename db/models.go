package db

import "context"

type Item struct {
	StoreId       string
	StoreName     string
	WaitingCups   int
	WaitingTime   int
	RunningStatus int
	Longitude     string
	Latitude      string
	RawData       string
}

type Client interface {
	AddItems(ctx context.Context, items *[]Item, vendor string) error
}
