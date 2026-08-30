package sentiment

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"strings"
)

type Result struct {
	NewsID    string  `json:"newsId"`
	Sentiment string  `json:"sentiment"`
	Score     float64 `json:"score"`
	Model     struct {
		Name    string `json:"name"`
		Version string `json:"version"`
	} `json:"model"`
	CreatedAt int64 `json:"createdAt"`
}
type Client struct {
	BaseURL string
	HTTP    *http.Client
}

func NewClient(baseURL string, httpClient *http.Client) *Client {
	if httpClient == nil {
		httpClient = http.DefaultClient
	}
	return &Client{BaseURL: strings.TrimRight(baseURL, "/"), HTTP: httpClient}
}
func (c *Client) Analyze(ctx context.Context, newsID, text string) (Result, error) {
	if newsID == "" || strings.TrimSpace(text) == "" {
		return Result{}, errors.New("news id and text are required")
	}
	body, _ := json.Marshal(map[string]string{"newsId": newsID, "text": text})
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, c.BaseURL+"/analyze", bytes.NewReader(body))
	if err != nil {
		return Result{}, err
	}
	req.Header.Set("Content-Type", "application/json")
	resp, err := c.HTTP.Do(req)
	if err != nil {
		return Result{}, fmt.Errorf("sentiment service: %w", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return Result{}, fmt.Errorf("sentiment service status %d", resp.StatusCode)
	}
	var result Result
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return Result{}, err
	}
	if result.Model.Name == "" || result.Model.Version == "" || result.Score < 0 || result.Score > 1 {
		return Result{}, errors.New("invalid sentiment response")
	}
	return result, nil
}
