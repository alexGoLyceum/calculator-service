package routes_test

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/alexGoLyceum/calculator-service/orchestrator/internal/transport/http/routes"
	"github.com/alexGoLyceum/calculator-service/orchestrator/mocks"

	"github.com/labstack/echo/v4"
	"github.com/stretchr/testify/assert"
	"go.uber.org/mock/gomock"
)

func TestRegisterRoutes(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockHandler := mocks.NewMockHandler(ctrl)
	e := echo.New()

	routes.RegisterRoutes(e, mockHandler)

	mockHandler.EXPECT().Register(gomock.Any()).Return(nil).Times(1)
	mockHandler.EXPECT().Login(gomock.Any()).Return(nil).Times(1)
	mockHandler.EXPECT().Calculate(gomock.Any()).Return(nil).Times(1)
	mockHandler.EXPECT().GetExpressions(gomock.Any()).Return(nil).Times(1)
	mockHandler.EXPECT().GetExpressionByID(gomock.Any()).Return(nil).Times(1)
	mockHandler.EXPECT().Ping(gomock.Any()).Return(nil).Times(1)

	req := httptest.NewRequest(http.MethodPost, "/api/v1/register", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	err := mockHandler.Register(c)
	if err != nil {
		t.Fatal(err)
	}
	assert.Equal(t, http.StatusOK, rec.Code)

	req = httptest.NewRequest(http.MethodPost, "/api/v1/login", nil)
	rec = httptest.NewRecorder()
	c = e.NewContext(req, rec)
	err = mockHandler.Login(c)
	if err != nil {
		t.Fatal(err)
	}
	assert.Equal(t, http.StatusOK, rec.Code)

	req = httptest.NewRequest(http.MethodPost, "/api/v1/calculate", nil)
	rec = httptest.NewRecorder()
	c = e.NewContext(req, rec)
	err = mockHandler.Calculate(c)
	if err != nil {
		t.Fatal(err)
	}
	assert.Equal(t, http.StatusOK, rec.Code)

	req = httptest.NewRequest(http.MethodGet, "/api/v1/expressions", nil)
	rec = httptest.NewRecorder()
	c = e.NewContext(req, rec)
	err = mockHandler.GetExpressions(c)
	if err != nil {
		t.Fatal(err)
	}
	assert.Equal(t, http.StatusOK, rec.Code)

	req = httptest.NewRequest(http.MethodGet, "/api/v1/expressions/1234", nil)
	rec = httptest.NewRecorder()
	c = e.NewContext(req, rec)
	err = mockHandler.GetExpressionByID(c)
	if err != nil {
		t.Fatal(err)
	}
	assert.Equal(t, http.StatusOK, rec.Code)

	req = httptest.NewRequest(http.MethodGet, "/api/v1/ping", nil)
	rec = httptest.NewRecorder()
	c = e.NewContext(req, rec)
	err = mockHandler.Ping(c)
	if err != nil {
		t.Fatal(err)
	}
	assert.Equal(t, http.StatusOK, rec.Code)
}
