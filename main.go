package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"shops/chagee"
	"shops/db"
	"shops/luckin"

	"github.com/joho/godotenv"
)

func doChagee(db db.Client) error {
	items, err := chagee.RequestAll()
	if err != nil {
		return fmt.Errorf("failed to request items from Chagee: %w", err)
	}
	err = db.AddItems(context.Background(), items, "chagee")
	if err != nil {
		return fmt.Errorf("failed to add items to database: %w", err)
	}
	log.Printf("[chagee] Successfully added %d items", len(*items))
	return nil
}

func doLuckin(db db.Client) error {
	items, err := luckin.RequestAll()
	if err != nil {
		return fmt.Errorf("failed to request items from Luckin: %w", err)
	}
	err = db.AddItems(context.Background(), items, "luckin")
	if err != nil {
		return fmt.Errorf("failed to add items to database: %w", err)
	}
	log.Printf("[luckin] Successfully added %d items", len(*items))
	return nil
}

func main() {
	err := godotenv.Load()
	if err != nil && os.Getenv("DOCKER") == "" {
		panic(fmt.Sprintf("failed to load .env file: %v", err))
	}
	db, err := db.NewPostgres(context.Background(), os.Getenv("POSTGRES_URI"))
	if err != nil {
		panic(fmt.Sprintf("failed to connect to database: %v", err))
	}

	if err := doChagee(db); err != nil {
		panic(fmt.Sprintf("failed to process Chagee items: %v", err))
	}
	if err := doLuckin(db); err != nil {
		panic(fmt.Sprintf("failed to process Luckin items: %v", err))
	}
}
