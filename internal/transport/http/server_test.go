package http_test

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/alexGoLyceum/calculator-service/internal/config"
	h "github.com/alexGoLyceum/calculator-service/internal/transport/http"

	"github.com/stretchr/testify/assert"
	"go.uber.org/zap"
)

func TestServer_StartStop(t *testing.T) {
	logger, _ := zap.NewDevelopment()
	cfg := &config.ServerConfig{
		Host: "localhost",
		Port: 8080,
	}

	server := h.NewServer(cfg, logger)
	go func() {
		_ = server.Start()
	}()
	time.Sleep(500 * time.Millisecond)

	resp, err := http.Get("http://localhost:8080")
	assert.NoError(t, err)
	assert.Equal(t, http.StatusNotFound, resp.StatusCode)
}

func TestRoutes_Integration(t *testing.T) {
	logger, _ := zap.NewDevelopment()
	cfg := &config.ServerConfig{
		Host: "localhost",
		Port: 8080,
	}

	server := h.NewServer(cfg, logger)
	e := server.Echo
	req := httptest.NewRequest(http.MethodPost, "/api/v1/calculate", nil)
	rec := httptest.NewRecorder()
	e.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusUnprocessableEntity, rec.Code)
}
