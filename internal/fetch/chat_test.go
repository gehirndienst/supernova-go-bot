package fetch

import (
	"net/http"
	"os"
	"testing"

	"github.com/joho/godotenv"
	"github.com/rs/zerolog"
	"github.com/stretchr/testify/assert"
)

func TestChatFetcher_Fetch(t *testing.T) {
	err := godotenv.Load("../../.env")
	if err != nil {
		t.Skip("no .env file found, skipping TestChatFetcher_Fetch")
	}

	apiKey := os.Getenv("OPENAI_API_KEY")
	if apiKey == "" {
		t.Skip("no OpenAI API key found in env, skipping TestChatFetcher_Fetch")
	}

	cf := ChatFetcher{
		BaseFetcher: BaseFetcher{
			APIKey: apiKey,
			logger: &zerolog.Logger{},
		},
		client: &http.Client{},
	}

	tests := []struct {
		name    string
		params  map[string]interface{}
		wantErr bool
	}{
		{
			name: "Valid message",
			params: map[string]interface{}{
				"message": "Hello, World!",
			},
			wantErr: false,
		},
		{
			name:    "Missing message",
			params:  map[string]interface{}{},
			wantErr: true,
		},
		{
			name: "Empty message",
			params: map[string]interface{}{
				"message": "",
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := cf.Fetch(tt.params)
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.NotEmpty(t, got)
			}
		})
	}
}
