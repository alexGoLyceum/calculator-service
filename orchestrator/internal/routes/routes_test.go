package routes_test

import (
	"bytes"
	"encoding/json"
	"log"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/alexGoLyceum/calculator-service/orchestrator/internal/config"
	"github.com/alexGoLyceum/calculator-service/orchestrator/internal/server"
	"github.com/alexGoLyceum/calculator-service/orchestrator/internal/taskmanager"
	"github.com/alexGoLyceum/calculator-service/pkg/logging"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"go.uber.org/zap"
)

func setupOrchestrator() *server.Orchestrator {
	orchestrator := config.OrchestratorConfig{
		Host: "localhost",
		Port: 8080,
	}
	taskManager := config.TaskManagerConfig{
		TimeAdditionMS:        1000,
		TimeSubtractionMS:     1000,
		TimeMultiplicationsMS: 1000,
		TimeDivisionsMS:       1000,
	}
	logger := logging.LogConfig{
		Level: "debug",
		Path:  "logs.log",
	}
	cfg := config.Config{TaskManager: taskManager, Log: logger, Orchestrator: orchestrator}

	l := zap.NewExample()
	o := server.NewOrchestrator(&cfg, l)
	if o == nil {
		log.Fatal("Failed to initialize orchestrator")
	}
	return o
}

func TestCalculateHandler(t *testing.T) {
	o := setupOrchestrator()

	reqBody := `{"expression":"2+2"}`
	req := httptest.NewRequest(http.MethodPost, "/api/v1/calculate", bytes.NewBufferString(reqBody))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()

	o.Echo.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusCreated, rec.Code)

	var resp struct {
		ID    *uuid.UUID `json:"id,omitempty"`
		Error string     `json:"error,omitempty"`
	}
	err := json.NewDecoder(rec.Body).Decode(&resp)
	assert.NoError(t, err)
	assert.NotNil(t, resp.ID)
	assert.Empty(t, resp.Error)
}

func TestGetExpressionsHandler(t *testing.T) {
	o := setupOrchestrator()

	req := httptest.NewRequest(http.MethodGet, "/api/v1/expressions", nil)
	rec := httptest.NewRecorder()

	o.Echo.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusOK, rec.Code)

	var resp struct {
		Expressions []taskmanager.Expression `json:"expressions"`
	}
	err := json.NewDecoder(rec.Body).Decode(&resp)
	assert.NoError(t, err)
	assert.Empty(t, resp.Expressions)
}

func TestGetExpressionByIDHandler(t *testing.T) {
	o := setupOrchestrator()

	id := uuid.New()
	req := httptest.NewRequest(http.MethodGet, "/api/v1/expressions/"+id.String(), nil)
	rec := httptest.NewRecorder()
	o.Echo.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusNotFound, rec.Code)

	var resp struct {
		Error string `json:"error"`
	}
	err := json.NewDecoder(rec.Body).Decode(&resp)
	assert.NoError(t, err)
	assert.NotEmpty(t, resp.Error)
}

func TestGetTaskHandler(t *testing.T) {
	o := setupOrchestrator()

	req := httptest.NewRequest(http.MethodGet, "/api/v1/internal/task", nil)
	rec := httptest.NewRecorder()

	o.Echo.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusNotFound, rec.Code)
}

func TestSetTaskResultHandler(t *testing.T) {
	o := setupOrchestrator()

	task := taskmanager.Task{
		ID:           uuid.New(),
		ExpressionID: uuid.New(),
		Operator:     "+",
	}
	result := 4.0

	requestBody, _ := json.Marshal(map[string]interface{}{
		"task":   task,
		"result": result,
	})

	req := httptest.NewRequest(http.MethodPost, "/api/v1/internal/task", bytes.NewBuffer(requestBody))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()

	o.Echo.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusNotFound, rec.Code)

	var resp struct {
		Error string `json:"error"`
	}
	err := json.NewDecoder(rec.Body).Decode(&resp)
	assert.NoError(t, err)
	assert.NotEmpty(t, resp.Error)
}
