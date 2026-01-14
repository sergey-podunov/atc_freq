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

// Startup is called when the app starts. The context is saved
// so we can call the runtime methods
func (a *App) Startup(ctx context.Context) {
	a.ctx = ctx
}

// GetFrequencies returns airport frequencies for the given ICAO code
func (a *App) GetFrequencies(icao string) ([]sim.AirportFrequency, error) {
	return a.simService.GetFrequency(icao)
}

// GetWeather returns weather information for the given waypoints
func (a *App) GetWeather(waypoints []string) (map[string]string, error) {
	return a.simService.GetWeather(waypoints)
}
