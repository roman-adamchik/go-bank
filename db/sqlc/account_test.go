package db

import "testing"

func TestPingDB(t *testing.T) {
	if testQueries == nil {
		t.Fatal("testQueries not initialized")
	}
}
