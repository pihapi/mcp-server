package tools

import (
	"encoding/json"
	"fmt"
	"mcp-server/server"
	"strconv"
	"strings"
)

type CalculatorTool struct{}

type CalculatorArgs struct {
	Expression string `json:"expression"`
}

func NewCalculatorTool() *CalculatorTool {
	return &CalculatorTool{}
}

func (t *CalculatorTool) GetDefinition() server.Tool {
	return server.Tool{
		Name:        "calculate",
		Description: "Perform simple calculations (supports +, -, *, /)",
		InputSchema: server.InputSchema{
			Type: "object",
			Properties: map[string]server.Property{
				"expression": {
					Type:        "string",
					Description: "Mathematical expression (e.g., '2 + 2', '10 * 5')",
				},
			},
			Required: []string{"expression"},
		},
	}
}

func (t *CalculatorTool) Execute(arguments json.RawMessage) (*server.CallToolResult, error) {
	var args CalculatorArgs
	if err := json.Unmarshal(arguments, &args); err != nil {
		return nil, fmt.Errorf("invalid arguments: %v", err)
	}

	result, err := t.evaluateSimpleExpression(args.Expression)
	if err != nil {
		return nil, err
	}

	return &server.CallToolResult{
		Content: []server.Content{
			{
				Type: "text",
				Text: fmt.Sprintf("Выражение: %s\nРезультат: %s", args.Expression, result),
			},
		},
	}, nil
}

// Простой калькулятор для базовых операций
func (t *CalculatorTool) evaluateSimpleExpression(expr string) (string, error) {
	expr = strings.TrimSpace(expr)

	// Поддержка простых бинарных операций
	operators := []string{"+", "-", "*", "/"}

	for _, op := range operators {
		parts := strings.Split(expr, op)
		if len(parts) == 2 {
			left, err1 := strconv.ParseFloat(strings.TrimSpace(parts[0]), 64)
			right, err2 := strconv.ParseFloat(strings.TrimSpace(parts[1]), 64)

			if err1 != nil || err2 != nil {
				continue
			}

			var result float64
			switch op {
			case "+":
				result = left + right
			case "-":
				result = left - right
			case "*":
				result = left * right
			case "/":
				if right == 0 {
					return "", fmt.Errorf("деление на ноль")
				}
				result = left / right
			}

			return fmt.Sprintf("%.2f", result), nil
		}
	}

	return "", fmt.Errorf("неподдерживаемое выражение. Используйте формат: число операция число")
}
