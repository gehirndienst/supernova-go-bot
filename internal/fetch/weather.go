package fetch

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"
)

const (
	FreeTierMaxDaysForecast  = 5
	FreeTierMaxHoursForecast = 12
)

type WeatherFetcher struct {
	BaseFetcher
}

type LocationResponse struct {
	Key string `json:"Key"`
}

type ForecastResponse struct {
	DailyForecasts  []DailyForecastResponse
	HourlyForecasts []HourlyForecastResponse
}

type DailyForecastResponses struct {
	DailyForecastResponses []DailyForecastResponse `json:"DailyForecasts"`
}

type DailyForecastResponse struct {
	Date        string `json:"Date"`
	Temperature struct {
		Minimum struct {
			Value float32 `json:"Value"`
			Unit  string  `json:"Unit"`
		} `json:"Minimum"`
		Maximum struct {
			Value float32 `json:"Value"`
			Unit  string  `json:"Unit"`
		} `json:"Maximum"`
	} `json:"Temperature"`
	Day struct {
		IconPhrase             string `json:"IconPhrase"`
		HasPrecipitation       bool   `json:"HasPrecipitation"`
		PrecipitationType      string `json:"PrecipitationType"`
		PrecipitationIntensity string `json:"PrecipitationIntensity"`
	} `json:"Day"`
	Night struct {
		IconPhrase             string `json:"IconPhrase"`
		HasPrecipitation       bool   `json:"HasPrecipitation"`
		PrecipitationType      string `json:"PrecipitationType"`
		PrecipitationIntensity string `json:"PrecipitationIntensity"`
	} `json:"Night"`
}

type HourlyForecastResponse struct {
	DateTime         string `json:"DateTime"`
	WeatherIcon      int    `json:"WeatherIcon"`
	IconPhrase       string `json:"IconPhrase"`
	HasPrecipitation bool   `json:"HasPrecipitation"`
	IsDaylight       bool   `json:"IsDaylight"`
	Temperature      struct {
		Value float32 `json:"Value"`
		Unit  string  `json:"Unit"`
	} `json:"Temperature"`
	PrecipitationProbability float32 `json:"PrecipitationProbability"`
	MobileLink               string  `json:"MobileLink"`
	Link                     string  `json:"Link"`
}

func utilTryRFCDateToHumanReadableDate(date string) string {
	t, err := time.Parse(time.RFC3339, date)
	if err != nil {
		return date
	}
	return t.Format("2006-01-02 15:04:05")
}

func utilFahrenheitToCelsius(f float32) float32 {
	return (f - 32) * 5 / 9
}

func (fr ForecastResponse) String() string {
	r := ""
	if len(fr.DailyForecasts) > 0 {
		for _, day := range fr.DailyForecasts {
			r += fmt.Sprintf("Date: %s\n", utilTryRFCDateToHumanReadableDate(day.Date))
			if day.Temperature.Minimum.Unit == "F" {
				day.Temperature.Minimum.Value = utilFahrenheitToCelsius(day.Temperature.Minimum.Value)
				day.Temperature.Minimum.Unit = "C"
			}
			r += fmt.Sprintf("Min Temp: %.2f %s\n", day.Temperature.Minimum.Value, day.Temperature.Minimum.Unit)
			if day.Temperature.Maximum.Unit == "F" {
				day.Temperature.Maximum.Value = utilFahrenheitToCelsius(day.Temperature.Maximum.Value)
				day.Temperature.Maximum.Unit = "C"
			}
			r += fmt.Sprintf("Max Temp: %.2f %s\n", day.Temperature.Maximum.Value, day.Temperature.Maximum.Unit)
			r += fmt.Sprintf("Day: \n\tWeather: %s \n\tPrecipitation: %t\n", day.Day.IconPhrase, day.Day.HasPrecipitation)
			r += fmt.Sprintf("Night: \n\tWeather: %s \n\tPrecipitation: %t\n", day.Night.IconPhrase, day.Night.HasPrecipitation)
			r += "\n"
		}
	} else {
		for _, hour := range fr.HourlyForecasts {
			r += fmt.Sprintf("Date: %s\n", utilTryRFCDateToHumanReadableDate(hour.DateTime))
			if hour.Temperature.Unit == "F" {
				hour.Temperature.Value = utilFahrenheitToCelsius(hour.Temperature.Value)
				hour.Temperature.Unit = "C"
			}
			r += fmt.Sprintf("Temp: %.2f %s\n", hour.Temperature.Value, hour.Temperature.Unit)
			r += fmt.Sprintf("Daylight: %t\n", hour.IsDaylight)
			r += fmt.Sprintf("Precipitation: %t\n", hour.HasPrecipitation)
			r += fmt.Sprintf("Precipitation Probability: %f\n", hour.PrecipitationProbability)
			r += "\n"
		}
	}
	return r
}

