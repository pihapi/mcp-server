package tools

import (
	"encoding/json"
	"fmt"
	"mcp-server/server"
	"time"
)

type TimeTool struct{}

type TimeArgs struct {
	Timezone string `json:"timezone"`
}

func NewTimeTool() *TimeTool {
	return &TimeTool{}
}

func (t *TimeTool) GetDefinition() server.Tool {
	return server.Tool{
		Name:        "get_time",
		Description: "Get current time in different timezones",
		InputSchema: server.InputSchema{
			Type: "object",
			Properties: map[string]server.Property{
				"timezone": {
					Type:        "string",
					Description: "Timezone (e.g., UTC, America/New_York, Europe/London)",
				},
			},
			Required: []string{},
		},
	}
}

func (t *TimeTool) Execute(arguments json.RawMessage) (*server.CallToolResult, error) {
	var args TimeArgs
	if len(arguments) > 0 {
		if err := json.Unmarshal(arguments, &args); err != nil {
			return nil, fmt.Errorf("invalid arguments: %v", err)
		}
	}

	loc := time.Local
	locationName := "Local"

	if args.Timezone != "" {
		loadedLoc, err := time.LoadLocation(args.Timezone)
		if err == nil {
			loc = loadedLoc
			locationName = args.Timezone
		} else {
			return nil, fmt.Errorf("invalid timezone: %s", args.Timezone)
		}
	}

	currentTime := time.Now().In(loc)
	formattedTime := currentTime.Format("15:04:05 MST, Monday, January 2, 2006")

	return &server.CallToolResult{
		Content: []server.Content{
			{
				Type: "text",
				Text: fmt.Sprintf("Текущее время в %s: %s", locationName, formattedTime),
			},
		},
	}, nil
}
