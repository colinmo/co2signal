package main

import (
	"encoding/json"
	"testing"
	"time"
)

// TestGitRunDiff talks to Git and gets a response type.
func TestParseForecastedCarbonIntensity(t *testing.T) {

	dataset := ForecastedCarbonIntensity{}
	b := []byte(`{
		"zone": "DK-DK2",
		"forecast": [
		  {
			"carbonIntensity": 326,
			"datetime": "2018-11-26T17:00:00.000Z"
		  },
		  {
			"carbonIntensity": 297,
			"datetime": "2018-11-26T18:00:00.000Z"
		  },
		  {
			"carbonIntensity": 194,
			"datetime": "2018-11-28T17:00:00.000Z"
		  }
		],
		"updatedAt": "2018-11-26T17:25:24.685Z"
	  }`)

	json.Unmarshal(b, &dataset)
	response := ParseForecastedCarbonIntensity(dataset)
	if response.Zone != "DK-DK2" {
		t.Fatalf("Failed to get Zone %s\n", response.Zone)
	}
	if response.UpdatedAt != "2018-11-26T17:25:24.685Z" {
		t.Fatalf("Failed to get the Updated At %s\n", response.UpdatedAt)
	}
	expected := time.Date(2018, 11, 26, 17, 25, 24, 685000000, time.UTC)
	if response.ParsedUpdatedAt != expected {
		t.Fatalf("Failed to get the right time %v|%v\n", response.ParsedUpdatedAt, expected)
	}
	if response.Forecast[0].CarbonIntensity != 326 {
		t.Fatalf("Failed to get first carbon intensity %v\n", response.Forecast[0].CarbonIntensity)
	}
	if response.Forecast[2].DateTime != "2018-11-28T17:00:00.000Z" {
		t.Fatalf("Failed to get last Datetime %v\n", response.Forecast[2].DateTime)
	}
	expected = time.Date(2018, 11, 28, 17, 0, 0, 0, time.UTC)
	if response.Forecast[2].ParsedDateTime != expected {
		t.Fatalf("Failed to parse last Datetime %v|%v\n", response.Forecast[2].ParsedDateTime, expected)
	}
}
