package db

import (
	"context"
	"fmt"
	"log"
	"os"
	"testing"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/roman-adamchik/simplebank/util"
)

var (
	testQueries *Queries
	testPool    *pgxpool.Pool
)

func TestMain(m *testing.M) {
	config, err := util.LoadConfig("../..")
	if err != nil {
		log.Fatal("cannot load config:", err)
	}

	ctx := context.Background()

	testPool, err = pgxpool.New(ctx, config.DBSource)

	if err != nil {
		fmt.Fprintf(os.Stderr, "Unable to create pool: %v\n", err)
		os.Exit(1)
	}

	if err := testPool.Ping(ctx); err != nil {
		fmt.Fprintf(os.Stderr, "Unable to connect to database: %v\n", err)
		os.Exit(1)
	}

	testQueries = New(testPool)
	exitCode := m.Run()
	testPool.Close()

	os.Exit(exitCode)
}
