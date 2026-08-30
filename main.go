package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"shops/chagee"
	"shops/db"
	"shops/heytea"
	"shops/luckin"
	"shops/models"

	"github.com/joho/godotenv"
	"golang.org/x/sync/errgroup"
)

func doChagee(db models.DBClient) error {
	items, errors := chagee.RequestAll()
	if len(errors) > 0 {
		for _, err := range errors {
			log.Printf("[chagee] Error occurred: %v", err)
		}
	}
	if items != nil {
		err := db.AddItems(context.Background(), items, "chagee")
		if err != nil {
			return fmt.Errorf("failed to add items to database: %w", err)
		}
		log.Printf("[chagee] Successfully added %d items", len(items))
	}
	return nil
}

func doLuckin(ctx context.Context, db models.DBClient) error {
	items, errors := luckin.RequestAll(ctx, db)
	if len(errors) > 0 {
		for _, err := range errors {
			log.Printf("[luckin] Error occurred: %v", err)
		}
	}
	if items != nil {
		err := db.AddItems(context.Background(), items, "luckin")
		if err != nil {
			return fmt.Errorf("failed to add items to database: %w", err)
		}
		log.Printf("[luckin] Successfully added %d items", len(items))
	}
	return nil
}

func doHeytea(db models.DBClient) error {
	items, errors := heytea.RequestAll()
	if len(errors) > 0 {
		for _, err := range errors {
			log.Printf("[heytea] Error occurred: %v", err)
		}
	}
	if items != nil {
		err := db.AddItems(context.Background(), items, "heytea")
		if err != nil {
			return fmt.Errorf("failed to add items to database: %w", err)
		}
		log.Printf("[heytea] Successfully added %d items", len(items))
	}
	return nil
}

func main() {
	err := godotenv.Load()
	if err != nil && os.Getenv("CONTAINER") == "" {
		panic(fmt.Sprintf("failed to load .env file: %v", err))
	}

	db, err := db.NewPostgres(
		context.Background(),
		os.Getenv("POSTGRES_URI"),
	)
	if err != nil {
		panic(fmt.Sprintf("failed to connect to database: %v", err))
	}

	var group errgroup.Group

	group.Go(func() error {
		if err := doChagee(db); err != nil {
			return fmt.Errorf("failed to process Chagee items: %w", err)
		}
		return nil
	})

	group.Go(func() error {
		if err := doLuckin(context.Background(), db); err != nil {
			return fmt.Errorf("failed to process Luckin items: %w", err)
		}
		return nil
	})

	group.Go(func() error {
		if err := doHeytea(db); err != nil {
			return fmt.Errorf("failed to process Heytea items: %w", err)
		}
		return nil
	})
	if err := group.Wait(); err != nil {
		panic(err)
	}
}
