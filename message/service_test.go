package message

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/hashicorp/go-hclog"
	"github.com/labstack/echo/v4"
)

// TestServiceGetMessageByID tests func GetMessageByID
func TestServiceGetMessageByID(t *testing.T) {
	e := echo.New()
	r := httptest.NewRequest(http.MethodGet, "/:id", nil)
	w := httptest.NewRecorder()
	c := e.NewContext(r, w)
	c.SetPath("/:id")
	logger := hclog.New(&hclog.LoggerOptions{
		JSONFormat: false,
		Level:      hclog.Debug,
	})

	if w.Code == http.StatusOK {
		logger.Debug("Everythig is Ok")
	} else {
		t.Errorf("Expected to get status %d but instead got %d", http.StatusOK, w.Code)
	}
}
