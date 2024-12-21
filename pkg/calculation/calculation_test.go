package calculation

import (
	"math"
	"testing"
)

const epsilon = 1e-6

func TestCalc(t *testing.T) {
	testCasesSuccess := []struct {
		name           string
		expression     string
		expectedResult float64
	}{
		{
			name:           "addition",
			expression:     "4+5",
			expectedResult: 9,
		},
		{
			name:           "addition",
			expression:     "4+7.48",
			expectedResult: 11.48,
		},
		{
			name:           "subtraction",
			expression:     "7-3",
			expectedResult: 4,
		},
		{
			name:           "subtraction",
			expression:     "19.49-4",
			expectedResult: 15.49,
		},
		{
			name:           "subtraction",
			expression:     "37.86-84.35",
			expectedResult: -46.49,
		},
		{
			name:           "multiplication",
			expression:     "4*7",
			expectedResult: 28,
		},
		{
			name:           "multiplication",
			expression:     "28.8*5.07",
			expectedResult: 146.016,
		},
		{
			name:           "division",
			expression:     "10/2",
			expectedResult: 5,
		},
		{
			name:           "division",
			expression:     "5.12/8",
			expectedResult: 0.64,
		},
		{
			name:           "priority",
			expression:     "2+2*2",
			expectedResult: 6,
		},
		{
			name:           "priority",
			expression:     "5/2+7*6*5+4/2",
			expectedResult: 214.5,
		},
		{
			name:           "mixed arithmetic operations",
			expression:     "24-19+4*5/2",
			expectedResult: 15,
		},
		{
			name:           "mixed arithmetic operations",
			expression:     "8.4-100.7/4+8.67*4.3",
			expectedResult: 20.506,
		},
		{
			name:           "expressions with parentheses",
			expression:     "48-((30+8)/4+9)/5",
			expectedResult: 44.3,
		},
		{
			name:           "expressions with parentheses",
			expression:     "53+87-(((((4+2)))))",
			expectedResult: 134,
		},
		{
			name:           "expressions with spaces",
			expression:     "   5 - 47  +8.9*     3/6     ",
			expectedResult: -37.55,
		},
	}

	for _, testCase := range testCasesSuccess {
		t.Run(testCase.name, func(t *testing.T) {
			val, err := Calc(testCase.expression)
			if err != nil {
				t.Fatalf("successful case %s returns error: %v", testCase.expression, err)
			}
			if math.Abs(val-testCase.expectedResult) > epsilon {
				t.Fatalf("%f should be equal %f", val, testCase.expectedResult)
			}
		})
	}

	testCasesFail := []struct {
		name       string
		expression string
	}{
		{
			name:       "Unmatched opening parenthesis",
			expression: "5*(3+2",
		},
		{
			name:       "Unmatched closing parenthesis",
			expression: "5+(3+2))",
		},
		{
			name:       "Division by zero",
			expression: "5/0",
		},
		{
			name:       "Invalid character",
			expression: "5+a",
		},
		{
			name:       "Empty expression",
			expression: "",
		},
		{
			name:       "Extra operators",
			expression: "5++2",
		},
		{
			name:       "Invalid syntax",
			expression: "5//2",
		},
		{
			name:       "Invalid syntax with multiple dots",
			expression: "5..5",
		},
		{
			name:       "Invalid syntax with multiple dots",
			expression: "5.4.3",
		},
		{
			name:       "Invalid parentheses operator placement",
			expression: "(+)",
		},
	}

	for _, testCase := range testCasesFail {
		t.Run(testCase.name, func(t *testing.T) {
			val, err := Calc(testCase.expression)
			if err == nil {
				t.Fatalf("expression %s is invalid but result %f was obtained", testCase.expression, val)
			}
		})
	}
}

func TestPerformOperation(t *testing.T) {
	val, err := performOperation('%', 5, 2)
	if err == nil {
		t.Fatalf("expression 5%%2 is invalid but result %f was obtained", val)
	}
}
