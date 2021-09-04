package main

import (
	"co2signal/icon"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"time"

	"github.com/getlantern/systray"
	"github.com/skratchdot/open-golang/open"
)

func main() {
	systray.Run(onReady, onExit)
}

type Config struct {
	ApiToken  string
	Zone      string
	Threshold int
}

func displayCurrentStatus(threshold int, cache ForecastedCarbonIntensity) string { // "2018-11-26T17:00:00.000Z"
	now := time.Now().UTC().Format("2006-01-02T15")
	icon1 := icon.Coal
	title := "?"
	for _, i := range cache.Forecast {
		fmt.Printf("%s vs %s\n", i.DateTime[0:13], now)
		if i.DateTime[0:13] == now {
			title = "CI: " + fmt.Sprintf("%d", i.CarbonIntensity)
			if i.CarbonIntensity < threshold {
				icon1 = icon.Green
			}
			break
		}
	}

	systray.SetTemplateIcon(icon1, icon1)
	systray.SetTitle(title)
	systray.SetTooltip(title)
	return title
}

func updateCache(config Config, cacheMenu *systray.MenuItem) ForecastedCarbonIntensity {
	e := ElectricityMapAPI{ApiToken: config.ApiToken}
	cache := e.RequestForecastedCarbonIntensity(config.Zone)
	SaveCache("../cache.json", cache)

	cacheMenu.SetTitle("Cache updated: " + cache.UpdatedAt)
	cacheMenu.SetTooltip("Cache updated: cache.UpdatedAt" + cache.UpdatedAt)
	return cache
}

func onReady() {
	// Load config and cache
	config := LoadPreferences("../preferences.json")
	cache := LoadCache("../cache.json")

	// If cache has expired, refresh
	if time.Now().After(cache.ParsedUpdatedAt.AddDate(0, 0, 1)) {
		fmt.Printf("%v after %v", time.Now(), cache.ParsedUpdatedAt)

		e := ElectricityMapAPI{ApiToken: config.ApiToken}
		cache = e.RequestForecastedCarbonIntensity(config.Zone)
		SaveCache("../cache.json", cache)
	}

	// Set up menu items
	systray.SetTemplateIcon(icon.Data, icon.Data)
	title := "Loading"
	systray.SetTitle(title)
	systray.SetTooltip(title)
	mSite := systray.AddMenuItem("Source of data", "Source of data")
	mCache := systray.AddMenuItem("Cache updated: "+cache.UpdatedAt, "Cache updated: cache.UpdatedAt"+cache.UpdatedAt)
	systray.AddSeparator()
	mAbout := systray.AddMenuItem("About", "About this app")
	mQuit := systray.AddMenuItem("Quit", "Quit this")

	// Display Menu
	go func() {
		for {
			select {
			case <-mSite.ClickedCh:
				open.Run("https://electricitymap.org")
			case <-mCache.ClickedCh:
				cache = updateCache(config, mCache)
			case <-mAbout.ClickedCh:
				open.Run("https://vonexplaino.com/")
			case <-mQuit.ClickedCh:
				systray.Quit()
				return
			}
		}
	}()
	// Update icon
	go func() {
		for {
			title = displayCurrentStatus(config.Threshold, cache)
			if title == "?" {
				cache = updateCache(config, mCache)
				displayCurrentStatus(config.Threshold, cache)
			}

			now := time.Now()
			waiting := time.Until(time.Date(now.Year(), now.Month(), now.Day(), now.Hour()+1, 0, 0, 0, now.Location()))
			fmt.Printf("Waiting for %s\n", waiting)
			time.Sleep(waiting)
		}
	}()
}

func onExit() {
	// now := time.Now()
	// ioutil.WriteFile(fmt.Sprintf(`on_exit_%d.txt`, now.UnixNano()), []byte(now.String()), 0644)
}

// Electricity Map API object
type ElectricityMapAPI struct {
	ApiToken string
}

type forecast struct {
	CarbonIntensity int    `json:"carbonIntensity"`
	DateTime        string `json:"datetime"`
	ParsedDateTime  time.Time
}

type ForecastedCarbonIntensity struct {
	Zone            string     `json:"zone"`
	Forecast        []forecast `json:"forecast"`
	UpdatedAt       string     `json:"updatedAt"`
	ParsedUpdatedAt time.Time
}

func (e ElectricityMapAPI) RequestForecastedCarbonIntensity(zone string) ForecastedCarbonIntensity {
	return ParseForecastedCarbonIntensity(e.GetDatasetFromRemote(zone))
}

func (e ElectricityMapAPI) GetDatasetFromRemote(zone string) ForecastedCarbonIntensity {
	dataset := ForecastedCarbonIntensity{}
	client := &http.Client{
		Timeout: time.Second * 10,
	}
	req, err := http.NewRequest("GET", fmt.Sprintf("https://api.electricitymap.org/v3/carbon-intensity/forecast?zone=%s", zone), nil)
	if err != nil {
		return dataset
	}
	req.Header.Set("auth-token", e.ApiToken)
	response, err := client.Do(req)
	if err != nil {
		return dataset
	}
	defer response.Body.Close()
	b, err := io.ReadAll(response.Body)
	if err != nil {
		return dataset
	}

	json.Unmarshal(b, &dataset)
	return dataset
}

func ParseForecastedCarbonIntensity(forecast ForecastedCarbonIntensity) ForecastedCarbonIntensity {
	for x, y := range forecast.Forecast {
		forecast.Forecast[x].ParsedDateTime, _ = time.Parse("2006-01-02T15:04:05.999Z", y.DateTime)
	}
	forecast.ParsedUpdatedAt, _ = time.Parse("2006-01-02T15:04:05.999Z", forecast.UpdatedAt)
	return forecast
}

// End object

func LoadPreferences(path string) Config {
	data := LoadJsonFile(path)

	var obj Config
	if json.Unmarshal(data, &obj) != nil {
		return Config{}
	}
	return obj
}

func SavePreferences(path string, config Config) {
	data, _ := json.Marshal(config)
	ioutil.WriteFile(path, data, 0666)
}

func LoadCache(path string) ForecastedCarbonIntensity {
	data := LoadJsonFile(path)

	var obj ForecastedCarbonIntensity
	if json.Unmarshal(data, &obj) != nil {
		return ForecastedCarbonIntensity{}
	}
	return ParseForecastedCarbonIntensity(obj)
}

func SaveCache(path string, cache ForecastedCarbonIntensity) {
	data, _ := json.Marshal(cache)
	ioutil.WriteFile(path, data, 0666)
}

func LoadJsonFile(path string) []byte {
	data, err := ioutil.ReadFile(path)
	if err != nil {
		fmt.Print(err)
	}

	return data
}
