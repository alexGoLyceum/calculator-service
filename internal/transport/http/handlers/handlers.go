package handlers

import (
	"errors"
	"net/http"
	"strconv"

	"github.com/alexGoLyceum/calculator-service/pkg/calculation"

	"github.com/labstack/echo/v4"
)

type Request struct {
	Expression string `json:"expression"`
}

type Response struct {
	Result string `json:"result,omitempty"`
	Error  string `json:"error,omitempty"`
}

func CalculateHandler(c echo.Context) error {
	var request Request
	if err := c.Bind(&request); err != nil {
		return c.JSON(http.StatusBadRequest, Response{Error: "Invalid JSON"})
	}

	result, err := calculation.Calc(request.Expression)
	if err != nil {
		if errors.Is(err, calculation.ErrInvalidExpression) {
			return c.JSON(http.StatusUnprocessableEntity, Response{Error: "Expression is not valid"})
		}
		return c.JSON(http.StatusInternalServerError, Response{Error: "Internal server error"})
	}

	return c.JSON(http.StatusOK, Response{Result: strconv.FormatFloat(result, 'f', -1, 64)})
}
