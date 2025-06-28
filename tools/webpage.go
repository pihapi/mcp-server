package tools

import (
	"encoding/json"
	"fmt"
	"golang.org/x/net/html"
	"io"
	"mcp-server/server"
	"net/http"
	"strings"
	"time"
)

type WebPageTool struct{}

type WebPageArgs struct {
	URL string `json:"url"`
}

func NewWebPageTool() *WebPageTool {
	return &WebPageTool{}
}

func (t *WebPageTool) GetDefinition() server.Tool {
	return server.Tool{
		Name:        "fetch_webpage",
		Description: "Fetch and extract text content from a webpage",
		InputSchema: server.InputSchema{
			Type: "object",
			Properties: map[string]server.Property{
				"url": {
					Type:        "string",
					Description: "URL of the webpage to fetch",
				},
			},
			Required: []string{"url"},
		},
	}
}

func (t *WebPageTool) Execute(arguments json.RawMessage) (*server.CallToolResult, error) {
	var args WebPageArgs
	if err := json.Unmarshal(arguments, &args); err != nil {
		return nil, fmt.Errorf("invalid arguments: %v", err)
	}

	// Валидация URL
	if !strings.HasPrefix(args.URL, "http://") && !strings.HasPrefix(args.URL, "https://") {
		return nil, fmt.Errorf("URL must start with http:// or https://")
	}

	// Загрузка страницы
	content, err := t.fetchPage(args.URL)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch page: %v", err)
	}

	return &server.CallToolResult{
		Content: []server.Content{
			{
				Type: "text",
				Text: content,
			},
		},
	}, nil
}

func (t *WebPageTool) fetchPage(url string) (string, error) {
	// Создаем HTTP клиент с таймаутом
	client := &http.Client{
		Timeout: 30 * time.Second,
	}

	// Создаем запрос
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return "", err
	}

	// Устанавливаем User-Agent
	req.Header.Set("User-Agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36")

	// Выполняем запрос
	resp, err := client.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	// Проверяем статус
	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("HTTP %d: %s", resp.StatusCode, resp.Status)
	}

	// Читаем тело ответа
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}

	// Определяем тип контента
	contentType := resp.Header.Get("Content-Type")

	// Если это HTML, извлекаем текст
	if strings.Contains(contentType, "text/html") {
		text := t.extractTextFromHTML(string(body))
		return fmt.Sprintf("URL: %s\nTitle: %s\n\nContent:\n%s",
			url,
			t.extractTitle(string(body)),
			text), nil
	}

	// Для других типов возвращаем как есть (если это текст)
	if strings.Contains(contentType, "text/") {
		return fmt.Sprintf("URL: %s\nContent-Type: %s\n\nContent:\n%s",
			url,
			contentType,
			string(body)), nil
	}

	return fmt.Sprintf("URL: %s\nContent-Type: %s\nSize: %d bytes\n\n[Binary content not displayed]",
		url,
		contentType,
		len(body)), nil
}

func (t *WebPageTool) extractTitle(htmlContent string) string {
	doc, err := html.Parse(strings.NewReader(htmlContent))
	if err != nil {
		return "No title"
	}

	var title string
	var findTitle func(*html.Node)
	findTitle = func(n *html.Node) {
		if n.Type == html.ElementNode && n.Data == "title" && n.FirstChild != nil {
			title = n.FirstChild.Data
			return
		}
		for c := n.FirstChild; c != nil; c = c.NextSibling {
			findTitle(c)
		}
	}
	findTitle(doc)

	if title == "" {
		return "No title"
	}
	return strings.TrimSpace(title)
}