// /////////////////////////////////////////////////////////////////////////////
// Struct
// /////////////////////////////////////////////////////////////////////////////

func (wf WeatherFetcher) buildCityURL(city string) string {
	return fmt.Sprintf("http://dataservice.accuweather.com/locations/v1/search?&q=%s&apikey=%s", strings.ToLower(city), wf.APIKey)
}

func (wf WeatherFetcher) getLocationKey(city string) (string, error) {
	url := wf.buildCityURL(city)

	resp, err := http.Get(url)
	if err != nil {
		wf.logger.Error().Err(err).Msg("error getting weather fetcher location key")
		return "", err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		wf.logger.Error().Err(err).Msg("error reading weather fetcher location key response")
		return "", err
	}

	var locations []LocationResponse
	if err := json.Unmarshal(body, &locations); err != nil {
		wf.logger.Error().Err(err).Msg("error unmarshalling weather fetcher location key response")
		return "", err
	}

	if len(locations) == 0 {
		wf.logger.Error().Msg("weather fetcher: no locations found")
		return "", errors.New("no locations found")
	}

	return locations[0].Key, nil
}

func (wf WeatherFetcher) buildURL(qParams map[string]interface{}) (string, error) {
	baseURL := "http://dataservice.accuweather.com/forecasts/v1/"

	city, ok := qParams["city"].(string)
	if city == "" || !ok {
		wf.logger.Error().Msg("weather fetcher: city is required")
		return "", errors.New("city is required")
	}

	locationKey, err := wf.getLocationKey(qParams["city"].(string))
	if err != nil {
		return "", err
	}

	rangeSegment := ""
	days, ok := qParams["days"].(int)
	if !ok {
		hours, ok := qParams["hours"].(int)
		if !ok {
			wf.logger.Error().Msg("weather fetcher: days or hours required")
			return "", errors.New("days or hours required")
		}

		// forecast API only supports 1 or 12 hours
		if hours <= 1 {
			hours = 1
		} else {
			hours = FreeTierMaxHoursForecast
		}
		rangeSegment = fmt.Sprintf("hourly/%dhour/", hours)
	} else {
		if days <= 1 {
			days = 1
		} else {
			days = FreeTierMaxDaysForecast
		}
		rangeSegment = fmt.Sprintf("daily/%dday/", days)
	}

	return fmt.Sprintf("%s%s%s?apikey=%s", baseURL, rangeSegment, locationKey, wf.APIKey), nil
}

// /////////////////////////////////////////////////////////////////////////////
// Interface : Fetchable
// /////////////////////////////////////////////////////////////////////////////

func (wf WeatherFetcher) Fetch(qParams map[string]interface{}) (string, error) {
	if !wf.isSet() {
		return "", errors.New("weather fetcher is not set")
	}

	url, err := wf.buildURL(qParams)
	if err != nil {
		return "", err
	}

	resp, err := http.Get(url)
	if err != nil {
		wf.logger.Error().Err(err).Msg("error getting weather fetcher forecast")
		return "", err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		wf.logger.Error().Err(err).Msg("error reading weather fetcher forecast response")
		return "", err
	}

	var forecast ForecastResponse
	if strings.Contains(url, "daily") {
		var dailyForecastResponses DailyForecastResponses
		if err := json.Unmarshal(body, &dailyForecastResponses); err != nil {
			wf.logger.Error().Err(err).Msg("error unmarshalling weather fetcher daily forecast response")
			return "", err
		}
		forecast.DailyForecasts = dailyForecastResponses.DailyForecastResponses
		days, ok := qParams["days"].(int)
		if ok && days > 0 {
			forecast.DailyForecasts = forecast.DailyForecasts[:min(days, FreeTierMaxDaysForecast)]
		}
	} else {
		var hourlyForecastResponses []HourlyForecastResponse
		if err := json.Unmarshal(body, &hourlyForecastResponses); err != nil {
			wf.logger.Error().Err(err).Msg("error unmarshalling weather fetcher hourly forecast response")
			return "", err
		}
		forecast.HourlyForecasts = hourlyForecastResponses
		hours, ok := qParams["hours"].(int)
		if ok && hours > 0 {
			forecast.HourlyForecasts = forecast.HourlyForecasts[:min(hours, FreeTierMaxHoursForecast)]
		}
	}

	return forecast.String(), nil
}
