package fetch

import (
	"os"
	"testing"

	"github.com/joho/godotenv"
	"github.com/rs/zerolog"
	"github.com/stretchr/testify/assert"
)

func TestWeatherFetcher_Fetch(t *testing.T) {
	err := godotenv.Load("../../.env")
	if err != nil {
		t.Skip("No .env file found, skipping TestWeatherFetcher_Fetch")
	}

	apiKey := os.Getenv("ACCU_WEATHER_API_KEY")
	if apiKey == "" {
		t.Skip("No AccuWeather API key found in env, skipping TestWeatherFetcher_Fetch")
	}

	wf := WeatherFetcher{
		BaseFetcher: BaseFetcher{
			APIKey: apiKey,
			logger: &zerolog.Logger{},
		},
		locationCache: make(map[string]string),
	}

	tests := []struct {
		name    string
		params  map[string]interface{}
		wantErr bool
	}{
		{
			name: "Valid city and days",
			params: map[string]interface{}{
				"city": "London",
				"days": 1,
			},
			wantErr: false,
		},
		{
			name: "Valid city and hours",
			params: map[string]interface{}{
				"city":  "London",
				"hours": 1,
			},
			wantErr: false,
		},
		{
			name: "Invalid city",
			params: map[string]interface{}{
				"city": "InvalidCity",
				"days": 1,
			},
			wantErr: true,
		},
		{
			name: "Missing city",
			params: map[string]interface{}{
				"days": 1,
			},
			wantErr: true,
		},
		{
			name: "Missing days and hours",
			params: map[string]interface{}{
				"city": "London",
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := wf.Fetch(tt.params)
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.NotEmpty(t, got)
			}
		})
	}
}
