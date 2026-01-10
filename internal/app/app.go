package app

import (
	"atc_freq/internal/sim"
	"context"
	"fmt"
	"time"
)

// App struct
type App struct {
	ctx       context.Context
	simClient *sim.SimClient
}

// NewApp creates a new App application struct
func NewApp() *App {
	client, err := sim.NewSimClient()
	if err != nil {
		fmt.Printf("Warning: failed to initialize SimClient: %v\n", err)
	}
	return &App{
		simClient: client,
	}
}

// Startup is called when the app starts. The context is saved
// so we can call the runtime methods
func (a *App) Startup(ctx context.Context) {
	a.ctx = ctx
}

// GetFrequencies returns airport frequencies for the given ICAO code
func (a *App) GetFrequencies(icao string) ([]sim.AirportFrequency, error) {
	if a.simClient == nil {
		var err error
		a.simClient, err = sim.NewSimClient()
		if err != nil {
			return nil, fmt.Errorf("sim client not initialized: %w", err)
		}
	}
	return a.simClient.GetAirportFrequencies(icao, 10*time.Second)
}
