package api

import (
	"testing"

	mockdb "github.com/roman-adamchik/simplebank/db/mock"
	"go.uber.org/mock/gomock"
)

func TestServerStartInvalidAddress(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	store := mockdb.NewMockStore(ctrl)
	server := newTestServer(t, store)

	// Invalid address (missing port) should cause an immediate error.
	err := server.Start("invalid")
	if err == nil {
		t.Fatalf("expected error for invalid address, got nil")
	}
}
