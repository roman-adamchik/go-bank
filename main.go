package main

import (
	"context"
	"fmt"
	"log"
	"os"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/roman-adamchik/simplebank/api"
	db "github.com/roman-adamchik/simplebank/db/sqlc"
	"github.com/roman-adamchik/simplebank/util"
)

func main() {
	config, err := util.LoadConfig(".")
	if err != nil {
		log.Fatal("cannot load config:", err)
	}

	ctx := context.Background()

	pool, err := pgxpool.New(ctx, config.DBSource)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Unable to create pool: %v\n", err)
		os.Exit(1)
	}
	defer pool.Close()

	if err := pool.Ping(ctx); err != nil {
		fmt.Fprintf(os.Stderr, "Unable to connect to database: %v\n", err)
		os.Exit(1)
	}

	store := db.NewStore(pool)
	server := api.NewServer(store)

	err = server.Start(config.ServerAddress)
	if err != nil {
		log.Fatal("Cannot start server:", err)
	}
}
