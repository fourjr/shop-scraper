package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"shops/chagee"
	"shops/db"

	"github.com/joho/godotenv"
)

func main() {
	err := godotenv.Load()
	if err != nil && os.Getenv("DOCKER") == "" {
		panic(fmt.Sprintf("failed to load .env file: %v", err))
	}
	db, err := db.NewPostgres(context.Background(), os.Getenv("POSTGRES_URI"))
	if err != nil {
		panic(fmt.Sprintf("failed to connect to database: %v", err))
	}

	// chagee
	chageeItems, err := chagee.RequestAll()
	if err != nil {
		panic(fmt.Sprintf("failed to request items from Chagee: %v", err))
	}
	err = db.AddItems(context.Background(), chageeItems)
	if err != nil {
		panic(fmt.Sprintf("failed to add items to database: %v", err))
	}
	log.Printf("[chagee] Successfully added %d items", len(*chageeItems))
}
