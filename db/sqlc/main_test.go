package db

import (
	"context"
	"fmt"
	"os"
	"testing"

	"github.com/jackc/pgx/v5/pgxpool"
)

const (
	dbSource = "postgresql://root:secret@localhost:5432/simple_bank?sslmode=disable"
)

var (
	testQueries *Queries
	testPool    *pgxpool.Pool
)

func TestMain(m *testing.M) {
	var err error
	ctx := context.Background()

	testPool, err = pgxpool.New(ctx, dbSource)

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
