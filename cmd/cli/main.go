//go:build windows
package main

import (
	"atc_freq/internal/sim"
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/urfave/cli/v2"
)

func main() {
	app := &cli.App{
		Name:  "atc_freq",
		Usage: "Get airport frequencies and weather information from MSFS",
		Commands: []*cli.Command{
			{
				Name:      "freq",
				Usage:     "Get all frequencies for an airfield",
				ArgsUsage: "<ICAO>",
				Action:    freqCommand,
				Description: "Retrieves and displays all available frequencies for the specified airport.\n\n   Example:\n      atc_freq freq EDDB",
			},
			{
				Name:      "weather",
				Usage:     "Get weather at waypoints",
				ArgsUsage: "<waypoint1,waypoint2,...>",
				Action:    weatherCommand,
				Description: "Retrieves weather information for a comma-separated list of waypoints.\n\n   Example:\n      atc_freq weather EDDB,UUMI,KJFK",
			},
		},
	}

	if err := app.Run(os.Args); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
}

func freqCommand(ctx *cli.Context) error {
	if ctx.NArg() != 1 {
		return fmt.Errorf("requires exactly one ICAO code argument")
	}

	icao := ctx.Args().Get(0)

	client, err := sim.NewSimClient()
	if err != nil {
		return fmt.Errorf("failed to load SimConnection: %w", err)
	}

	freqs, err := client.GetAirportFrequencies(icao, 10*time.Second)
	if err != nil {
		return fmt.Errorf("failed to get frequencies: %w", err)
	}

	if len(freqs) == 0 {
		fmt.Printf("No frequencies found for %s\n", icao)
		return nil
	}

	fmt.Printf("Frequencies for %s:\n", strings.ToUpper(icao))
	for _, f := range freqs {
		fmt.Printf("  %-10s %8.3f MHz  %s\n", f.Type, f.MHz, f.Name)
	}

	return nil
}

func weatherCommand(ctx *cli.Context) error {
	if ctx.NArg() != 1 {
		return fmt.Errorf("requires exactly one argument (comma-separated waypoints)")
	}

	waypointsStr := ctx.Args().Get(0)
	waypoints := strings.Split(waypointsStr, ",")

	// Trim whitespace from each waypoint
	for i := range waypoints {
		waypoints[i] = strings.TrimSpace(waypoints[i])
	}

	if len(waypoints) == 0 {
		return fmt.Errorf("no waypoints provided")
	}

	fmt.Println("Weather command - to be implemented")
	fmt.Printf("Requested waypoints: %v\n", waypoints)

	// TODO: Implement weather retrieval logic
	// For now, just show what waypoints were requested
	for _, wp := range waypoints {
		if wp != "" {
			fmt.Printf("  - %s: (weather data not yet implemented)\n", strings.ToUpper(wp))
		}
	}

	return nil
}
