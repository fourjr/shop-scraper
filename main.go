package main

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"strings"
	"time"

	_ "embed"

	"github.com/jackc/pgx/v5"
	"github.com/joho/godotenv"
)

const baseUrl = "https://api-sea.chagee.com/api/navigation/store/list"

type Client struct {
	db   *pgx.Conn
	http *http.Client
}

type APIRequest struct {
	PageSize    int    `json:"pageSize"`
	ChannelCode string `json:"channelCode"`
}

type APIResponse struct {
	ErrCode string `json:"errcode"`
	ErrMsg  string `json:"errmsg"`
	Data    struct {
		PageList []json.RawMessage `json:"pageList"`
	} `json:"data"`
}
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

func (c *Client) request() (*[]Item, error) {
	body := APIRequest{
		PageSize:    100,
		ChannelCode: "H5",
	}
	reader, err := json.Marshal(body)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request body: %v", err)
	}
	req, err := http.NewRequest(http.MethodPost, baseUrl, bytes.NewBuffer(reader))
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %v", err)
	}

	// 3. Set your custom and standard HTTP headers
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer YOUR_SECRET_TOKEN")
	req.Header.Set("Language", "en-us")
	req.Header.Set("Region", "SG")
	req.Header.Set("devicetimezoneregion", "Asia/Singapore")
	req.Header.Set("Channel", "H5")
	resp, err := c.http.Do(req)
	if err != nil {
		return nil, fmt.Errorf("request execution failed: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode < http.StatusOK || resp.StatusCode >= http.StatusMultipleChoices {
		return nil, fmt.Errorf("request failed with status: %s", resp.Status)
	}
	return c.parseResponse(resp.Body)
}

func (c *Client) parseResponse(body io.Reader) (*[]Item, error) {
	var response APIResponse
	if err := json.NewDecoder(body).Decode(&response); err != nil {
		return nil, fmt.Errorf("failed to decode response: %v", err)
	}
	if response.ErrCode != "0" {
		return nil, fmt.Errorf("api request failed with error code %s: %s", response.ErrCode, response.ErrMsg)
	}

	items := make([]Item, 0, len(response.Data.PageList))
	for _, rawItem := range response.Data.PageList {
		var item Item
		if err := json.Unmarshal(rawItem, &item); err != nil {
			return nil, fmt.Errorf("failed to decode item: %v", err)
		}

		if item.RunningStatus == 1 {
			continue
		}
		// Keep the complete item object for insertion into the raw_data JSONB column.
		item.RawData = string(rawItem)
		items = append(items, item)
	}

	return &items, nil
}

func (c *Client) updateDb(ctx context.Context, items *[]Item) error {
	if items == nil || len(*items) == 0 {
		return nil
	}

	const fieldsPerItem = 7
	var query strings.Builder
	query.WriteString(`INSERT INTO entry (
		store_id, store_name, vendor, raw_data, waiting_cups, waiting_time, coordinates
	) VALUES `)

	args := make([]any, 0, len(*items)*fieldsPerItem)
	for i, item := range *items {
		if i > 0 {
			query.WriteString(", ")
		}

		parameter := i*fieldsPerItem + 1
		fmt.Fprintf(
			&query,
			"($%d, $%d, 'chagee', $%d::jsonb, $%d, $%d, point($%d, $%d))",
			parameter,
			parameter+1,
			parameter+2,
			parameter+3,
			parameter+4,
			parameter+5,
			parameter+6,
		)
		args = append(args, item.StoreNo, item.StoreName, item.RawData, item.WaitingCups, item.WaitingTime, item.Longitude, item.Latitude)
	}

	if _, err := c.db.Exec(ctx, query.String(), args...); err != nil {
		return fmt.Errorf("failed to insert items: %w", err)
	}

	return nil
}

func newClient() (*Client, error) {
	err := godotenv.Load()
	if err != nil && os.Getenv("DOCKER") == "" {
		return nil, fmt.Errorf("failed to load .env file: %w", err)
	}
	conn, err := pgx.Connect(context.Background(), os.Getenv("POSTGRES_URI"))
	if err != nil {
		return nil, fmt.Errorf("failed to connect to database: %w", err)
	}
	httpClient := &http.Client{
		Timeout: 10 * time.Second,
	}
	c := &Client{
		db:   conn,
		http: httpClient,
	}
	if err := c.initSql(); err != nil {
		return nil, fmt.Errorf("failed to initialize database tables: %w", err)
	}
	return c, nil
}

//go:embed init.sql
var initSQL string

func (c *Client) initSql() error {
	if _, err := c.db.Exec(context.Background(), initSQL); err != nil {
		return fmt.Errorf("failed to create database tables: %w", err)
	}
	return nil
}

func main() {
	c, err := newClient()
	if err != nil {
		log.Fatalf("failed to create client: %v", err)
	}
	items, err := c.request()
	if err != nil {
		log.Fatalf("API request failed: %v", err)
	}
	if err := c.updateDb(context.Background(), items); err != nil {
		log.Fatalf("Database update failed: %v", err)
	}
	now := time.Now().Format(time.RFC3339)
	log.Printf("[%s] Successfully added %d items", now, len(*items))
}