func (t *WebPageTool) extractTextFromHTML(htmlContent string) string {
	doc, err := html.Parse(strings.NewReader(htmlContent))
	if err != nil {
		return htmlContent
	}

	var textBuilder strings.Builder

	// Теги, которые полностью пропускаем вместе с содержимым
	skipTagsCompletely := map[string]bool{
		"script":   true,
		"style":    true,
		"noscript": true,
		"nav":      true,
		"header":   true,
		"footer":   true,
		"aside":    true,
		"menu":     true,
		"menuitem": true,
		"button":   true,
		"form":     true,
		"input":    true,
		"select":   true,
		"textarea": true,
		"label":    true,
		"iframe":   true,
		"object":   true,
		"embed":    true,
		"svg":      true,
		"canvas":   true,
		"map":      true,
		"area":     true,
		"audio":    true,
		"video":    true,
		"track":    true,
		"source":   true,
		"picture":  true,
		"meta":     true,
		"link":     true,
		"base":     true,
		"br":       true,
		"hr":       true,
		"wbr":      true,
		"dialog":   true,
		"template": true,
		"slot":     true,
	}

	// Блочные элементы для форматирования
	blockElements := map[string]bool{
		"p":          true,
		"div":        true,
		"h1":         true,
		"h2":         true,
		"h3":         true,
		"h4":         true,
		"h5":         true,
		"h6":         true,
		"li":         true,
		"section":    true,
		"article":    true,
		"main":       true,
		"blockquote": true,
		"pre":        true,
		"address":    true,
		"figcaption": true,
		"dt":         true,
		"dd":         true,
		"tr":         true,
		"td":         true,
		"th":         true,
	}

	// Элементы списков для специального форматирования
	listItemElements := map[string]bool{
		"li": true,
	}

	var extractText func(*html.Node, bool)
	extractText = func(n *html.Node, skipContent bool) {
		// Если нужно пропустить содержимое
		if skipContent {
			return
		}

		// Проверяем, нужно ли пропустить этот элемент и все его содержимое
		if n.Type == html.ElementNode && skipTagsCompletely[n.Data] {
			return
		}

		// Обрабатываем текстовые узлы
		if n.Type == html.TextNode {
			text := strings.TrimSpace(n.Data)
			if text != "" {
				// Проверяем родительский элемент для списков
				parent := n.Parent
				if parent != nil && listItemElements[parent.Data] {
					textBuilder.WriteString("• ")
				}
				textBuilder.WriteString(text)
				textBuilder.WriteString(" ")
			}
		}

		// Добавляем переносы перед блочными элементами
		if n.Type == html.ElementNode && blockElements[n.Data] {
			currentText := textBuilder.String()
			if len(currentText) > 0 && !strings.HasSuffix(currentText, "\n") {
				textBuilder.WriteString("\n")
			}
		}

		// Рекурсивно обходим дочерние узлы
		for c := n.FirstChild; c != nil; c = c.NextSibling {
			extractText(c, false)
		}

		// Добавляем переносы после блочных элементов
		if n.Type == html.ElementNode && blockElements[n.Data] {
			currentText := textBuilder.String()
			if len(currentText) > 0 && !strings.HasSuffix(currentText, "\n") {
				textBuilder.WriteString("\n")
			}
		}
	}

	// Ищем основной контент (main, article) или обрабатываем весь документ
	var findMainContent func(*html.Node) *html.Node
	findMainContent = func(n *html.Node) *html.Node {
		if n.Type == html.ElementNode && (n.Data == "main" || n.Data == "article") {
			return n
		}
		for c := n.FirstChild; c != nil; c = c.NextSibling {
			if result := findMainContent(c); result != nil {
				return result
			}
		}
		return nil
	}

	// Пытаемся найти основной контент
	mainContent := findMainContent(doc)
	if mainContent != nil {
		extractText(mainContent, false)
	} else {
		// Если не нашли main/article, обрабатываем body
		var findBody func(*html.Node) *html.Node
		findBody = func(n *html.Node) *html.Node {
			if n.Type == html.ElementNode && n.Data == "body" {
				return n
			}
			for c := n.FirstChild; c != nil; c = c.NextSibling {
				if result := findBody(c); result != nil {
					return result
				}
			}
			return nil
		}

		if body := findBody(doc); body != nil {
			extractText(body, false)
		} else {
			extractText(doc, false)
		}
	}

	// Очищаем результат
	result := textBuilder.String()

	// Убираем множественные пробелы и переносы
	lines := strings.Split(result, "\n")
	var cleanLines []string

	for _, line := range lines {
		line = strings.TrimSpace(line)
		// Заменяем множественные пробелы на один
		line = strings.Join(strings.Fields(line), " ")
		if line != "" {
			cleanLines = append(cleanLines, line)
		}
	}

	// Убираем дублирующиеся пустые строки
	var finalLines []string
	for i, line := range cleanLines {
		if i == 0 || line != "" || (i > 0 && cleanLines[i-1] != "") {
			finalLines = append(finalLines, line)
		}
	}

	result = strings.Join(finalLines, "\n")

	// Ограничиваем размер результата
	const maxLength = 10000
	if len(result) > maxLength {
		result = result[:maxLength] + "\n\n[Content truncated...]"
	}

	return result
}
