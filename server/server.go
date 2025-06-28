package server

import (
	"bufio"
	"encoding/json"
	"fmt"
	"log"
	"os"
)

type MCPServer struct {
	tools   map[string]ToolHandler
	logger  *log.Logger
	scanner *bufio.Scanner
	writer  *bufio.Writer
}

func NewMCPServer() *MCPServer {
	// Настройка логирования
	logFile, err := os.OpenFile("mcp-server.log", os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0666)
	var logger *log.Logger
	if err == nil {
		logger = log.New(logFile, "", log.LstdFlags)
	} else {
		logger = log.New(os.Stderr, "", log.LstdFlags)
	}

	return &MCPServer{
		tools:   make(map[string]ToolHandler),
		logger:  logger,
		scanner: bufio.NewScanner(os.Stdin),
		writer:  bufio.NewWriter(os.Stdout),
	}
}

func (s *MCPServer) RegisterTool(name string, handler ToolHandler) {
	s.tools[name] = handler
	s.logger.Printf("Registered tool: %s", name)
}

func (s *MCPServer) Run() {
	s.logger.Println("MCP Server started")

	for s.scanner.Scan() {
		line := s.scanner.Text()
		s.logger.Printf("Received: %s", line)

		var req Request
		if err := json.Unmarshal([]byte(line), &req); err != nil {
			s.logger.Printf("Error parsing request: %v", err)
			continue
		}

		response := s.handleRequest(&req)

		responseBytes, _ := json.Marshal(response)
		s.logger.Printf("Sending: %s", string(responseBytes))
		fmt.Fprintf(s.writer, "%s\n", responseBytes)
		s.writer.Flush()
	}

	if err := s.scanner.Err(); err != nil {
		s.logger.Printf("Scanner error: %v", err)
	}
}

func (s *MCPServer) handleRequest(req *Request) *Response {
	response := &Response{
		Jsonrpc: "2.0",
		ID:      req.ID,
	}

	switch req.Method {
	case "initialize":
		response.Result = s.handleInitialize()
	case "tools/list":
		response.Result = s.handleToolsList()
	case "tools/call":
		result, err := s.handleToolCall(req.Params)
		if err != nil {
			response.Error = err
		} else {
			response.Result = result
		}
	default:
		response.Error = &Error{
			Code:    -32601,
			Message: fmt.Sprintf("Method not found: %s", req.Method),
		}
	}

	return response
}

func (s *MCPServer) handleInitialize() interface{} {
	return map[string]interface{}{
		"protocolVersion": "2024-11-05",
		"capabilities": map[string]interface{}{
			"tools": map[string]interface{}{},
		},
		"serverInfo": map[string]interface{}{
			"name":    "go-mcp-server",
			"version": "1.0.0",
		},
	}
}

func (s *MCPServer) handleToolsList() interface{} {
	tools := make([]Tool, 0, len(s.tools))
	for _, handler := range s.tools {
		tools = append(tools, handler.GetDefinition())
	}

	return map[string]interface{}{
		"tools": tools,
	}
}

func (s *MCPServer) handleToolCall(params json.RawMessage) (*CallToolResult, *Error) {
	var callParams CallToolParams
	if err := json.Unmarshal(params, &callParams); err != nil {
		return nil, &Error{
			Code:    -32602,
			Message: "Invalid params",
		}
	}

	handler, exists := s.tools[callParams.Name]
	if !exists {
		return nil, &Error{
			Code:    -32601,
			Message: fmt.Sprintf("Tool not found: %s", callParams.Name),
		}
	}

	result, err := handler.Execute(callParams.Arguments)
	if err != nil {
		return nil, &Error{
			Code:    -32603,
			Message: err.Error(),
		}
	}

	return result, nil
}
