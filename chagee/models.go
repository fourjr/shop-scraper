package chagee

import "shops/db"

type Item struct {
	StoreId       string `json:"storeNo"`
	StoreName     string `json:"storeName"`
	WaitingCups   int    `json:"waitingCups"`
	WaitingTime   int    `json:"waitingTime"`
	RunningStatus int    `json:"runningStatus"`
	Longitude     string `json:"longitude"`
	Latitude      string `json:"latitude"`
	RawData       string `json:"-"`
}

func (i Item) ToDBItem() db.Item {
	return db.Item(i)
}
