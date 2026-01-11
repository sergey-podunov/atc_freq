package app

import (
	"atc_freq/internal/sim"
	"context"
	"fmt"
)

// App struct
type App struct {
	ctx        context.Context
	simService *sim.Service
}

// NewApp creates a new App application struct
func NewApp() *App {
	client, err := sim.NewClient()
	if err != nil {
		fmt.Printf("Warning: failed to initialize Client: %v\n", err)
		return &App{
			simService: nil,
		}
	}
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
	if a.simService == nil {
		client, err := sim.NewClient()
		if err != nil {
			return nil, fmt.Errorf("sim service not initialized: %w", err)
		}
		a.simService = sim.NewService(client)
	}

	return a.simService.GetFrequency(icao)
}

// GetWeather returns weather information for the given waypoints
func (a *App) GetWeather(waypoints []string) (map[string]string, error) {
	if a.simService == nil {
		client, err := sim.NewClient()
		if err != nil {
			return nil, fmt.Errorf("sim service not initialized: %w", err)
		}
		a.simService = sim.NewService(client)
	}

	return a.simService.GetWeather(waypoints)
}
