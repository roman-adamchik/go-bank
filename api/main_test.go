package api

import (
	"os"
	"testing"

	"github.com/gin-gonic/gin"
)

// TestMain sets up the test environment before running all tests in this package.
// It configures Gin to run in test mode, which disables debug logging and
// improves test performance.
func TestMain(m *testing.M) {
	gin.SetMode(gin.TestMode)

	os.Exit(m.Run())
}
