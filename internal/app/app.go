package app

import (
	"atc_freq/internal/sim"
	"context"
)

type App struct {
	ctx        context.Context
	simService *sim.Service
}

func NewApp(connection sim.Connection) *App {
	client := sim.NewClient(connection)

	return &App{
		simService: sim.NewService(client),
	}
}

// AddContext is called when the Wails app starts. The context is saved
// so we can call the runtime methods
func (a *App) AddContext(ctx context.Context) {
	a.ctx = ctx
}

// GetFrequencies returns airport frequencies for the given ICAO code
func (a *App) GetFrequencies(icao string) ([]sim.AirportFrequency, error) {
	return a.simService.GetFrequency(icao)
}

// GetWeather returns weather information for the given waypoints
func (a *App) GetWeather(waypoints []string) (map[string]*sim.Weather, error) {
	return a.simService.GetWeather(waypoints)
}

// GetClouds returns weather information for the given waypoints
// Implementation of SimConnect_WeatherRequestCloudState in MSFS202 SDK API is broken and always returns 0
func (a *App) GetClouds(waypoints []string) (map[string][]sim.CloudDensity, error) {
	return a.simService.GetCloudDensityAtCoords(waypoints)
}
