package tools

import (
	"encoding/json"
	"fmt"
	"mcp-server/server"
)

type WeatherTool struct{}

type WeatherArgs struct {
	City string `json:"city"`
}

func NewWeatherTool() *WeatherTool {
	return &WeatherTool{}
}

func (t *WeatherTool) GetDefinition() server.Tool {
	return server.Tool{
		Name:        "get_weather",
		Description: "Get current weather for a city",
		InputSchema: server.InputSchema{
			Type: "object",
			Properties: map[string]server.Property{
				"city": {
					Type:        "string",
					Description: "City name",
				},
			},
			Required: []string{"city"},
		},
	}
}

func (t *WeatherTool) Execute(arguments json.RawMessage) (*server.CallToolResult, error) {
	var args WeatherArgs
	if err := json.Unmarshal(arguments, &args); err != nil {
		return nil, fmt.Errorf("invalid arguments: %v", err)
	}

	// Симуляция данных о погоде
	weatherData := map[string]string{
		"Moscow":   "−5°C, снег",
		"London":   "8°C, дождь",
		"New York": "12°C, облачно",
		"Tokyo":    "15°C, ясно",
		"Sydney":   "25°C, солнечно",
		"Paris":    "10°C, туман",
		"Berlin":   "6°C, пасмурно",
	}

	weather, exists := weatherData[args.City]
	if !exists {
		weather = "20°C, переменная облачность"
	}

	return &server.CallToolResult{
		Content: []server.Content{
			{
				Type: "text",
				Text: fmt.Sprintf("Погода в %s: %s", args.City, weather),
			},
		},
	}, nil
}
