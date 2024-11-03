package fetch

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"time"

	"github.com/rs/zerolog"
)

type ChatFetcher struct {
	BaseFetcher
	client *http.Client
}

type Message struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

type ChatGPTRequest struct {
	Model    string    `json:"model"`
	Messages []Message `json:"messages"`
}

type ChatGPTResponse struct {
	Choices []struct {
		Message struct {
			Content string `json:"content"`
		} `json:"message"`
	} `json:"choices"`
}

func (cf *ChatFetcher) Set(APIKey string, logger *zerolog.Logger) error {
	if err := cf.BaseFetcher.Set(APIKey, logger); err != nil {
		return err
	}
	cf.client = &http.Client{
		Timeout: 10 * time.Second,
	}
	return nil
}

func (cf *ChatFetcher) Fetch(qParams map[string]interface{}) (string, error) {
	if !cf.isSet() {
		cf.logger.Error().Msg("chat fetcher is not set")
		return "", errors.New("chat fetcher is not set")
	}

	userMessage, ok := qParams["message"].(string)
	if !ok || userMessage == "" {
		cf.logger.Error().Msg("message is required")
		return "", errors.New("message is required")
	}

	reqBody := ChatGPTRequest{
		Model: "gpt-3.5-turbo",
		Messages: []Message{
			{Role: "user", Content: userMessage},
		},
	}

	jsonBody, err := json.Marshal(reqBody)
	if err != nil {
		cf.logger.Error().Err(err).Msg("error marshalling request body")
		return "", err
	}

	req, err := http.NewRequest("POST", "https://api.openai.com/v1/chat/completions", bytes.NewBuffer(jsonBody))
	if err != nil {
		cf.logger.Error().Err(err).Msg("error creating request")
		return "", err
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", cf.APIKey))

	resp, err := cf.client.Do(req)
	if err != nil {
		cf.logger.Error().Err(err).Msg("error sending request")
		return "", err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		cf.logger.Error().Err(err).Msg("error reading response body")
		return "", err
	}

	if resp.StatusCode != http.StatusOK {
		cf.logger.Error().
			Int("status_code", resp.StatusCode).
			Msgf("chat API request failed: %s", string(body))
		return "", fmt.Errorf("chat API request failed with status %d", resp.StatusCode)
	}

	var chatGPTResp ChatGPTResponse
	if err := json.Unmarshal(body, &chatGPTResp); err != nil {
		cf.logger.Error().Err(err).Msg("error unmarshalling response body")
		return "", err
	}

	if len(chatGPTResp.Choices) == 0 || chatGPTResp.Choices[0].Message.Content == "" {
		cf.logger.Error().Msg("no valid response from ChatGPT")
		return "", errors.New("no valid response from ChatGPT")
	}

	return chatGPTResp.Choices[0].Message.Content, nil
}
