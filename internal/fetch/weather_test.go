package fetch

import (
	"os"
	"testing"

	"github.com/joho/godotenv"
	"github.com/rs/zerolog"
	"github.com/stretchr/testify/assert"
)

func TestWeatherFetcher(t *testing.T) {
	err := godotenv.Load("../../.env")
	if err != nil {
		t.Skip("No .env file found, skipping TestWeatherFetcher")
	}

	apiKey := os.Getenv("ACCU_WEATHER_API_KEY")
	if apiKey == "" {
		t.Skip("No AccuWeather API key found in env, skipping TestWeatherFetcher")
	}

	wf := WeatherFetcher{
		BaseFetcher: BaseFetcher{
			APIKey: apiKey,
			logger: &zerolog.Logger{},
		},
	}

	params := map[string]interface{}{
		"city": "London",
		"days": 1,
	}

	r, err := wf.Fetch(params)

	assert.NoError(t, err)
	assert.NotEmpty(t, r)
	assert.Contains(t, r, "Temp")
	assert.Contains(t, r, "Day")
	assert.Contains(t, r, "Night")
}
