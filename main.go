package main

import (
	"mcp-server/server"
	"mcp-server/tools"
)

func main() {
	// Создаем сервер
	mcpServer := server.NewMCPServer()

	// Регистрируем все инструменты
	tools.RegisterAllTools(mcpServer)

	// Запускаем сервер
	mcpServer.Run()
}
