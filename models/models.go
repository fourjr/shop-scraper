package models

import (
	"context"
)

type DBItem struct {
	StoreId       string
	StoreName     string
	WaitingCups   int
	WaitingTime   int
	RunningStatus int
	Longitude     string
	Latitude      string
	RawData       string
}

type DBClient interface {
	AddItems(ctx context.Context, items []DBItem, vendor string) error
	GetLuckinAccount(ctx context.Context, consideration int) (*LuckinAccount, error)
	UpdateLuckinToken(ctx context.Context, account LuckinAccount) (string, error)
}

type LuckinAccount struct {
	Email    string
	Password string
	Token    *string
}
