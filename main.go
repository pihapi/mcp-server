package main

import (
	"bufio"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"time"
)

// Типы для JSON-RPC протокола
type Request struct {
	Jsonrpc string          `json:"jsonrpc"`
	Method  string          `json:"method"`
	Params  json.RawMessage `json:"params,omitempty"`
	ID      interface{}     `json:"id"`
}

type Response struct {
	Jsonrpc string      `json:"jsonrpc"`
	Result  interface{} `json:"result,omitempty"`
	Error   *Error      `json:"error,omitempty"`
	ID      interface{} `json:"id"`
}

type Error struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
}

// Типы для MCP
type Tool struct {
	Name        string      `json:"name"`
	Description string      `json:"description"`
	InputSchema InputSchema `json:"inputSchema"`
}

type InputSchema struct {
	Type       string              `json:"type"`
	Properties map[string]Property `json:"properties"`
	Required   []string            `json:"required"`
}

type Property struct {
	Type        string `json:"type"`
	Description string `json:"description"`
}

type CallToolParams struct {
	Name      string          `json:"name"`
	Arguments json.RawMessage `json:"arguments"`
}

type CallToolResult struct {
	Content []Content `json:"content"`
}

type Content struct {
	Type string `json:"type"`
	Text string `json:"text"`
}

// Аргументы для инструментов
type WeatherArgs struct {
	City string `json:"city"`
}

type CalculateArgs struct {
	Expression string `json:"expression"`
}

func main() {
	// Логирование в файл для отладки
	logFile, err := os.OpenFile("mcp-server.log", os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0666)
	if err == nil {
		log.SetOutput(logFile)
		defer logFile.Close()
	}

	log.Println("MCP Server started")

	scanner := bufio.NewScanner(os.Stdin)
	writer := bufio.NewWriter(os.Stdout)

	for scanner.Scan() {
		line := scanner.Text()
		log.Printf("Received: %s", line)

		var req Request
		if err := json.Unmarshal([]byte(line), &req); err != nil {
			log.Printf("Error parsing request: %v", err)
			continue
		}

		var response Response
		response.Jsonrpc = "2.0"
		response.ID = req.ID

		switch req.Method {
		case "initialize":
			response.Result = map[string]interface{}{
				"protocolVersion": "2024-11-05",
				"capabilities": map[string]interface{}{
					"tools": map[string]interface{}{},
				},
				"serverInfo": map[string]interface{}{
					"name":    "go-mcp-server",
					"version": "1.0.0",
				},
			}

		case "tools/list":
			tools := []Tool{
				{
					Name:        "get_weather",
					Description: "Get current weather for a city",
					InputSchema: InputSchema{
						Type: "object",
						Properties: map[string]Property{
							"city": {
								Type:        "string",
								Description: "City name",
							},
						},
						Required: []string{"city"},
					},
				},
				{
					Name:        "get_time",
					Description: "Get current time in different timezones",
					InputSchema: InputSchema{
						Type: "object",
						Properties: map[string]Property{
							"timezone": {
								Type:        "string",
								Description: "Timezone (e.g., UTC, EST, PST)",
							},
						},
						Required: []string{},
					},
				},
				{
					Name:        "calculate",
					Description: "Perform simple calculations",
					InputSchema: InputSchema{
						Type: "object",
						Properties: map[string]Property{
							"expression": {
								Type:        "string",
								Description: "Mathematical expression to evaluate",
							},
						},
						Required: []string{"expression"},
					},
				},
			}
			response.Result = map[string]interface{}{
				"tools": tools,
			}

		case "tools/call":
			var params CallToolParams
			if err := json.Unmarshal(req.Params, &params); err != nil {
				response.Error = &Error{
					Code:    -32602,
					Message: "Invalid params",
				}
				break
			}

			switch params.Name {
			case "get_weather":
				var args WeatherArgs
				if err := json.Unmarshal(params.Arguments, &args); err != nil {
					response.Error = &Error{
						Code:    -32602,
						Message: "Invalid arguments",
					}
					break
				}

				// Симуляция получения погоды
				weatherData := map[string]string{
					"Moscow":   "−5°C, снег",
					"London":   "8°C, дождь",
					"New York": "12°C, облачно",
					"Tokyo":    "15°C, ясно",
					"Sydney":   "25°C, солнечно",
				}

				weather, exists := weatherData[args.City]
				if !exists {
					weather = "20°C, переменная облачность"
				}

				response.Result = CallToolResult{
					Content: []Content{
						{
							Type: "text",
							Text: fmt.Sprintf("Погода в %s: %s", args.City, weather),
						},
					},
				}

			case "get_time":
				var args struct {
					Timezone string `json:"timezone"`
				}
				json.Unmarshal(params.Arguments, &args)

				loc := time.Local
				if args.Timezone != "" {
					if l, err := time.LoadLocation(args.Timezone); err == nil {
						loc = l
					}
				}

				currentTime := time.Now().In(loc).Format("15:04:05 MST")
				response.Result = CallToolResult{
					Content: []Content{
						{
							Type: "text",
							Text: fmt.Sprintf("Текущее время: %s", currentTime),
						},
					},
				}

			case "calculate":
				var args CalculateArgs
				if err := json.Unmarshal(params.Arguments, &args); err != nil {
					response.Error = &Error{
						Code:    -32602,
						Message: "Invalid arguments",
					}
					break
				}

				// Простой калькулятор (в реальном приложении используйте безопасный парсер)
				result := "Для безопасности поддерживаются только простые операции"

				response.Result = CallToolResult{
					Content: []Content{
						{
							Type: "text",
							Text: fmt.Sprintf("Выражение: %s\nРезультат: %s", args.Expression, result),
						},
					},
				}

			default:
				response.Error = &Error{
					Code:    -32601,
					Message: fmt.Sprintf("Tool not found: %s", params.Name),
				}
			}

		default:
			response.Error = &Error{
				Code:    -32601,
				Message: fmt.Sprintf("Method not found: %s", req.Method),
			}
		}

		// Отправка ответа
		responseBytes, _ := json.Marshal(response)
		log.Printf("Sending: %s", string(responseBytes))
		fmt.Fprintf(writer, "%s\n", responseBytes)
		writer.Flush()
	}

	if err := scanner.Err(); err != nil {
		log.Printf("Scanner error: %v", err)
	}
}
