package deepseek

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"sync"
)

type Client struct {
	apiKey string
	client *http.Client
	// Хранилище истории чатов для каждого пользователя
	chatHistories map[int64][]Message
	mu            sync.RWMutex
}

type Message struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

type Request struct {
	Model       string    `json:"model"`
	Messages    []Message `json:"messages"`
	Temperature float64   `json:"temperature,omitempty"`
	MaxTokens   int       `json:"max_tokens,omitempty"`
}

type Response struct {
	Choices []struct {
		Message struct {
			Content string `json:"content"`
		} `json:"message"`
	} `json:"choices"`
}

func NewClient() (*Client, error) {
	apiKey := os.Getenv("DEEPSEEK_API_KEY")
	if apiKey == "" {
		return nil, nil
	}

	return &Client{
		apiKey:        apiKey,
		client:        &http.Client{},
		chatHistories: make(map[int64][]Message),
	}, nil
}

// ResetChatHistory очищает историю чата для указанного пользователя
func (c *Client) ResetChatHistory(userID int64) {
	c.mu.Lock()
	defer c.mu.Unlock()
	delete(c.chatHistories, userID)
}

// GetResponse получает ответ от DeepSeek с учетом истории чата
func (c *Client) GetResponse(ctx context.Context, userID int64, prompt string) (string, error) {
	if c.apiKey == "" {
		return "DeepSeek API не настроен. Пожалуйста, установите переменную окружения DEEPSEEK_API_KEY", nil
	}

	c.mu.Lock()
	history, exists := c.chatHistories[userID]
	if !exists {
		history = make([]Message, 0)
	}

	// Добавляем новое сообщение пользователя
	history = append(history, Message{
		Role:    "user",
		Content: prompt,
	})
	c.chatHistories[userID] = history
	c.mu.Unlock()

	reqBody := Request{
		Model:       "deepseek-chat",
		Messages:    history,
		Temperature: 0.7,
		MaxTokens:   1000,
	}

	jsonData, err := json.Marshal(reqBody)
	if err != nil {
		return "", fmt.Errorf("error marshaling request: %v", err)
	}

	log.Printf("Sending request to DeepSeek API with body: %s", string(jsonData))

	req, err := http.NewRequestWithContext(ctx, "POST", "https://api.deepseek.com/beta/chat/completions", bytes.NewBuffer(jsonData))
	if err != nil {
		return "", fmt.Errorf("error creating request: %v", err)
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+c.apiKey)

	resp, err := c.client.Do(req)
	if err != nil {
		log.Printf("Request details: URL=%s, Headers=%v", req.URL, req.Header)
		return "", fmt.Errorf("error making request: %v", err)
	}
	defer resp.Body.Close()

	body, _ := io.ReadAll(resp.Body)
	if resp.StatusCode != http.StatusOK {
		log.Printf("API Error Response: Status=%d, Body=%s", resp.StatusCode, string(body))
		return "", fmt.Errorf("API error: %s, status: %d, body: %s", resp.Status, resp.StatusCode, string(body))
	}

	var response Response
	if err := json.Unmarshal(body, &response); err != nil {
		return "", fmt.Errorf("error decoding response: %v, body: %s", err, string(body))
	}

	if len(response.Choices) == 0 {
		return "", fmt.Errorf("no choices in response: %s", string(body))
	}

	// Сохраняем ответ ассистента в историю
	c.mu.Lock()
	c.chatHistories[userID] = append(c.chatHistories[userID], Message{
		Role:    "assistant",
		Content: response.Choices[0].Message.Content,
	})
	c.mu.Unlock()

	return response.Choices[0].Message.Content, nil
}
