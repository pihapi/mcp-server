Сервер включает:

Три инструмента: получение погоды, времени и калькулятор

Логирование в файл mcp-server.log для отладки

Обработку ошибок

Полную поддержку протокола MCP

После сборки вы получите исполняемый файл, который можно использовать с Claude Desktop, LM-Studio (tested 0.3.17).


### Build

    Windows: go build -o mcp-server.exe main.go
    
    Linux/Mac: go build -o mcp-server main.go

build.bat (для Windows)

    @echo off
    echo Building MCP Server...
    go build -o mcp-server.exe main.go
    echo Build complete: mcp-server.exe

build.sh (для Linux/Mac)

    #!/bin/bash
    echo "Building MCP Server..."
    go build -o mcp-server main.go
    chmod +x mcp-server
    echo "Build complete: mcp-server"

### Test

test.bat (для тестирования на Windows)

    @echo off
    echo Testing initialize...
    echo {"jsonrpc":"2.0","method":"initialize","id":1} | mcp-server.exe

    echo.
    echo Testing tools/list...
    echo {"jsonrpc":"2.0","method":"tools/list","id":2} | mcp-server.exe

    echo.
    echo Testing get_weather...
    echo {"jsonrpc":"2.0","method":"tools/call","params":{"name":"get_weather","arguments":{"city":"Moscow"}},"id":3} | mcp-server.exe

### Конфигурация для Claude Desktop

#### Для Windows:
    {
        "mcpServers": {
            "go-mcp-server": {
                "command": "C:\\path\\to\\your\\mcp-server.exe"
            }
        }
    }
#### Для Linux/Mac:
    {
    "mcpServers": {
        "go-mcp-server": {
                "command": "/path/to/your/mcp-server"
            }
        }
    }

