package nws

import "time"

// PointsResponse represents the response from /points for a given lat/lon.
type PointsResponse struct {
	Properties struct {
		Forecast         string `json:"forecast"`
		ForecastHourly   string `json:"forecastHourly"`
		ForecastGridData string `json:"forecastGridData"`
		GridID           string `json:"gridID"`
		GridX            int    `json:"gridX"`
		GridY            int    `json:"gridY"`
	} `json:"properties"`
}

// Forecast is the NWS forecast document with periods and metadata.
type Forecast struct {
	Properties struct {
		Updated time.Time `json:"updated"`
		Units   string    `json:"units"`
		Periods []Period  `json:"periods"`
	} `json:"properties"`
}

// Period is an individual forecast period (e.g., GetTodaysForcast, Tonight).
type Period struct {
	Name             string    `json:"name"`
	StartTime        time.Time `json:"startTime"`
	EndTime          time.Time `json:"endTime"`
	IsDaytime        bool      `json:"isDaytime"`
	Temperature      int       `json:"temperature"`
	TemperatureUnit  string    `json:"temperatureUnit"`
	ShortForecast    string    `json:"shortForecast"`
	DetailedForecast string    `json:"detailedForecast"`
}
