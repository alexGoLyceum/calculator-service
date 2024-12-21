package calculation

import (
	"strconv"
	"strings"
	"unicode"
)

var priority = map[rune]int{
	'+': 1, '-': 1,
	'*': 2, '/': 2,
}

func Calc(expression string) (float64, error) {
	expression = strings.ReplaceAll(expression, " ", "")

	if !validateBrackets(expression) {
		return 0, ErrInvalidExpression
	}

	numbers := make([]float64, 0)
	operations := make([]rune, 0)

	for i := 0; i < len(expression); i++ {
		char := rune(expression[i])
		switch {
		case char == '(':
			operations = append(operations, char)
		case char == ')':
			for len(operations) > 0 && operations[len(operations)-1] != '(' {
				if err := processOperation(&numbers, &operations); err != nil {
					return 0, err
				}
			}
			operations = operations[:len(operations)-1]
		case unicode.IsDigit(char) || char == '.':
			startIndex := i
			for i < len(expression) && (unicode.IsDigit(rune(expression[i])) || rune(expression[i]) == '.') {
				i++
			}
			number, err := strconv.ParseFloat(expression[startIndex:i], 64)
			if err != nil {
				return 0, err
			}
			numbers = append(numbers, number)
			i--
		case strings.ContainsRune("+-*/", char):
			for len(operations) > 0 && priority[char] <= priority[operations[len(operations)-1]] {
				if err := processOperation(&numbers, &operations); err != nil {
					return 0, err
				}
			}
			operations = append(operations, char)
		default:
			return 0, ErrInvalidExpression
		}
	}

	for len(operations) > 0 {
		if err := processOperation(&numbers, &operations); err != nil {
			return 0, err
		}
	}

	if len(numbers) != 1 {
		return 0, ErrInvalidExpression
	}

	return numbers[0], nil
}

func validateBrackets(expression string) bool {
	count := 0
	for _, char := range expression {
		if char == '(' {
			count++
		} else if char == ')' {
			count--
		}
		if count < 0 {
			return false
		}
	}
	return count == 0
}

func performOperation(operation rune, num1, num2 float64) (float64, error) {
	switch operation {
	case '+':
		return num1 + num2, nil
	case '-':
		return num1 - num2, nil
	case '*':
		return num1 * num2, nil
	case '/':
		if num2 == 0 {
			return 0, ErrDivisionByZero
		}
		return num1 / num2, nil
	}
	return 0, ErrInvalidExpression
}

func processOperation(numbers *[]float64, operations *[]rune) error {
	if len(*numbers) < 2 || len(*operations) == 0 {
		return ErrInvalidExpression
	}

	num1 := (*numbers)[len(*numbers)-2]
	num2 := (*numbers)[len(*numbers)-1]
	*numbers = (*numbers)[:len(*numbers)-2]

	operation := (*operations)[len(*operations)-1]
	*operations = (*operations)[:len(*operations)-1]

	num, err := performOperation(operation, num1, num2)
	if err != nil {
		return err
	}
	*numbers = append(*numbers, num)

	return nil
}
