package tools

import (
	"mcp-server/server"
)

// RegisterAllTools регистрирует все доступные инструменты
func RegisterAllTools(s *server.MCPServer) {
	s.RegisterTool("get_weather", NewWeatherTool())
	s.RegisterTool("get_time", NewTimeTool())
	s.RegisterTool("calculate", NewCalculatorTool())
}
