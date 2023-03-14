package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"strings"
	"unicode/utf8"

	"github.com/spf13/viper"
)

type WeatherData struct {
	Name    string        `json:"name"`
	Main    Main          `json:"main"`
	Weather []WeatherInfo `json:"weather"`
}

type Main struct {
	Temp      float64 `json:"temp"`
	FeelsLike float64 `json:"feels_like"`
	Pressure  float64 `json:"pressure"`
	Humidity  float64 `json:"humidity"`
}

type WeatherInfo struct {
	Main        string `json:"main"`
	Description string `json:"description"`
}

func main() {
	// Initialize logging
	logFile, err := os.OpenFile("weather.log", os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0644)
	if err != nil {
		log.Fatalln("Error opening log file:", err)
	}
	defer logFile.Close()
	logger := log.New(logFile, "", log.Ldate|log.Ltime)

	// Check if user has provided command-line argument for location
	if len(os.Args) < 2 {
		logger.Println("No location provided")
		fmt.Println("Please provide a location as a command-line argument.")
		return
	}

	// Get location from command-line argument and validate input
	location := strings.TrimSpace(os.Args[1])
	if location == "" {
		logger.Println("Invalid input: empty location")
		fmt.Println("Please provide a valid location.")
		return
	}
	if !utf8.ValidString(location) {
		logger.Println("Invalid input: non-UTF8 character in location")
		fmt.Println("Please provide a valid location.")
		return
	}
	if strings.ContainsAny(location, "!@#$%^&*()_+={}[]|\\;:'\"<>,.?/~`") {
		logger.Println("Invalid input: invalid character in location")
		fmt.Println("Please provide a valid location.")
		return
	}

	// Read API key from config file
	viper.SetConfigFile("config.yaml")
	err = viper.ReadInConfig()
	if err != nil {
		logger.Fatalf("Error reading config file: %v", err)
	}
	apiKey := viper.GetString("api_key")
	if apiKey == "" {
		logger.Fatalln("API key is missing in the config file")
	}

	// Call OpenWeatherMap API to get current conditions for the specified location
	weatherData, err := getWeatherData(location, apiKey)
	if err != nil {
		logger.Println("Error getting weather data:", err)
		fmt.Println("Unable to retrieve weather data. Please check your location and try again.")
		return
	}

	// Print the weather data to the console
	fmt.Println("Location: ", weatherData.Name)
	fmt.Println("Temperature: ", weatherData.Main.Temp, "°C")
	fmt.Println("Feels like: ", weatherData.Main.FeelsLike, "°C")
	fmt.Println("Pressure: ", weatherData.Main.Pressure, "hPa")
	fmt.Println("Humidity: ", weatherData.Main.Humidity, "%")
	fmt.Println("Weather: ", weatherData.Weather[0].Main)
	fmt.Println("Description: ", weatherData.Weather[0].Description)
}

func getWeatherData(location string, apiKey string) (*WeatherData, error) {
	location = strings.ReplaceAll(location, " ", "%20")

	client := &http.Client{}

	req, err := http.NewRequest("GET", fmt.Sprintf("https://api.openweathermap.org/data/2.5/weather?q=%s&appid=%s&units=metric", location, apiKey), nil)
	if err != nil {
		return nil, fmt.Errorf("error creating HTTP request: %v", err)
	}

	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("error calling weather API: %v", err)
	}
	defer resp.Body.Close()

	// Read response body into a byte slice
	body, err := ioutil.ReadAll(resp.Body)

	// Debug with this line if expected data is not found; otherwise comment it out
	// fmt.Println(string(body))
	// Found result for "Port Angeles" but not "Los Angeles"

	if err != nil {
		return nil, fmt.Errorf("error reading response body: %v", err)
	}

	// Unmarshal the response JSON into a WeatherData struct
	var weatherData WeatherData
	err = json.Unmarshal(body, &weatherData)
	if err != nil {
		return nil, fmt.Errorf("error unmarshalling JSON: %v", err)
	}

	// Check if the weather data contains any information
	if len(weatherData.Weather) == 0 {
		return nil, errors.New(fmt.Sprintf("no weather data available for the location '%s'", location))
	}

	return &weatherData, nil
}
