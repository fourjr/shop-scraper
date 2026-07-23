package db

import (
	"context"
	_ "embed"
	"fmt"
	"strings"

	"github.com/jackc/pgx/v4/pgxpool"
)

type Postgres struct {
	db *pgxpool.Pool
}

func NewPostgres(ctx context.Context, uri string) (*Postgres, error) {
	conn, err := pgxpool.Connect(ctx, uri)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to database: %w", err)
	}
	pg := &Postgres{db: conn}
	if err := pg.init(ctx); err != nil {
		return nil, fmt.Errorf("failed to initialize database tables: %w", err)
	}
	return pg, nil
}

func (c *Postgres) AddItems(ctx context.Context, items []Item, vendor string) error {
	if items == nil || len(items) == 0 {
		return nil
	}

	const fieldsPerItem = 8
	var query strings.Builder
	query.WriteString(`INSERT INTO entry (
		store_id, store_name, vendor, raw_data, waiting_cups, waiting_time, coordinates
	) VALUES `)

	args := make([]any, 0, len(items)*fieldsPerItem)
	for i, item := range items {
		if i > 0 {
			query.WriteString(", ")
		}

		parameter := i*fieldsPerItem + 1
		fmt.Fprintf(
			&query,
			"($%d, $%d, $%d, $%d::jsonb, $%d, $%d, point($%d, $%d))",
			parameter,
			parameter+1,
			parameter+2,
			parameter+3,
			parameter+4,
			parameter+5,
			parameter+6,
			parameter+7,
		)
		args = append(args, item.StoreId, item.StoreName, vendor, item.RawData, item.WaitingCups, item.WaitingTime, item.Longitude, item.Latitude)
	}

	if _, err := c.db.Exec(ctx, query.String(), args...); err != nil {
		return fmt.Errorf("failed to insert items: %w", err)
	}

	return nil
}

//go:embed init.sql
var initSQL string

func (c *Postgres) init(ctx context.Context) error {
	if _, err := c.db.Exec(ctx, initSQL); err != nil {
		return fmt.Errorf("failed to create database tables: %w", err)
	}
	return nil
}
